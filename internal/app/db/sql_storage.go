// sql_storage.go

package db

import (
    "database/sql"
    "password-manager/internal/app/model"
)

type SQLStorage struct {
    DB *sql.DB
}

func NewSQLStorage(db *sql.DB) Storage {
    return &SQLStorage{DB: db}
}

// Create a new password entry
func (s *SQLStorage) CreatePassword(p model.Password) (int64, string, error) {
    res, err := s.DB.Exec(
        "INSERT INTO passwords (service, username, link, password, category) VALUES (?, ?, ?, ?, ?)",
        p.Service, p.Username, p.Link, p.Password, p.Category,
    )
    if err != nil {
        return 0, "", err
    }
    lastID, _ := res.LastInsertId()

    var createdAt string
    err = s.DB.QueryRow("SELECT created_at FROM passwords WHERE id = ?", lastID).Scan(&createdAt)
    if err != nil {
        return 0, "", err
    }

    return lastID, createdAt, nil
}

// Retrieve all entries without passwords
func (s *SQLStorage) GetAllPasswords() ([]model.PasswordListItem, error) {
    rows, err := s.DB.Query("SELECT id, service, username, link, created_at FROM passwords")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var list []model.PasswordListItem
    for rows.Next() {
        var item model.PasswordListItem
        if err := rows.Scan(&item.ID, &item.Service, &item.Username, &item.Link, &item.CreatedAt); err != nil {
            return nil, err
        }
        list = append(list, item)
    }
    return list, nil
}

// Retrieve a single entry without password
func (s *SQLStorage) GetPasswordByID(id string) (model.PasswordListItem, error) {
    var p model.PasswordListItem
    err := s.DB.QueryRow("SELECT id, service, username, link, created_at FROM passwords WHERE id = ?", id).
        Scan(&p.ID, &p.Service, &p.Username, &p.Link, &p.CreatedAt)
    return p, err
}

// Retrieve encrypted password
func (s *SQLStorage) GetEncryptedPassword(id string) (string, error) {
    var encrypted string
    err := s.DB.QueryRow("SELECT password FROM passwords WHERE id = ?", id).Scan(&encrypted)
    return encrypted, err
}

// Update an existing password entry
func (s *SQLStorage) UpdatePassword(id string, p model.Password) error {
    _, err := s.DB.Exec(
        "UPDATE passwords SET service = ?, username = ?, link = ?, password = ?, category = ? WHERE id = ?",
        p.Service, p.Username, p.Link, p.Password, p.Category, id,
    )
    return err
}

// Delete a password entry
func (s *SQLStorage) DeletePassword(id string) error {
    _, err := s.DB.Exec("DELETE FROM passwords WHERE id = ?", id)
    return err
}

// Close the database connection
func (s *SQLStorage) Close() error {
    return s.DB.Close()
}

// Retrieve entries filtered by service, username, and category
func (s *SQLStorage) GetFilteredPasswords(service, username, category string) ([]model.PasswordListItem, error) {
    query := "SELECT id, service, username, link, category, created_at FROM passwords WHERE 1=1"
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
        if err := rows.Scan(&item.ID, &item.Service, &item.Username, &item.Link, &item.Category, &item.CreatedAt); err != nil {
            return nil, err
        }
        list = append(list, item)
    }
    return list, nil
}
