package main

import (
    "errors"
    "os"

    "password-manager/internal/app"
    "password-manager/internal/gui"

    "fyne.io/fyne/v2"
    fyneApp "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"

    "github.com/joho/godotenv"
    "github.com/labstack/echo/v4"
)

func main() {
    _ = godotenv.Load()

    if os.Getenv("ENCRYPTION_KEY") == "" || os.Getenv("MASTER_PASSWORD") == "" {
        panic("Missing ENCRYPTION_KEY or MASTER_PASSWORD in .env")
    }

    a := fyneApp.New()
    w := a.NewWindow("Unlock Password Manager")

    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder("Enter master password")

    submit := widget.NewButton("Unlock", func() {
        if passwordEntry.Text != os.Getenv("MASTER_PASSWORD") {
            dialog.ShowError(errors.New("Invalid master password"), w)
            return
        }
        e := echo.New()
        appInstance := app.InitApp(e)
        gui.ShowMainWindow(a, appInstance)
        w.Close()
    })

    w.SetContent(container.NewVBox(
        widget.NewLabel("Master Password"),
        passwordEntry,
        submit,
    ))
    w.Resize(fyne.NewSize(300, 150))
    w.ShowAndRun()
}
