package app

import (
    "log"
    "os"

    "github.com/joho/godotenv"
    "password-manager/internal/app/db"
    "password-manager/pkg/utils"
    "github.com/labstack/echo/v4"
)

type App struct {
    DB                 db.Storage
    Crypto             *utils.CryptoService
    Logger             echo.Logger
    MasterPasswordHash string
}

// InitApp loads the encryption key, initializes the storage and crypto service
func InitApp(e *echo.Echo) *App {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    master := os.Getenv("MASTER_PASSWORD")
    if master == "" {
        log.Fatal("MASTER_PASSWORD not set in .env")
    }

    hashed, err := utils.HashPassword(master)
    if err != nil {
        log.Fatal(err)
    }

    utils.InitKey()
    cryptoSvc := utils.NewCryptoService(utils.EncryptionKey)

    storage, err := db.InitDB("./passwords.db")
    if err != nil {
        log.Fatal(err)
    }

    return &App{
        DB:                 storage,
        Crypto:             cryptoSvc,
        Logger:             e.Logger,
        MasterPasswordHash: string(hashed),
    }
}

func (a *App) VerifyMasterPassword(input string) bool {
    return input == os.Getenv("MASTER_PASSWORD")
}
