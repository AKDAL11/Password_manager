package db

import (
    "crypto/sha256"
    "database/sql"
    "encoding/base64"
    "errors"
    "fmt"
    "log"
    "time"

    "password-manager/internal/app/model"
    "password-manager/pkg/security"
    "password-manager/pkg/utils"
)

// Storage — интерфейс твоего хранилища (если он у тебя есть, оставь как есть)
// type Storage interface { ... }

type SQLStorage struct {
    DB     *sql.DB
    Crypto *utils.CryptoService
}

func NewSQLStorage(db *sql.DB, crypto *utils.CryptoService) Storage {
    return &SQLStorage{DB: db, Crypto: crypto}
}

// Реализация новых методов интерфейса
func (s *SQLStorage) HasMeta() bool {
    var count int
    _ = s.DB.QueryRow(`SELECT COUNT(*) FROM meta WHERE id = 1`).Scan(&count)
    return count == 1
}

func (s *SQLStorage) SetCrypto(c *utils.CryptoService) {
    s.Crypto = c
}

// ---------------- Generic guards ----------------

func (s *SQLStorage) requireCrypto() error {
    if s.Crypto == nil {
        return errors.New("locked: master password not verified or set")
    }
    return nil
}

// ---------------- Passwords ----------------

func (s *SQLStorage) CreatePassword(p model.Password) (int64, string, error) {
    if err := s.requireCrypto(); err != nil {
        return 0, "", err
    }

    createdAt := time.Now().UTC().Format(time.RFC3339)

    res, err := s.DB.Exec(
        "INSERT INTO passwords (service, username, link, password, category, created_at) VALUES (?, ?, ?, ?, ?, ?)",
        p.Service, p.Username, p.Link, p.Password, p.Category, createdAt,
    )
    if err != nil {
        return 0, "", err
    }

    newID, _ := res.LastInsertId()
    return newID, createdAt, nil
}

func (s *SQLStorage) UpdatePassword(id string, p model.Password) error {
    if err := s.requireCrypto(); err != nil {
        return err
    }
    encrypted, err := s.Crypto.Encrypt(p.Password)
    if err != nil {
        return err
    }
    _, err = s.DB.Exec(
        "UPDATE passwords SET service = ?, username = ?, link = ?, password = ?, category = ? WHERE id = ?",
        p.Service, p.Username, p.Link, encrypted, p.Category, id,
    )
    return err
}

func (s *SQLStorage) DeletePassword(id string) error {
    _, err := s.DB.Exec("DELETE FROM passwords WHERE id = ?", id)
    return err
}

func (s *SQLStorage) GetAllPasswords() ([]model.PasswordListItem, error) {
    rows, err := s.DB.Query("SELECT id, service, username, link, category, created_at, password FROM passwords")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var list []model.PasswordListItem
    for rows.Next() {
        var item model.PasswordListItem
        var encrypted string
        if err := rows.Scan(&item.ID, &item.Service, &item.Username, &item.Link, &item.Category, &item.CreatedAt, &encrypted); err != nil {
            return nil, err
        }

        // IMPORTANT: do not leak encrypted base64 into display Password field
        item.Password = "" // leave empty so UI doesn't show encrypted nonsense

        list = append(list, item)
    }
    return list, nil
}

func (s *SQLStorage) GetFilteredPasswords(service, username, category string) ([]model.PasswordListItem, error) {
    query := "SELECT id, service, username, link, category, created_at, password FROM passwords WHERE 1=1"
    var args []interface{}
    if service != "" {
        query += " AND service LIKE ?"
        args = append(args, "%"+service+"%")
    }
    if username != "" {
        query += " AND username LIKE ?"
        args = append(args, "%"+username+"%")
    }
    if category != "" {
        query += " AND category = ?"
        args = append(args, category)
    }

    rows, err := s.DB.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var list []model.PasswordListItem
    for rows.Next() {
        var item model.PasswordListItem
        var encrypted string
        if err := rows.Scan(
            &item.ID,
            &item.Service,
            &item.Username,
            &item.Link,
            &item.Category,
            &item.CreatedAt,
            &encrypted,
        ); err != nil {
            return nil, err
        }

        // Do not set encrypted value into item.Password
        item.Password = ""

        list = append(list, item)
    }
    return list, nil
}

