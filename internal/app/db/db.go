// db.go
package db

import (
    "database/sql"

    _ "github.com/mattn/go-sqlite3"
)

// InitDB открывает SQLite, создаёт таблицу и возвращает реализацию Storage
func InitDB(path string) (Storage, error) {
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
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`)
    if err != nil {
        return nil, err
    }

    return NewSQLStorage(conn), nil
}
