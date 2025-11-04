package appandroid

import (
    "log"

    "fyne.io/fyne/v2"
    pmapp "password-manager/internal/app"
    "password-manager/internal/app/db"
)

func InitApp(a fyne.App) *pmapp.App {
    dbPath := a.Storage().RootURI().Path() + "/passwords.db"

    storage, err := db.InitDB(dbPath, nil)
    if err != nil {
        log.Println("DB init error:", err)
        return nil
    }

    return &pmapp.App{
        DB:     storage,
        Crypto: nil,
    }
}
