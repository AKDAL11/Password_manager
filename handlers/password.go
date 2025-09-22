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

// добавление нового пароля
func CreatePassword(c echo.Context) error {
	var p models.Password
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid body"})
	}

	// шифруем пароль перед записью
	encryptedPass, err := utils.EncryptAES(p.Password)
	if err != nil {
		c.Logger().Error("Encrypt error:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "encryption failed"})
	}

	// записываем в БД
	res, err := db.DB.Exec(
		"INSERT INTO passwords (service, username, link, password) VALUES (?, ?, ?, ?)",
		p.Service, p.Username, p.Link, encryptedPass,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// получаем ID и дату создания
	lastID, _ := res.LastInsertId()
	var createdAt string
	err = db.DB.QueryRow("SELECT created_at FROM passwords WHERE id = ?", lastID).Scan(&createdAt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// возвращаем безопасный объект (без пароля)
	resp := models.PasswordListItem{
		ID:        int(lastID),
		Service:   p.Service,
		Username:  p.Username,
		Link:      p.Link,
		CreatedAt: createdAt,
	}

	return c.JSON(http.StatusCreated, resp)
}

// обновление пароля
func UpdatePassword(c echo.Context) error {
	id := c.Param("id")
	var p models.Password
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid body"})
	}

	// шифруем новый пароль
	encryptedPass, err := utils.EncryptAES(p.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "encryption failed"})
	}

	_, err = db.DB.Exec("UPDATE passwords SET service = ?, username = ?, link = ?, password = ? WHERE id = ?",
		p.Service, p.Username, p.Link, encryptedPass, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "Updated successfully"})
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

// копирует пароль в буфер обмена и очищает через 15 секунд
func CopyPassword(c echo.Context) error {
	id := c.Param("id")
	var encrypted string
	err := db.DB.QueryRow("SELECT password FROM passwords WHERE id = ?", id).Scan(&encrypted)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// расшифровываем пароль
	password, err := utils.DecryptAES(encrypted)
	if err != nil {
		c.Logger().Error("Decrypt error:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "decryption failed"})
	}

	// копируем в буфер
	if err := utils.CopyToClipboard(password); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to copy to clipboard"})
	}

	// через 10 секунд очищаем буфер
	go func() {
		time.Sleep(10 * time.Second)
		if err := utils.CopyToClipboard(""); err != nil {
			c.Logger().Error("Failed to clear clipboard:", err)
		}
	}()

	return c.JSON(http.StatusOK, map[string]string{"status": "Password copied to clipboard. Will be cleared in 10 seconds."})
}


