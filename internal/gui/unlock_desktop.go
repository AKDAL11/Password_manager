// internal/gui/unlock_desktop.go
//go:build !android

package gui

import (
    "errors"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"

    "github.com/labstack/echo/v4"
    pmapp "password-manager/internal/app"
)

func LaunchWithUnlock(a fyne.App) {
    e := echo.New()
    appInstance := pmapp.InitApp(e, "passwords.db")

    if !appInstance.DB.HasMasterPassword() {
        showCreateMasterPasswordForm(a, appInstance)
        return
    }

    w := a.NewWindow("Unlock Password Manager")

    // Заголовок по центру
    title := widget.NewLabel("🔐 Unlock Password Manager")
    title.Alignment = fyne.TextAlignCenter

    // Поле ввода пароля
    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder("Enter master password")

    // Кнопка разблокировки
    unlockBtn := widget.NewButton("Unlock", func() {
        if ok := appInstance.VerifyMasterPassword(passwordEntry.Text); !ok {
            dialog.ShowError(errors.New("invalid master password"), w)
            return
        }
        w.Hide()
        ShowMainWindow(a, appInstance)
    })

    // Вертикальное расположение: заголовок → поле → кнопка
    content := container.NewVBox(
        title,
        passwordEntry,
        unlockBtn,
    )

    // Центрируем всё в окне
    w.SetContent(container.NewCenter(content))
    w.Resize(fyne.NewSize(400, 200))
    w.CenterOnScreen()
    w.Show()
}


func showCreateMasterPasswordForm(a fyne.App, appInstance *pmapp.App) {
    w := a.NewWindow("Create Master Password")

    emailEntry := widget.NewEntry()
    emailEntry.SetPlaceHolder("Enter your email")

    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder("Create master password")

    confirmEntry := widget.NewPasswordEntry()
    confirmEntry.SetPlaceHolder("Confirm master password")

    save := widget.NewButton("Save", func() {
        if passwordEntry.Text != confirmEntry.Text {
            dialog.ShowError(errors.New("passwords do not match"), w)
            return
        }
        if emailEntry.Text == "" || passwordEntry.Text == "" {
            dialog.ShowError(errors.New("email and password required"), w)
            return
        }
        if err := appInstance.DB.SaveMasterPassword(emailEntry.Text, passwordEntry.Text); err != nil {
            dialog.ShowError(err, w)
            return
        }
        dialog.ShowInformation("Success", "Master password saved", w)
        w.Close()
        LaunchWithUnlock(a)
    })

    content := container.NewVBox(
        widget.NewLabel("🔐 First-time setup"),
        emailEntry,
        passwordEntry,
        confirmEntry,
        save,
    )

    w.SetContent(container.NewCenter(content))
    w.Resize(fyne.NewSize(640, 420))
    w.Show()
}
