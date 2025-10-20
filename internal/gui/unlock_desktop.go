//go:build !android

package gui

import (
    "errors"
    "fmt"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"

    pmapp "password-manager/internal/app"
    "password-manager/internal/i18n"
)

func LaunchWithUnlock(a fyne.App) {
    factory := CurrentFactory()
    a.Settings().SetTheme(factory.Theme())

    appInstance := pmapp.InitDesktopApp("passwords.db")

    if !appInstance.DB.HasMasterPassword() {
        showCreateMasterPasswordForm(a, appInstance)
        return
    }

    w := a.NewWindow(i18n.T("Unlock_Password_Manager"))
    w.Resize(factory.WindowSize())
    w.CenterOnScreen()

    title := widget.NewLabel("🔐 " + i18n.T("Unlock_Password_Manager"))
    title.Alignment = fyne.TextAlignCenter

    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder(i18n.T("Enter_master_password"))

    unlockBtn := widget.NewButton(i18n.T("Unlock"), func() {
        if ok := appInstance.VerifyMasterPassword(passwordEntry.Text); !ok {
            dialog.ShowError(errors.New(i18n.T("invalid_master_password")), w)
            return
        }
        w.Hide()
        ShowMainWindow(a, appInstance)
    })

    langSelect := widget.NewSelect([]string{"en", "ru", "be"}, func(lang string) {
        if err := i18n.LoadLocale(lang); err != nil {
            fmt.Println("Ошибка загрузки языка:", err)
            return
        }
        w.SetTitle(i18n.T("Unlock_Password_Manager"))
        title.SetText("🔐 " + i18n.T("Unlock_Password_Manager"))
        passwordEntry.SetPlaceHolder(i18n.T("Enter_master_password"))
        unlockBtn.SetText(i18n.T("Unlock"))
    })
    langSelect.SetSelected(i18n.CurrentLang())

    content := container.NewVBox(
        title,
        passwordEntry,
        unlockBtn,
        langSelect,
    )

    w.SetContent(container.NewCenter(content))
    w.Show()
}

func showCreateMasterPasswordForm(a fyne.App, appInstance *pmapp.App) {
    factory := CurrentFactory()
    a.Settings().SetTheme(factory.Theme())

    w := a.NewWindow(i18n.T("Create_Master_Password"))
    w.Resize(factory.WindowSize())
    w.CenterOnScreen()

    emailEntry := widget.NewEntry()
    emailEntry.SetPlaceHolder(i18n.T("Enter_your_email"))

    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder(i18n.T("Create_Master_Password"))

    confirmEntry := widget.NewPasswordEntry()
    confirmEntry.SetPlaceHolder(i18n.T("Confirm_master_password"))

    save := widget.NewButton(i18n.T("Save"), func() {
        if passwordEntry.Text != confirmEntry.Text {
            dialog.ShowError(errors.New(i18n.T("passwords_do_not_match")), w)
            return
        }
        if emailEntry.Text == "" || passwordEntry.Text == "" {
            dialog.ShowError(errors.New(i18n.T("email_and_password_required")), w)
            return
        }
        if err := appInstance.DB.SaveMasterPassword(emailEntry.Text, passwordEntry.Text); err != nil {
            dialog.ShowError(err, w)
            return
        }
        dialog.ShowInformation(i18n.T("Success"), i18n.T("Master_password_saved"), w)
        w.Hide()
        LaunchWithUnlock(a)
    })

    langSelect := widget.NewSelect([]string{"en", "ru", "be"}, func(lang string) {
        if err := i18n.LoadLocale(lang); err != nil {
            fmt.Println("Ошибка загрузки языка:", err)
            return
        }
        w.SetTitle(i18n.T("Create_Master_Password"))
        emailEntry.SetPlaceHolder(i18n.T("Enter_your_email"))
        passwordEntry.SetPlaceHolder(i18n.T("Create_Master_Password"))
        confirmEntry.SetPlaceHolder(i18n.T("Confirm_master_password"))
        save.SetText(i18n.T("Save"))
    })
    langSelect.SetSelected(i18n.CurrentLang())

    content := container.NewVBox(
        widget.NewLabel("🔐 " + i18n.T("First-time_setup")),
        emailEntry,
        passwordEntry,
        confirmEntry,
        save,
        langSelect,
    )

    w.SetContent(container.NewCenter(content))
    w.Show()
}
