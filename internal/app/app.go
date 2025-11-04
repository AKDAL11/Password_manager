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

// Веб-инициализация
func InitApp(e *echo.Echo, dbPath string) *App {
    storage, err := db.InitDB(dbPath, nil)
    if err != nil {
        log.Fatal(err)
    }
    return &App{DB: storage, Crypto: nil, Logger: e.Logger}
}

// Десктоп-инициализация
func InitDesktopApp(dbPath string) *App {
    storage, err := db.InitDB(dbPath, nil)
    if err != nil {
        log.Fatal(err)
    }
    return &App{DB: storage, Crypto: nil, Logger: nil}
}

// Установка Crypto после успешной проверки пароля
func (a *App) SetCryptoFromKey(key []byte) {
    a.Crypto = utils.NewCryptoService(key)
    a.DB.SetCrypto(a.Crypto)
}

// Проверка наличия meta (соль+верификатор)
func (a *App) HasMeta() bool {
    if a.DB == nil {
        return false
    }
    return a.DB.HasMeta()
}

// Инициализация/проверка мастер-пароля
func (a *App) InitializeMasterWithPassword(password string) error {
    sqlStore, ok := a.DB.(*db.SQLStorage)
    if !ok {
        return errors.New("invalid storage")
    }
    key, err := db.LoadOrInitMasterFromDB(sqlStore.DB, password)
    if err != nil {
        return err
    }
    a.SetCryptoFromKey(key)
    return nil
}
