package endpoint

import (
    "net/http"
    "strconv"
    "time"

    "password-manager/internal/app"
    "password-manager/internal/app/model"
    "password-manager/pkg/utils"

    "github.com/labstack/echo/v4"
    "github.com/wagslane/go-password-validator"
)

type Handler struct {
    App *app.App
}

const minEntropy = 60 // Recommended minimum entropy

// Register routes and inject dependencies
func RegisterRoutes(e *echo.Echo, appInstance *app.App) {
    h := &Handler{App: appInstance}

    e.GET("/passwords", h.GetFilteredPasswords)
    e.GET("/passwords/:id", h.GetPassword)
    e.GET("/generate-password", h.GeneratePassword)
    e.POST("/passwords", h.CreatePassword)
    e.PUT("/passwords/:id", h.UpdatePassword)
    e.DELETE("/passwords/:id", h.DeletePassword)
    e.POST("/passwords/:id/copy", h.CopyPassword)
}

// Retrieve all entries without passwords
func (h *Handler) GetPasswords(c echo.Context) error {
    list, err := h.App.DB.GetAllPasswords()
    if err != nil {
        return c.JSON(http.StatusInternalServerError, utils.JSONError("Failed to retrieve passwords"))
    }
    return c.JSON(http.StatusOK, list)
}

// Retrieve a single entry without the password
func (h *Handler) GetPassword(c echo.Context) error {
    id := c.Param("id")
    p, err := h.App.DB.GetPasswordByID(id)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, utils.JSONError("Failed to retrieve password"))
    }
    return c.JSON(http.StatusOK, p)
}

// Create a new password entry
func (h *Handler) CreatePassword(c echo.Context) error {
    var p model.Password
    if err := c.Bind(&p); err != nil {
        return c.JSON(http.StatusBadRequest, utils.JSONError("Invalid request body"))
    }

    if err := utils.ValidatePasswordStrength(p.Password, minEntropy); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error":  "Password is too weak",
            "reason": err.Error(),
        })
    }

    encryptedPass, err := h.App.Crypto.Encrypt(p.Password)
    if err != nil {
        h.App.Logger.Error("Encrypt error:", err)
        return c.JSON(http.StatusInternalServerError, utils.JSONError("Encryption failed"))
    }
    p.Password = encryptedPass

    id, createdAt, err := h.App.DB.CreatePassword(p)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, utils.JSONError("Failed to save password"))
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
        return c.JSON(http.StatusBadRequest, utils.JSONError("Invalid request body"))
    }

    encryptedPass, err := h.App.Crypto.Encrypt(p.Password)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, utils.JSONError("Encryption failed"))
    }
    p.Password = encryptedPass

    if err := h.App.DB.UpdatePassword(id, p); err != nil {
        return c.JSON(http.StatusInternalServerError, utils.JSONError("Failed to update password"))
    }

    return c.JSON(http.StatusOK, map[string]string{"status": "Updated successfully"})
}

// Copy password to clipboard
func (h *Handler) CopyPassword(c echo.Context) error {
    idStr := c.Param("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        return c.JSON(http.StatusBadRequest, utils.JSONError("Некорректный ID"))
    }

    encB64, err := h.App.DB.GetEncryptedPasswordByID(id)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, utils.JSONError("Не удалось получить зашифрованный пароль"))
    }

    if err := utils.CopyToClipboard(encB64, h.App.Crypto); err != nil {
        h.App.Logger.Error("Ошибка копирования:", err)
        return c.JSON(http.StatusInternalServerError, utils.JSONError("Ошибка при копировании"))
    }

    go func() {
        time.Sleep(10 * time.Second)
        _ = utils.CopyToClipboard("", h.App.Crypto)
    }()

    return c.JSON(http.StatusOK, map[string]string{
        "status": "Пароль скопирован в буфер обмена. Будет очищен через 10 секунд.",
    })
}

// Delete a password entry
func (h *Handler) DeletePassword(c echo.Context) error {
    id := c.Param("id")
    if err := h.App.DB.DeletePassword(id); err != nil {
        return c.JSON(http.StatusInternalServerError, utils.JSONError("Failed to delete password"))
    }
    return c.NoContent(http.StatusNoContent)
}

// Filter password entries or return all if no filters
func (h *Handler) GetFilteredPasswords(c echo.Context) error {
    service := c.QueryParam("service")
    username := c.QueryParam("username")
    category := c.QueryParam("category")

    if service == "" && username == "" && category == "" {
        return h.GetPasswords(c)
    }

    list, err := h.App.DB.GetFilteredPasswords(service, username, category)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, utils.JSONError("Failed to filter passwords"))
    }
    return c.JSON(http.StatusOK, list)
}

// Generate a password with custom settings and return entropy
func (h *Handler) GeneratePassword(c echo.Context) error {
    length, err := strconv.Atoi(c.QueryParam("length"))
    if err != nil || length <= 0 {
        return c.JSON(http.StatusBadRequest, utils.JSONError("Invalid length"))
    }

    exclude := c.QueryParam("exclude")
    useUpper := c.QueryParam("upper") == "true"
    useLower := c.QueryParam("lower") == "true"
    useDigits := c.QueryParam("digits") == "true"
    useSymbols := c.QueryParam("symbols") == "true"

    pass, err := utils.GeneratePassword(length, useUpper, useLower, useDigits, useSymbols, exclude)
    if err != nil {
        return c.JSON(http.StatusBadRequest, utils.JSONError(err.Error()))
    }

    entropy := passwordvalidator.GetEntropy(pass)

    return c.JSON(http.StatusOK, map[string]interface{}{
        "password": pass,
        "entropy":  entropy,
    })
}
