// internal/app/db/sql_storage.go
package db

import (
    "database/sql"
    "errors"
    "time"

    "golang.org/x/crypto/bcrypt"
    "password-manager/internal/app/model"
    "password-manager/pkg/utils"
)

type SQLStorage struct {
    DB     *sql.DB
    Crypto *utils.CryptoService
}

func NewSQLStorage(db *sql.DB, crypto *utils.CryptoService) Storage {
    return &SQLStorage{DB: db, Crypto: crypto}
}

func (s *SQLStorage) CreatePassword(p model.Password) (int64, string, error) {
    createdAt := time.Now().UTC().Format(time.RFC3339)

    encrypted, err := s.Crypto.Encrypt(p.Password)
    if err != nil {
        return 0, "", err
    }

    // вычисляем MAX(id)
    var maxID int64
    err = s.DB.QueryRow("SELECT IFNULL(MAX(id), 0) FROM passwords").Scan(&maxID)
    if err != nil {
        return 0, "", err
    }
    newID := maxID + 1

    _, err = s.DB.Exec(
        "INSERT INTO passwords (id, service, username, link, password, category, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
        newID, p.Service, p.Username, p.Link, encrypted, p.Category, createdAt,
    )
    if err != nil {
        return 0, "", err
    }

    return newID, createdAt, nil
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
        decrypted, err := s.Crypto.Decrypt(encrypted)
        if err != nil {
            return nil, err
        }
        item.Password = decrypted
        list = append(list, item)
    }
    return list, nil
}

func (s *SQLStorage) GetPasswordByID(id string) (model.PasswordListItem, error) {
    var p model.PasswordListItem
    err := s.DB.QueryRow("SELECT id, service, username, link, category, created_at FROM passwords WHERE id = ?", id).
        Scan(&p.ID, &p.Service, &p.Username, &p.Link, &p.Category, &p.CreatedAt)
    return p, err
}

func (s *SQLStorage) GetEncryptedPassword(id string) (string, error) {
    var encrypted string
    if err := s.DB.QueryRow("SELECT password FROM passwords WHERE id = ?", id).Scan(&encrypted); err != nil {
        return "", err
    }
    return s.Crypto.Decrypt(encrypted)
}

func (s *SQLStorage) UpdatePassword(id string, p model.Password) error {
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
    tx, err := s.DB.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // удаляем запись
    _, err = tx.Exec("DELETE FROM passwords WHERE id = ?", id)
    if err != nil {
        return err
    }

    // сдвигаем все ID после удалённого
    _, err = tx.Exec("UPDATE passwords SET id = id - 1 WHERE id > ?", id)
    if err != nil {
        return err
    }

    return tx.Commit()
}

func (s *SQLStorage) Close() error {
    return s.DB.Close()
}

func (s *SQLStorage) GetFilteredPasswords(service, username, category string) ([]model.PasswordListItem, error) {
    query := "SELECT id, service, username, link, category, created_at, password FROM passwords WHERE 1=1"
    args := []interface{}{}

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
        if err := rows.Scan(&item.ID, &item.Service, &item.Username, &item.Link, &item.Category, &item.CreatedAt, &encrypted); err != nil {
            return nil, err
        }
        decrypted, err := s.Crypto.Decrypt(encrypted)
        if err != nil {
            return nil, err
        }
        item.Password = decrypted
        list = append(list, item)
    }
    return list, nil
}

func (s *SQLStorage) HasMasterPassword() bool {
    var count int
    _ = s.DB.QueryRow(`SELECT COUNT(*) FROM master_password`).Scan(&count)
    return count > 0
}

func (s *SQLStorage) SaveMasterPassword(email, plain string) error {
    hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    _, err = s.DB.Exec(`INSERT INTO master_password (email, hash) VALUES (?, ?)`, email, hash)
    return err
}

func (s *SQLStorage) VerifyMasterPassword(input string) error {
    var hash string
    if err := s.DB.QueryRow(`SELECT hash FROM master_password ORDER BY id DESC LIMIT 1`).Scan(&hash); err != nil {
        return errors.New("no master password set")
    }
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(input))
}

func (s *SQLStorage) GetRecoveryEmail() (string, error) {
    var email string
    err := s.DB.QueryRow(`SELECT email FROM master_password ORDER BY id DESC LIMIT 1`).Scan(&email)
    return email, err
}
