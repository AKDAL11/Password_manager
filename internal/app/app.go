//app.go

package app

import (
    "log"
    "password-manager/internal/app/db"
    "password-manager/pkg/utils"

    "github.com/labstack/echo/v4"
)

type App struct {
    DB     db.Storage
    Crypto *utils.CryptoService
    Logger echo.Logger
}

// InitApp loads the encryption key, initializes the storage and crypto service
func InitApp(e *echo.Echo) *App {
    utils.InitKey()
    cryptoSvc := utils.NewCryptoService(utils.EncryptionKey)

    storage, err := db.InitDB("./passwords.db")
    if err != nil {
        log.Fatal(err)
    }

    return &App{
        DB:     storage,
        Crypto: cryptoSvc,
        Logger: e.Logger,
    }
}

