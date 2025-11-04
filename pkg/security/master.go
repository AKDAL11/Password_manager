package security

import (
    "crypto/sha256"
    "database/sql"
    "encoding/base64"
    "errors"
    "fmt"
)

// LoadOrInitMasterFromDB:
// - Если в таблице meta нет записи — создаёт соль, выводит ключ из введённого пароля,
//   сохраняет соль и верификатор, возвращает key.
// - Если запись есть — проверяет пароль, возвращает key при успехе.
func LoadOrInitMasterFromDB(db *sql.DB, masterPassword string) ([]byte, error) {
    // создаём таблицу meta при необходимости
    _, err := db.Exec(`CREATE TABLE IF NOT EXISTS meta (
        id INTEGER PRIMARY KEY CHECK (id = 1),
        salt TEXT NOT NULL,
        verifier TEXT NOT NULL
    )`)
    if err != nil {
        return nil, err
    }

    var saltB64, verB64 string
    err = db.QueryRow("SELECT salt, verifier FROM meta WHERE id=1").Scan(&saltB64, &verB64)
    if err == sql.ErrNoRows {
        // первый запуск
        salt := GenerateSalt(16)
        key := DeriveKey([]byte(masterPassword), salt)
        ver := sha256.Sum256(key)

        _, err = db.Exec(
            "INSERT INTO meta (id, salt, verifier) VALUES (1, ?, ?)",
            base64.StdEncoding.EncodeToString(salt),
            base64.StdEncoding.EncodeToString(ver[:]),
        )
        if err != nil {
            return nil, err
        }
        return key, nil
    }
    if err != nil {
        return nil, err
    }

    // повторный запуск: проверяем пароль
    salt, err := base64.StdEncoding.DecodeString(saltB64)
    if err != nil {
        return nil, err
    }
    expectedVer, err := base64.StdEncoding.DecodeString(verB64)
    if err != nil {
        return nil, err
    }

    key := DeriveKey([]byte(masterPassword), salt)
    actualVer := sha256.Sum256(key)
    if !BytesEqual(actualVer[:], expectedVer) {
        return nil, errors.New("invalid master password")
    }
    return key, nil
}

// ChangeMasterPassword обновляет верификатор под новый пароль.
// Перешифровку данных выполняй снаружи, используя oldKey и newKey.
func ChangeMasterPassword(db *sql.DB, oldPassword, newPassword string) ([]byte, []byte, error) {
    var saltB64, verB64 string
    if err := db.QueryRow("SELECT salt, verifier FROM meta WHERE id=1").Scan(&saltB64, &verB64); err != nil {
        return nil, nil, fmt.Errorf("master not initialized: %w", err)
    }

    salt, err := base64.StdEncoding.DecodeString(saltB64)
    if err != nil {
        return nil, nil, err
    }

    oldKey := DeriveKey([]byte(oldPassword), salt)
    oldVer := sha256.Sum256(oldKey)

    expectedVer, err := base64.StdEncoding.DecodeString(verB64)
    if err != nil {
        return nil, nil, err
    }
    if !BytesEqual(oldVer[:], expectedVer) {
        return nil, nil, errors.New("wrong old master password")
    }

    newKey := DeriveKey([]byte(newPassword), salt)
    newVer := sha256.Sum256(newKey)
    _, err = db.Exec("UPDATE meta SET verifier=? WHERE id=1", base64.StdEncoding.EncodeToString(newVer[:]))
    if err != nil {
        return nil, nil, err
    }
    return oldKey, newKey, nil
}
