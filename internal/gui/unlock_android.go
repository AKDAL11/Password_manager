//go:build android

package gui

import (
    "errors"
    "fmt"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/theme"
    "fyne.io/fyne/v2/widget"

    "password-manager/internal/i18n"
    pmapp "password-manager/internal/app"
    "password-manager/pkg/appandroid"
)

func LaunchWithUnlock(a fyne.App) {
    factory := CurrentFactory()
    a.Settings().SetTheme(factory.Theme())

    appInstance := appandroid.InitApp(a) // возвращает *pmapp.App
    if appInstance == nil {
        w := a.NewWindow("❌ " + i18n.T("Error"))
        label := widget.NewLabel("❌ " + i18n.T("DB_init_failed"))
        label.Alignment = fyne.TextAlignCenter
        label.TextStyle = fyne.TextStyle{Bold: true}

        card := container.NewVBox(
            widget.NewLabelWithStyle("❗ "+i18n.T("Initialization_error"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
            widget.NewSeparator(),
            container.NewCenter(label),
        )

        fyne.Do(func() {
            w.SetContent(container.NewPadded(card))
            w.Resize(factory.SmallWindowSize())
            w.Show()
        })
        return
    }

    _ = i18n.LoadLocale(i18n.CurrentLang())

    // Было: appInstance.HasMasterPassword()
    // Стало: проверка наличия meta (соль+верификатор)
    if !appInstance.HasMeta() {
        showRegistrationWindow(a, appInstance)
        return
    }

    showUnlockWindow(a, appInstance)
}

// --- Окно регистрации ---
func showRegistrationWindow(a fyne.App, appInstance *pmapp.App) {
    factory := CurrentFactory()
    w := a.NewWindow("🔐 " + i18n.T("Create_Master_Password"))
    w.Resize(factory.SmallWindowSize())
    w.CenterOnScreen()

    title := widget.NewLabel("🔐 " + i18n.T("First-time_setup"))
    title.Alignment = fyne.TextAlignCenter
    title.TextStyle = fyne.TextStyle{Bold: true}

    emailEntry := widget.NewEntry()
    emailEntry.SetPlaceHolder(i18n.T("Enter_your_email"))

    pass1 := widget.NewPasswordEntry()
    pass1.SetPlaceHolder(i18n.T("Create_Master_Password"))

    pass2 := widget.NewPasswordEntry()
    pass2.SetPlaceHolder(i18n.T("Confirm_master_password"))

    saveBtn := widget.NewButtonWithIcon(i18n.T("Save"), theme.ConfirmIcon(), func() {
        if pass1.Text == "" || emailEntry.Text == "" {
            dialog.ShowError(errors.New(i18n.T("email_and_password_required")), w)
            return
        }
        if pass1.Text != pass2.Text {
            dialog.ShowError(errors.New(i18n.T("passwords_do_not_match")), w)
            return
        }
        // Было: appInstance.SetMasterPassword(...)
        // Стало: инициализация meta через пароль
        if err := appInstance.InitializeMasterWithPassword(pass1.Text); err != nil {
            dialog.ShowError(err, w)
            return
        }
        dialog.ShowInformation(i18n.T("Success"), i18n.T("Master_password_saved"), w)
        w.Hide()
        showUnlockWindow(a, appInstance)
    })
    saveBtn.Importance = widget.HighImportance

    langSelect := widget.NewSelect([]string{"en", "ru", "be"}, func(lang string) {
        if err := i18n.LoadLocale(lang); err != nil {
            fmt.Println("Ошибка загрузки языка:", err)
            return
        }
        w.SetTitle("🔐 " + i18n.T("Create_Master_Password"))
        title.SetText("🔐 " + i18n.T("First-time_setup"))
        emailEntry.SetPlaceHolder(i18n.T("Enter_your_email"))
        pass1.SetPlaceHolder(i18n.T("Create_Master_Password"))
        pass2.SetPlaceHolder(i18n.T("Confirm_master_password"))
        saveBtn.SetText(i18n.T("Save"))
    })
    langSelect.SetSelected(i18n.CurrentLang())

    form := container.NewVBox(
        widget.NewLabelWithStyle("📧 "+i18n.T("Email"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
        emailEntry,
        widget.NewLabelWithStyle("🔑 "+i18n.T("Master_Password"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
        pass1,
        widget.NewLabelWithStyle("🔑 "+i18n.T("Confirm_master_password"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
        pass2,
    )

    actions := container.NewHBox(saveBtn, langSelect)

    card := container.NewVBox(title, widget.NewSeparator(), form, widget.NewSeparator(), actions)

    fyne.Do(func() {
        w.SetContent(container.NewCenter(container.NewPadded(card)))
        w.Show()
    })
}

// --- Окно разблокировки ---
func showUnlockWindow(a fyne.App, appInstance *pmapp.App) {
    factory := CurrentFactory()
    w := a.NewWindow("🔐 " + i18n.T("Unlock_Password_Manager"))
    configureWindow(w)
    w.Resize(factory.SmallWindowSize())
    w.CenterOnScreen()

    title := widget.NewLabel("🔐 " + i18n.T("Unlock_Password_Manager"))
    title.Alignment = fyne.TextAlignCenter
    title.TextStyle = fyne.TextStyle{Bold: true}

    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder(i18n.T("Enter_master_password"))

    unlockBtn := widget.NewButtonWithIcon(i18n.T("Unlock"), theme.LoginIcon(), func() {
        // Было: VerifyMasterPassword(...)
        // Стало: InitializeMasterWithPassword(...) — инициализация ключа через meta
        if err := appInstance.InitializeMasterWithPassword(passwordEntry.Text); err != nil {
            fyne.Do(func() {
                dialog.ShowError(errors.New(i18n.T("invalid_master_password")), w)
            })
            return
        }
        fyne.Do(func() {
            w.Hide()
            ShowMainWindow(a, appInstance)
        })
    })
    unlockBtn.Importance = widget.HighImportance

    langSelect := widget.NewSelect([]string{"en", "ru", "be"}, func(lang string) {
        if err := i18n.LoadLocale(lang); err != nil {
            fmt.Println("Ошибка загрузки языка:", err)
            return
        }
        fyne.Do(func() {
            w.SetTitle("🔐 " + i18n.T("Unlock_Password_Manager"))
            title.SetText("🔐 " + i18n.T("Unlock_Password_Manager"))
            passwordEntry.SetPlaceHolder(i18n.T("Enter_master_password"))
            unlockBtn.SetText(i18n.T("Unlock"))
        })
    })
    langSelect.SetSelected(i18n.CurrentLang())

    card := container.NewVBox(
        title,
        widget.NewSeparator(),
        widget.NewLabelWithStyle("🔑 "+i18n.T("Master_Password"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
        passwordEntry,
        widget.NewSeparator(),
        container.NewHBox(unlockBtn, langSelect),
    )

    fyne.Do(func() {
        w.SetContent(container.NewCenter(container.NewPadded(card)))
        w.Show()
    })
}
