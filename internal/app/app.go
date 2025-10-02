package app

import (
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

func InitApp(e *echo.Echo, dbPath string) *App {
    // Загружаем/генерируем ключ
    key, err := utils.LoadEncryptionKey(nil)
    if err != nil {
        log.Fatal("Encryption key error:", err)
    }
    crypto := utils.NewCryptoService(key)

    // Инициализация БД
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

// Проверка мастер‑пароля: теперь только через БД
func (a *App) VerifyMasterPassword(input string) bool {
    return a.DB.VerifyMasterPassword(input) == nil
}
