package handlers

import (
	"net/http"
	"password-manager/db"
	"password-manager/models"
	"password-manager/utils"

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

	res, err := db.DB.Exec("INSERT INTO passwords (service, username, link, password) VALUES (?, ?, ?, ?)",
		p.Service, p.Username, p.Link, p.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	lastID, _ := res.LastInsertId()
	p.ID = int(lastID)
	err = db.DB.QueryRow("SELECT created_at FROM passwords WHERE id = ?", p.ID).Scan(&p.CreatedAt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, p)
}

// обновление пароля
func UpdatePassword(c echo.Context) error {
	id := c.Param("id")
	var p models.Password
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid body"})
	}

	_, err := db.DB.Exec("UPDATE passwords SET service = ?, username = ?, link = ?, password = ? WHERE id = ?",
		p.Service, p.Username, p.Link, p.Password, id)
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

// копирование пароля в буфер обмена
func CopyPassword(c echo.Context) error {
	id := c.Param("id")
	var p models.Password
	err := db.DB.QueryRow("SELECT password FROM passwords WHERE id = ?", id).Scan(&p.Password)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if err := utils.CopyToClipboard(p.Password); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to copy to clipboard"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "Password copied to clipboard"})
}
