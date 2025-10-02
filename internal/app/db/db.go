// internal/app/db/db.go
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

    _, err = conn.Exec(`CREATE TABLE IF NOT EXISTS passwords (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        service TEXT NOT NULL,
        username TEXT NOT NULL,
        link TEXT NOT NULL,
        password TEXT NOT NULL,
        category TEXT NOT NULL,
        created_at TEXT NOT NULL
    )`)
    if err != nil {
        return nil, err
    }

    _, err = conn.Exec(`CREATE TABLE IF NOT EXISTS master_password (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT NOT NULL,
        hash TEXT NOT NULL
    )`)
    if err != nil {
        return nil, err
    }

    return NewSQLStorage(conn, crypto), nil
}
