//go:build !android

package gui

import (
    "errors"
    "fmt"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/theme"
    "fyne.io/fyne/v2/widget"

    pmapp "password-manager/internal/app"
    "password-manager/internal/i18n"
)

func LaunchWithUnlock(a fyne.App) {
    factory := CurrentFactory()
    a.Settings().SetTheme(factory.Theme())

    appInstance := pmapp.InitDesktopApp("passwords.db")

    // Было: appInstance.DB.HasMasterPassword()
    // Стало: appInstance.HasMeta()
    if !appInstance.HasMeta() {
        showCreateMasterPasswordForm(a, appInstance)
        return
    }

    w := a.NewWindow(i18n.T("Unlock_Password_Manager"))
    w.Resize(factory.SmallWindowSize())
    w.CenterOnScreen()

    title := widget.NewLabel("🔐 " + i18n.T("Unlock_Password_Manager"))
    title.Alignment = fyne.TextAlignCenter
    title.TextStyle = fyne.TextStyle{Bold: true}

    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder(i18n.T("Enter_master_password"))

    help := widget.NewLabel(i18n.T("Enter_master_password_to_continue"))
    help.Alignment = fyne.TextAlignCenter

    unlockBtn := widget.NewButtonWithIcon(i18n.T("Unlock"), theme.ConfirmIcon(), func() {
        // Было: VerifyMasterPassword(...)
        // Стало: InitializeMasterWithPassword(...)
        if err := appInstance.InitializeMasterWithPassword(passwordEntry.Text); err != nil {
            dialog.ShowError(errors.New(i18n.T("invalid_master_password")), w)
            return
        }
        w.Hide()
        ShowMainWindow(a, appInstance)
    })
    unlockBtn.Importance = widget.HighImportance

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

    form := container.NewVBox(
        title,
        widget.NewSeparator(),
        help,
        passwordEntry,
        unlockBtn,
        widget.NewSeparator(),
        widget.NewLabel("🌐 "+i18n.T("Language")),
        langSelect,
    )

    w.SetContent(container.NewCenter(container.NewPadded(form)))
    w.Show()
}

func showCreateMasterPasswordForm(a fyne.App, appInstance *pmapp.App) {
    factory := CurrentFactory()
    a.Settings().SetTheme(factory.Theme())

    w := a.NewWindow(i18n.T("Create_Master_Password"))
    w.Resize(factory.SmallWindowSize())
    w.CenterOnScreen()

    title := widget.NewLabel("🔐 " + i18n.T("First-time_setup"))
    title.Alignment = fyne.TextAlignCenter
    title.TextStyle = fyne.TextStyle{Bold: true}

    emailEntry := widget.NewEntry()
    emailEntry.SetPlaceHolder(i18n.T("Enter_your_email"))

    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder(i18n.T("Create_Master_Password"))

    confirmEntry := widget.NewPasswordEntry()
    confirmEntry.SetPlaceHolder(i18n.T("Confirm_master_password"))

    hint := widget.NewLabel(i18n.T("Remember_master_password_hint"))
    hint.Alignment = fyne.TextAlignCenter

    save := widget.NewButtonWithIcon(i18n.T("Save"), theme.DocumentSaveIcon(), func() {
        if passwordEntry.Text != confirmEntry.Text {
            dialog.ShowError(errors.New(i18n.T("passwords_do_not_match")), w)
            return
        }
        if emailEntry.Text == "" || passwordEntry.Text == "" {
            dialog.ShowError(errors.New(i18n.T("email_and_password_required")), w)
            return
        }
        // Было: appInstance.DB.SaveMasterPassword(...)
        // Стало: InitializeMasterWithPassword(...)
        if err := appInstance.InitializeMasterWithPassword(passwordEntry.Text); err != nil {
            dialog.ShowError(err, w)
            return
        }
        dialog.ShowInformation(i18n.T("Success"), i18n.T("Master_password_saved"), w)
        w.Hide()
        LaunchWithUnlock(a)
    })
    save.Importance = widget.HighImportance

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

    form := container.NewVBox(
        title,
        widget.NewSeparator(),
        widget.NewLabel("📧 "+i18n.T("Email")), emailEntry,
        widget.NewLabel("🔑 "+i18n.T("Password")), passwordEntry,
        widget.NewLabel("✅ "+i18n.T("Confirm")), confirmEntry,
        hint,
        save,
        widget.NewSeparator(),
        widget.NewLabel("🌐 "+i18n.T("Language")),
        langSelect,
    )

    w.SetContent(container.NewCenter(container.NewPadded(form)))
    w.Show()
}
