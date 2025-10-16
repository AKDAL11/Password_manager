package app

import (
    "errors"
    "log"

    "github.com/labstack/echo/v4"
    "password-manager/internal/app/db"
    "password-manager/pkg/utils"
)

type App struct {
    DB     db.Storage
    Crypto *utils.CryptoService
    Logger echo.Logger
}

// Веб-инициализация (с Echo)
func InitApp(e *echo.Echo, dbPath string) *App {
    key, err := utils.LoadEncryptionKey(nil)
    if err != nil {
        log.Fatal("Encryption key error:", err)
    }
    crypto := utils.NewCryptoService(key)

    storage, err := db.InitDB(dbPath, crypto)
    if err != nil {
        log.Fatal(err)
    }

    return &App{
        DB:     storage,
        Crypto: crypto,
        Logger: e.Logger,
    }
}

// Десктоп-инициализация (без Echo)
func InitDesktopApp(dbPath string) *App {
    key, err := utils.LoadEncryptionKey(nil)
    if err != nil {
        log.Fatal("Encryption key error:", err)
    }
    crypto := utils.NewCryptoService(key)

    storage, err := db.InitDB(dbPath, crypto)
    if err != nil {
        log.Fatal(err)
    }

    return &App{
        DB:     storage,
        Crypto: crypto,
        Logger: nil,
    }
}

// Проверка мастер‑пароля
func (a *App) VerifyMasterPassword(input string) bool {
    if a.DB == nil {
        return false
    }
    return a.DB.VerifyMasterPassword(input) == nil
}

// Проверка: есть ли мастер‑пароль
func (a *App) HasMasterPassword() bool {
    if a.DB == nil {
        return false
    }
    return a.DB.HasMasterPassword()
}

// Установка нового мастер‑пароля
func (a *App) SetMasterPassword(email, password string) error {
    if a.DB == nil {
        return errors.New("DB not initialized")
    }
    // ВАЖНО: Storage.SaveMasterPassword принимает 2 аргумента (email, plain)
    return a.DB.SaveMasterPassword(email, password)
}

// Получение email для восстановления (по желанию)
func (a *App) GetRecoveryEmail() (string, error) {
    if a.DB == nil {
        return "", errors.New("DB not initialized")
    }
    return a.DB.GetRecoveryEmail()
}
