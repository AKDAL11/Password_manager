// password.go
package endpoint

import (
    "net/http"
    "password-manager/internal/app"
    "password-manager/internal/app/model"
    "password-manager/pkg/utils"
    "time"
    "github.com/wagslane/go-password-validator"
    "strconv"

    "github.com/labstack/echo/v4"
)

type Handler struct {
    App *app.App
}

// Register routes and inject dependencies
func RegisterRoutes(e *echo.Echo, appInstance *app.App) {
    h := &Handler{App: appInstance}

    e.GET("/passwords", h.GetPasswords)
    e.GET("/passwords/:id", h.GetPassword)
    e.GET("/passwords", h.GetFilteredPasswords)
    e.GET("/generate-password", h.GeneratePassword)
    e.POST("/passwords", h.CreatePassword)
    e.PUT("/passwords/:id", h.UpdatePassword)
    e.DELETE("/passwords/:id", h.DeletePassword)
    e.POST("/passwords/:id/copy", h.CopyPassword)
}

// List of entries without passwords
func (h *Handler) GetPasswords(c echo.Context) error {
    list, err := h.App.DB.GetAllPasswords()
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, list)
}

// Retrieve a single entry without the password
func (h *Handler) GetPassword(c echo.Context) error {
    id := c.Param("id")
    p, err := h.App.DB.GetPasswordByID(id)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, p)
}

const minEntropy = 60 // Recommended minimum entropy

// Create a new password entry
func (h *Handler) CreatePassword(c echo.Context) error {
    var p model.Password
    if err := c.Bind(&p); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid body"})
    }

    // Evaluate password strength
    err := passwordvalidator.Validate(p.Password, minEntropy)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error":  "Password is too weak",
            "reason": err.Error(),
        })
    }

    encryptedPass, err := h.App.Crypto.Encrypt(p.Password)
    if err != nil {
        h.App.Logger.Error("Encrypt error:", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Encryption failed"})
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

// Update an existing password entry
func (h *Handler) UpdatePassword(c echo.Context) error {
    id := c.Param("id")
    var p model.Password
    if err := c.Bind(&p); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid body"})
    }

    encryptedPass, err := h.App.Crypto.Encrypt(p.Password)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Encryption failed"})
    }
    p.Password = encryptedPass

    if err := h.App.DB.UpdatePassword(id, p); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }

    return c.JSON(http.StatusOK, map[string]string{"status": "Updated successfully"})
}

// Copy password to clipboard
func (h *Handler) CopyPassword(c echo.Context) error {
    id := c.Param("id")
    encrypted, err := h.App.DB.GetEncryptedPassword(id)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }

    password, err := h.App.Crypto.Decrypt(encrypted)
    if err != nil {
        h.App.Logger.Error("Decrypt error:", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Decryption failed"})
    }

    if err := utils.CopyToClipboard(password); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to copy to clipboard"})
    }

    go func() {
        time.Sleep(10 * time.Second)
        if err := utils.CopyToClipboard(""); err != nil {
            h.App.Logger.Error("Failed to clear clipboard:", err)
        }
    }()

    return c.JSON(http.StatusOK, map[string]string{"status": "Password copied to clipboard. Will be cleared in 10 seconds."})
}

// Delete a password entry
func (h *Handler) DeletePassword(c echo.Context) error {
    id := c.Param("id")
    if err := h.App.DB.DeletePassword(id); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    return c.NoContent(http.StatusNoContent)
}

// Filter password entries
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

// Generate a password with custom settings
func (h *Handler) GeneratePassword(c echo.Context) error {
    length, err := strconv.Atoi(c.QueryParam("length"))
    if err != nil || length <= 0 {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid length"})
    }

    exclude := c.QueryParam("exclude")
    useUpper := c.QueryParam("upper") == "true"
    useLower := c.QueryParam("lower") == "true"
    useDigits := c.QueryParam("digits") == "true"
    useSymbols := c.QueryParam("symbols") == "true"

    pass, err := utils.GeneratePassword(length, useUpper, useLower, useDigits, useSymbols, exclude)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }

    return c.JSON(http.StatusOK, map[string]string{"password": pass})
}
