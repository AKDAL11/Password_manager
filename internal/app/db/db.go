package db

import (
    "database/sql"

    _ "github.com/mattn/go-sqlite3"
    "password-manager/pkg/utils"
)

func InitDB(path string, crypto *utils.CryptoService) (Storage, error) {
    conn, err := sql.Open("sqlite3", path)
    if err != nil {
        return nil, err
    }

    // Таблица паролей: хранится base64(AES-GCM)
    if _, err = conn.Exec(`CREATE TABLE IF NOT EXISTS passwords (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        service TEXT NOT NULL,
        username TEXT NOT NULL,
        link TEXT NOT NULL,
        password TEXT NOT NULL,    -- base64(nonce||ciphertext)
        category TEXT NOT NULL,
        created_at TEXT NOT NULL
    )`); err != nil {
        return nil, err
    }

    // Одна запись мастер-пароля: id=1
    if _, err = conn.Exec(`CREATE TABLE IF NOT EXISTS master_password (
        id INTEGER PRIMARY KEY,
        email TEXT NOT NULL,
        salt TEXT NOT NULL,        -- base64
        verifier TEXT NOT NULL     -- base64 SHA256(key)
    )`); err != nil {
        return nil, err
    }

    return NewSQLStorage(conn, crypto), nil
}
