// pkg/appandroid/init.go
package appandroid

import (
    "log"

    "fyne.io/fyne/v2"
    pmapp "password-manager/internal/app"
    "password-manager/internal/app/db"
    "password-manager/pkg/utils"
)

func InitApp(a fyne.App) *pmapp.App {
    key, err := utils.LoadEncryptionKey(a)
    if err != nil {
        log.Println("Encryption key error:", err)
        return nil
    }
    crypto := utils.NewCryptoService(key)

    dbPath := a.Storage().RootURI().Path() + "/passwords.db"
    storage, err := db.InitDB(dbPath, crypto)
    if err != nil {
        log.Println("DB init error:", err)
        return nil
    }

    return &pmapp.App{
        DB:     storage,
        Crypto: crypto,
    }
}
