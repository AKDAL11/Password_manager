package handlers

import (
	"database/sql"
	"net/http"
	"password-manager/db"
	"password-manager/models"
	"password-manager/utils"
	"time"

	"github.com/labstack/echo/v4"
)

var cryptoSvc *utils.CryptoService
func InitCryptoService(svc *utils.CryptoService) {
    cryptoSvc = svc
}

// список без паролей
func GetPasswords(c echo.Context) error {
	rows, err := db.DB.Query("SELECT id, service, username, link, created_at FROM passwords")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer rows.Close()

	var list []models.PasswordListItem
	for rows.Next() {
		var item models.PasswordListItem
		if err := rows.Scan(&item.ID, &item.Service, &item.Username, &item.Link, &item.CreatedAt); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		list = append(list, item)
	}
	return c.JSON(http.StatusOK, list)
}

// вывод одной записи без пароля
func GetPassword(c echo.Context) error {
	id := c.Param("id")
	var p models.PasswordListItem
	err := db.DB.QueryRow("SELECT id, service, username, link, created_at FROM passwords WHERE id = ?", id).
		Scan(&p.ID, &p.Service, &p.Username, &p.Link, &p.CreatedAt)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, p)
}

// CreatePassword handles adding a new password to the database
func CreatePassword(c echo.Context) error {
    var p models.Password

    // Parse request body into Password model
    if err := c.Bind(&p); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid body"})
    }

    // Encrypt the password before saving
    encryptedPass, err := cryptoSvc.Encrypt(p.Password)
    if err != nil {
        c.Logger().Error("Encrypt error:", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "encryption failed"})
    }

    // Save encrypted password to the database
    res, err := db.DB.Exec(
        "INSERT INTO passwords (service, username, link, password) VALUES (?, ?, ?, ?)",
        p.Service, p.Username, p.Link, encryptedPass,
    )
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }

    // Get the ID of the newly inserted record
    lastID, _ := res.LastInsertId()

    // Get the creation date of the new password
    var createdAt string
    err = db.DB.QueryRow("SELECT created_at FROM passwords WHERE id = ?", lastID).Scan(&createdAt)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }

    // Return password info without the actual password
    resp := models.PasswordListItem{
        ID:        int(lastID),
        Service:   p.Service,
        Username:  p.Username,
        Link:      p.Link,
        CreatedAt: createdAt,
    }

    return c.JSON(http.StatusCreated, resp)
}

// UpdatePassword handles updating an existing password
func UpdatePassword(c echo.Context) error {
    id := c.Param("id")
    var p models.Password

    // Parse request body
    if err := c.Bind(&p); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid body"})
    }

    // Encrypt the new password
    encryptedPass, err := cryptoSvc.Encrypt(p.Password)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "encryption failed"})
    }

    // Update the password record in the database
    _, err = db.DB.Exec("UPDATE passwords SET service = ?, username = ?, link = ?, password = ? WHERE id = ?",
        p.Service, p.Username, p.Link, encryptedPass, id)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }

    return c.JSON(http.StatusOK, map[string]string{"status": "Updated successfully"})
}

// CopyPassword decrypts the password and copies it to clipboard
func CopyPassword(c echo.Context) error {
    id := c.Param("id")
    var encrypted string

    // Get encrypted password from database
    err := db.DB.QueryRow("SELECT password FROM passwords WHERE id = ?", id).Scan(&encrypted)
    if err != nil {
        if err == sql.ErrNoRows {
            return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
        }
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }

    // Decrypt the password
    password, err := cryptoSvc.Decrypt(encrypted)
    if err != nil {
        c.Logger().Error("Decrypt error:", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "decryption failed"})
    }

    // Copy password to clipboard
    if err := utils.CopyToClipboard(password); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to copy to clipboard"})
    }

    // Clear clipboard after 10 seconds
    go func() {
        time.Sleep(10 * time.Second)
        if err := utils.CopyToClipboard(""); err != nil {
            c.Logger().Error("Failed to clear clipboard:", err)
        }
    }()

    return c.JSON(http.StatusOK, map[string]string{"status": "Password copied to clipboard. Will be cleared in 10 seconds."})
}


// удаление пароля
func DeletePassword(c echo.Context) error {
	id := c.Param("id")
	_, err := db.DB.Exec("DELETE FROM passwords WHERE id = ?", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