// Только метаданные (без пароля), если нужно
func (s *SQLStorage) GetPasswordByID(id string) (model.PasswordListItem, error) {
    var p model.PasswordListItem
    err := s.DB.QueryRow("SELECT id, service, username, link, category, created_at FROM passwords WHERE id = ?", id).
        Scan(&p.ID, &p.Service, &p.Username, &p.Link, &p.Category, &p.CreatedAt)
    return p, err
}

// Для копирования: всегда берём актуальный шифртекст из БД
func (s *SQLStorage) GetEncryptedPasswordByID(id int) (string, error) {
    var enc string
    err := s.DB.QueryRow("SELECT password FROM passwords WHERE id = ?", id).Scan(&enc)
    if err != nil {
        return "", err
    }
    return enc, nil
}

// ---------------- Meta: соль и верификатор ----------------

func EnsureMeta(db *sql.DB) error {
    _, err := db.Exec(`CREATE TABLE IF NOT EXISTS meta (
        id INTEGER PRIMARY KEY CHECK (id = 1),
        salt TEXT NOT NULL,
        verifier TEXT NOT NULL
    )`)
    return err
}

func LoadOrInitMasterFromDB(db *sql.DB, masterPassword string) ([]byte, error) {
    if err := EnsureMeta(db); err != nil {
        return nil, err
    }

    var saltB64, verB64 string
    err := db.QueryRow("SELECT salt, verifier FROM meta WHERE id=1").Scan(&saltB64, &verB64)
    if err == sql.ErrNoRows {
        salt := security.GenerateSalt(16)
        key := security.DeriveKey([]byte(masterPassword), salt)
        ver := sha256.Sum256(key)

        _, err = db.Exec(
            "INSERT INTO meta (id, salt, verifier) VALUES (1, ?, ?)",
            base64.StdEncoding.EncodeToString(salt),
            base64.StdEncoding.EncodeToString(ver[:]),
        )
        if err != nil {
            return nil, err
        }
        log.Printf("DEBUG: set salt=%s", base64.StdEncoding.EncodeToString(salt))
        log.Printf("DEBUG: set key hash=%x", ver)
        return key, nil
    }
    if err != nil {
        return nil, err
    }

    salt, err := base64.StdEncoding.DecodeString(saltB64)
    if err != nil {
        return nil, err
    }
    expectedVer, err := base64.StdEncoding.DecodeString(verB64)
    if err != nil {
        return nil, err
    }

    key := security.DeriveKey([]byte(masterPassword), salt)
    actualVer := sha256.Sum256(key)
    if !security.BytesEqual(actualVer[:], expectedVer) {
        return nil, errors.New("invalid master password")
    }

    log.Printf("DEBUG: verify salt=%s", saltB64)
    log.Printf("DEBUG: verify key hash=%x", actualVer)
    return key, nil
}

func UpdateVerifier(db *sql.DB, key []byte) error {
    ver := sha256.Sum256(key)
    _, err := db.Exec("UPDATE meta SET verifier=? WHERE id=1", base64.StdEncoding.EncodeToString(ver[:]))
    return err
}

func ReencryptAll(db *sql.DB, oldKey, newKey []byte) error {
    rows, err := db.Query("SELECT id, password FROM passwords")
    if err != nil {
        return err
    }
    defer rows.Close()

    tx, err := db.Begin()
    if err != nil {
        return err
    }

    for rows.Next() {
        var id int
        var encB64 string
        if err := rows.Scan(&id, &encB64); err != nil {
            tx.Rollback()
            return err
        }

        enc, err := base64.StdEncoding.DecodeString(encB64)
        if err != nil {
            tx.Rollback()
            return err
        }

        pt, err := security.DecryptAESGCM(oldKey, enc)
        if err != nil {
            tx.Rollback()
            return fmt.Errorf("id=%d decrypt failed: %w", id, err)
        }

        newEnc, err := security.EncryptAESGCM(newKey, pt)
        if err != nil {
            tx.Rollback()
            return fmt.Errorf("id=%d encrypt failed: %w", id, err)
        }

        _, err = tx.Exec("UPDATE passwords SET password=? WHERE id=?", base64.StdEncoding.EncodeToString(newEnc), id)
        if err != nil {
            tx.Rollback()
            return err
        }
    }
    if err := rows.Err(); err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit()
}

func (s *SQLStorage) Close() error {
    return s.DB.Close()
}
