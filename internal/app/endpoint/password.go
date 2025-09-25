// password.go

package endpoint

import (
    "net/http"
    "password-manager/internal/app"
    "password-manager/internal/app/model"
    "password-manager/pkg/utils"
    "time"
    "github.com/wagslane/go-password-validator"

    "github.com/labstack/echo/v4"
)

type Handler struct {
    App *app.App
}

// Регистрируем маршруты и передаём зависимости
func RegisterRoutes(e *echo.Echo, appInstance *app.App) {
    h := &Handler{App: appInstance}

    e.GET("/passwords", h.GetPasswords)
    e.GET("/passwords/:id", h.GetPassword)
    e.GET("/passwords", h.GetFilteredPasswords)
    e.POST("/passwords", h.CreatePassword)
    e.PUT("/passwords/:id", h.UpdatePassword)
    e.DELETE("/passwords/:id", h.DeletePassword)
    e.POST("/passwords/:id/copy", h.CopyPassword)
}

// список без паролей
func (h *Handler) GetPasswords(c echo.Context) error {
    list, err := h.App.DB.GetAllPasswords()
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, list)
}

// вывод одной записи без пароля
func (h *Handler) GetPassword(c echo.Context) error {
    id := c.Param("id")
    p, err := h.App.DB.GetPasswordByID(id)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, p)
}

const minEntropy = 60 // рекомендуемый минимум
// создание новой записи
func (h *Handler) CreatePassword(c echo.Context) error {
    var p model.Password
    if err := c.Bind(&p); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid body"})
    }

    // Оценка надёжности пароля
    err := passwordvalidator.Validate(p.Password, minEntropy)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error":  "Пароль слишком слабый",
            "reason": err.Error(),
        })
    }

    encryptedPass, err := h.App.Crypto.Encrypt(p.Password)
    if err != nil {
        h.App.Logger.Error("Encrypt error:", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "encryption failed"})
    }
    p.Password = encryptedPass

    id, createdAt, err := h.App.DB.CreatePassword(p)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }

    resp := model.PasswordListItem{
        ID:        int(id),
        Service:   p.Service,
        Username:  p.Username,
        Link:      p.Link,
        Category:  p.Category,
        CreatedAt: createdAt,
    }

    return c.JSON(http.StatusCreated, resp)
}

// обновление записи
func (h *Handler) UpdatePassword(c echo.Context) error {
    id := c.Param("id")
    var p model.Password
    if err := c.Bind(&p); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid body"})
    }

    encryptedPass, err := h.App.Crypto.Encrypt(p.Password)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "encryption failed"})
    }
    p.Password = encryptedPass

    if err := h.App.DB.UpdatePassword(id, p); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    
    return c.JSON(http.StatusOK, map[string]string{"status": "Updated successfully"})
}

// копирование пароля в буфер
func (h *Handler) CopyPassword(c echo.Context) error {
    id := c.Param("id")
    encrypted, err := h.App.DB.GetEncryptedPassword(id)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }

    password, err := h.App.Crypto.Decrypt(encrypted)
    if err != nil {
        h.App.Logger.Error("Decrypt error:", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "decryption failed"})
    }

    if err := utils.CopyToClipboard(password); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to copy to clipboard"})
    }

    go func() {
        time.Sleep(10 * time.Second)
        if err := utils.CopyToClipboard(""); err != nil {
            h.App.Logger.Error("Failed to clear clipboard:", err)
        }
    }()

    return c.JSON(http.StatusOK, map[string]string{"status": "Password copied to clipboard. Will be cleared in 10 seconds."})
}

// удаление записи
func (h *Handler) DeletePassword(c echo.Context) error {
    id := c.Param("id")
    if err := h.App.DB.DeletePassword(id); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    return c.NoContent(http.StatusNoContent)
}

// filter
func (h *Handler) GetFilteredPasswords(c echo.Context) error {
    service := c.QueryParam("service")
    username := c.QueryParam("username")
    category := c.QueryParam("category")

    list, err := h.App.DB.GetFilteredPasswords(service, username, category)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, list)
}
