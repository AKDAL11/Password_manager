// internal/gui/unlock_android.go
//go:build android

package gui

import (
    "errors"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"

    "password-manager/pkg/appandroid"
)


func LaunchWithUnlock(a fyne.App) {
    appInstance := appandroid.InitApp(a)
    if appInstance == nil {
        return
    }

    w := a.NewWindow("Unlock Password Manager")

    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder("Enter master password")

    form := widget.NewForm(widget.NewFormItem("Master Password", passwordEntry))

    unlockBtn := widget.NewButton("Unlock", func() {
        if ok := appInstance.VerifyMasterPassword(passwordEntry.Text); !ok {
            dialog.ShowError(errors.New("invalid master password"), w)
            return
        }
        w.Hide()
        ShowMainWindow(a, appInstance)
    })

    content := container.NewVBox(form, unlockBtn)
    w.SetContent(container.NewCenter(content))
    w.Resize(fyne.NewSize(360, 640))
    w.Show()
}
