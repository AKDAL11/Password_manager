package gui

import (
	"crypto/rand"
	"encoding/base64"
	"image/color"
	"strconv"
	"strings"
	"time"

	"github.com/zalando/go-keyring"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"password-manager/internal/app"
	"password-manager/internal/app/model"
	"password-manager/internal/i18n"
	"password-manager/pkg/utils"
)

const (
	keyringService = "PasswordManager"
	keyringUser    = "encryption-key"
)

// генерирует новый ключ (32 байта)
func generateKey() []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	return key
}

// получает ключ из системного хранилища или создаёт новый (ключ хранится в base64)
func getOrCreateKey() []byte {
	keyStr, err := keyring.Get(keyringService, keyringUser)
	if err == keyring.ErrNotFound {
		key := generateKey()
		_ = keyring.Set(keyringService, keyringUser, base64.StdEncoding.EncodeToString(key))
		return key
	} else if err != nil {
		// если хранилище недоступно — аварийно генерируем ключ в памяти (старые зашифрованные данные не расшифруются)
		return generateKey()
	}
	decoded, decErr := base64.StdEncoding.DecodeString(keyStr)
	if decErr != nil || len(decoded) != 32 {
		// если повреждён ключ в хранилище — генерируем новый (лучше залогировать и попросить пользователя мигрировать)
		return generateKey()
	}
	return decoded
}

func ShowMainWindow(a fyne.App, appInstance *app.App) {
	factory := CurrentFactory()
	a.Settings().SetTheme(factory.Theme())

	w := a.NewWindow(i18n.T("Password_Manager"))
	w.SetOnClosed(func() { a.Quit() })
	w.Resize(factory.WindowSize())
	w.CenterOnScreen()

	passwords, err := appInstance.DB.GetAllPasswords()
	if err != nil {
		dialog.ShowError(err, w)
		return
	}

	currentList := passwords
	statusLabel := widget.NewLabel("")

	// Ключ шифрования
	key := getOrCreateKey()
	cryptoSvc := utils.NewCryptoService(key)
	appInstance.Crypto = cryptoSvc

	table, tableContainer := buildPasswordTable(&currentList, statusLabel, w, cryptoSvc)

	// Таймаут блокировки
	lastActivity := time.Now()
	isLocked := false
	const idleTimeout = 2 * time.Minute

	updateActivity := func() {
		if !isLocked {
			lastActivity = time.Now()
		}
	}

	go func() {
		for {
			time.Sleep(1 * time.Second)
			if time.Since(lastActivity) > idleTimeout && !isLocked {
				isLocked = true
				fyne.Do(func() {
					passwordEntry := widget.NewPasswordEntry()
					passwordEntry.SetPlaceHolder(i18n.T("Enter_master_password"))

					info := widget.NewLabel("🔒 " + i18n.T("Session_locked"))
					form := widget.NewForm(
						widget.NewFormItem(i18n.T("Master_Password"), passwordEntry),
					)

					spacer := canvas.NewRectangle(color.Transparent)
					spacer.SetMinSize(fyne.NewSize(300, 0))

					content := container.NewVBox(spacer, info, form)

					dialogWindow := dialog.NewCustomConfirm(
						i18n.T("Unlock_Session"),
						i18n.T("Unlock"),
						i18n.T("Exit"),
						content,
						func(confirm bool) {
							if confirm && appInstance.VerifyMasterPassword(passwordEntry.Text) {
								isLocked = false
								lastActivity = time.Now()
							} else {
								a.Quit()
							}
						}, w)

					dialogWindow.Resize(factory.SmallWindowSize())
					dialogWindow.Show()
				})
			}
		}
	}()

	// Элементы интерфейса
	actionsLabel := widget.NewLabel("📁 " + i18n.T("Actions"))
	headerLabel := widget.NewLabel(i18n.T("Your_Passwords"))
	headerLabel.TextStyle = fyne.TextStyle{Bold: true}


	addBtn := widget.NewButton("➕ "+i18n.T("Add"), func() {
		updateActivity()
		ShowCreateForm(a, appInstance)
	})
	updateBtn := widget.NewButton("✏️ "+i18n.T("Update"), func() {
		updateActivity()
		ShowUpdateWindow(a, appInstance)
	})
	deleteBtn := widget.NewButton("❌ "+i18n.T("Delete"), func() {
		updateActivity()
		ShowDeleteWindow(a, appInstance)
	})
	refreshBtn := widget.NewButton("🔄 "+i18n.T("Refresh"), func() {
		updateActivity()
		newList, err := appInstance.DB.GetAllPasswords()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		currentList = newList
		table.Length = func() (int, int) { return len(currentList) + 1, len(tableColumns) }
		table.Refresh()
	})
	filterBtn := widget.NewButton("🔍 "+i18n.T("Show_Filters"), func() {
		updateActivity()
		ShowFilterWindow(a, appInstance)
	})

	// Селектор языка
	langSelect := widget.NewSelect([]string{"en", "ru", "be"}, func(lang string) {
		if err := i18n.LoadLocale(lang); err != nil {
			dialog.ShowError(err, w)
			return
		}
		actionsLabel.SetText("📁 " + i18n.T("Actions"))
		addBtn.SetText("➕ " + i18n.T("Add"))
		updateBtn.SetText("✏️ " + i18n.T("Update"))
		deleteBtn.SetText("❌ " + i18n.T("Delete"))
		refreshBtn.SetText("🔄 " + i18n.T("Refresh"))
		filterBtn.SetText("🔍 " + i18n.T("Show_Filters"))
		headerLabel.SetText(i18n.T("Your_Passwords"))

		for i, key := range []string{"ID", "Service", "Username", "Category", "Created_At", "Link", "Password"} {
			tableColumns[i] = i18n.T(key)
		}
		table.Refresh()
	})
	langSelect.SetSelected(i18n.CurrentLang())

	sidebar := container.NewVBox(
		actionsLabel,
		addBtn,
		updateBtn,
		deleteBtn,
		refreshBtn,
		widget.NewLabel("🌐 "+i18n.T("Language")),
		langSelect,
	)

	sidebarBox := container.NewVBox(sidebar)
	sidebarBox.Resize(fyne.NewSize(factory.SidebarWidth(), w.Canvas().Size().Height))

	mainContent := container.NewBorder(
		container.NewVBox(headerLabel, filterBtn),
		nil, nil, nil,
		tableContainer,
	)

	split := container.NewHSplit(sidebarBox, mainContent)
	split.Offset = factory.SidebarRatio()

	w.SetContent(split)
	w.Show()
}

func clearStatusLater(label *widget.Label) {
	go func() {
		time.Sleep(3 * time.Second)
		fyne.Do(func() {
			label.SetText("")
		})
	}()
}

var tableColumns = []string{
	i18n.T("ID"),
	i18n.T("Service"),
	i18n.T("Username"),
	i18n.T("Category"),
	i18n.T("Created_At"),
	i18n.T("Link"),
	i18n.T("Password"),
}

func buildPasswordTable(
    currentList *[]model.PasswordListItem,
    statusLabel *widget.Label,
    w fyne.Window,
    cryptoSvc *utils.CryptoService,
) (*widget.Table, fyne.CanvasObject) {

    size := w.Canvas().Size()
    columnWidths := []float32{50, 150, 200, 120, 180, 200, 120}
    rowHeights := make(map[int]float32)

    var table *widget.Table
    table = widget.NewTable(
        func() (int, int) { return len(*currentList) + 1, len(tableColumns) },
        func() fyne.CanvasObject {
            lbl := widget.NewLabel("")
            lbl.Wrapping = fyne.TextWrapOff
            lbl.Truncation = fyne.TextTruncateEllipsis
            tap := newTapOverlay(nil)
            return container.NewMax(lbl, tap)
        },
        func(cell widget.TableCellID, o fyne.CanvasObject) {
            c := o.(*fyne.Container)
            label := c.Objects[0].(*widget.Label)
            tap := c.Objects[1].(*tapOverlay)

            label.Hide()

            if cell.Row == 0 {
                label.SetText(tableColumns[cell.Col])
                label.TextStyle = fyne.TextStyle{Bold: true}
                label.Alignment = fyne.TextAlignCenter
                label.Show()
                return
            }

            i := cell.Row - 1
            if i < 0 || i >= len(*currentList) {
                return
            }
            row := (*currentList)[i]

            var text string
            switch cell.Col {
            case 0:
                text = strconv.Itoa(row.ID)
            case 1:
                text = row.Service
            case 2:
                text = row.Username
            case 3:
                text = row.Category
            case 4:
                if t, err := time.Parse(time.RFC3339, row.CreatedAt); err == nil {
                    text = t.Local().Format("02 January 2006, 15:04")
                } else {
                    text = row.CreatedAt
                }
            case 5:
                text = row.Link
            case 6:
                text = i18n.T("Copy_Password")
            }

            label.SetText(text)
            label.Show()

            // колонка Password: копирование
            if cell.Col == 6 {
                tap.onTap = func() {
                    decrypted, err := cryptoSvc.Decrypt(row.Password)
                    if err != nil {
                        dialog.ShowError(err, w)
                        return
                    }
                    _ = utils.CopyToClipboard(decrypted)
                    statusLabel.SetText(i18n.T("Password_copied"))
                    clearStatusLater(statusLabel)
                }
                return
            }

            // колонка Link: копирование ссылки
            if cell.Col == 5 {
                tap.onTap = func() {
                    if row.Link != "" {
                        _ = utils.CopyToClipboard(row.Link)
                        statusLabel.SetText(i18n.T("Link_copied"))
                        clearStatusLater(statusLabel)
                    }
                }
                return
            }

            // остальные колонки: раскрытие строки
            tap.onTap = func() {
                if rowHeights[cell.Row] == 0 {
                    rowHeights[cell.Row] = 80
                    table.SetRowHeight(cell.Row, 80)
                    label.Wrapping = fyne.TextWrapWord
                    label.Truncation = fyne.TextTruncateOff
                } else {
                    rowHeights[cell.Row] = 0
                    table.SetRowHeight(cell.Row, 30)
                    label.Wrapping = fyne.TextWrapOff
                    label.Truncation = fyne.TextTruncateEllipsis
                }
                table.Refresh()
            }
        },
    )

    for i, wcol := range columnWidths {
        table.SetColumnWidth(i, wcol)
    }
    for r := 0; r < len(*currentList)+1; r++ {
        table.SetRowHeight(r, 30)
    }

    scroll := container.NewScroll(table)
    scroll.SetMinSize(fyne.NewSize(size.Width, size.Height*0.6))

    statusBox := container.NewVBox(statusLabel)
    statusBox.Resize(fyne.NewSize(size.Width, 30))

    content := container.NewBorder(nil, statusBox, nil, nil, scroll)
    return table, content
}

func ShowCreateForm(a fyne.App, appInstance *app.App) {
    factory := CurrentFactory()
    a.Settings().SetTheme(factory.Theme())

    w := a.NewWindow(i18n.T("Create_Password"))
    w.Resize(factory.WindowSize())
    w.CenterOnScreen()

    passwords, _ := appInstance.DB.GetAllPasswords()
    services, usernames, categories, links := extractSuggestions(passwords)

    service := widget.NewSelectEntry(services)
    username := widget.NewSelectEntry(usernames)
    link := widget.NewSelectEntry(links)
    category := widget.NewSelectEntry(categories)
    passwordEntry := widget.NewPasswordEntry()
    statusLabel := widget.NewLabel("")

    strengthLabel := widget.NewLabel("")
    strengthLabel.TextStyle = fyne.TextStyle{Bold: true}
    strengthLabel.Wrapping = fyne.TextWrapWord

    passwordEntry.OnChanged = func(p string) {
        missing := []string{}
        if len(p) < 8 {
            missing = append(missing, i18n.T("length_8"))
        }
        if !strings.ContainsAny(p, "0123456789") {
            missing = append(missing, i18n.T("digit"))
        }
        if !strings.ContainsAny(p, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
            missing = append(missing, i18n.T("uppercase"))
        }
        if !strings.ContainsAny(p, "abcdefghijklmnopqrstuvwxyz") {
            missing = append(missing, i18n.T("lowercase"))
        }
        if !strings.ContainsAny(p, "!@#$%^&*()-_=+[]{}<>?/") {
            missing = append(missing, i18n.T("symbol"))
        }

        if len(missing) > 0 {
            strengthLabel.SetText("❌ " + i18n.T("Weak_password") + ": " +
                i18n.T("missing") + " " + strings.Join(missing, ", "))
        } else {
            err := utils.ValidatePasswordStrength(p, 60)
            if err != nil {
                strengthLabel.SetText("❌ " + i18n.T("Weak_password") + ": " + err.Error())
            } else {
                strengthLabel.SetText("✅ " + i18n.T("Strong_password"))
            }
        }
        strengthLabel.Refresh()
    }

    lengthEntry := widget.NewEntry()
    lengthEntry.SetText("16")
    excludeEntry := widget.NewEntry()

    useUpper := widget.NewCheck("A-Z", nil)
    useUpper.SetChecked(true)
    useLower := widget.NewCheck("a-z", nil)
    useLower.SetChecked(true)
    useDigits := widget.NewCheck("0-9", nil)
    useDigits.SetChecked(true)
    useSymbols := widget.NewCheck("!@#", nil)
    useSymbols.SetChecked(true)

    generateBtn := widget.NewButton("🔁 "+i18n.T("Generate"), func() {
        length, err := strconv.Atoi(lengthEntry.Text)
        if err != nil || length <= 0 {
            statusLabel.SetText(i18n.T("Invalid_length"))
            clearStatusLater(statusLabel)
            return
        }
        password, err := utils.GeneratePassword(length, useUpper.Checked, useLower.Checked, useDigits.Checked, useSymbols.Checked, excludeEntry.Text)
        if err != nil {
            statusLabel.SetText(i18n.T("Generation_error") + ": " + err.Error())
            clearStatusLater(statusLabel)
            return
        }
        passwordEntry.SetText(password)
        statusLabel.SetText(i18n.T("Generated_inserted"))
        clearStatusLater(statusLabel)
    })

    passwordRow := container.NewGridWithColumns(2,
        container.NewVBox(passwordEntry),
        container.NewVBox(generateBtn),
    )

    optionsGrid := container.NewGridWithColumns(2,
        container.NewVBox(widget.NewLabel(i18n.T("Length")), lengthEntry),
        container.NewVBox(widget.NewLabel(i18n.T("Exclude")), excludeEntry),
    )

    checkboxGrid := container.NewGridWithColumns(4, useUpper, useLower, useDigits, useSymbols)

    passwordSection := container.NewVBox(
        passwordRow,
        optionsGrid,
        checkboxGrid,
        strengthLabel,
    )

    form := container.NewVBox(
        widget.NewLabel(i18n.T("Service")), service,
        widget.NewLabel(i18n.T("Username")), username,
        widget.NewLabel(i18n.T("Link")), link,
        widget.NewLabel(i18n.T("Password")), passwordSection,
        widget.NewLabel(i18n.T("Category")), category,
    )

    submitBtn := widget.NewButton(i18n.T("Save"), func() {
        p := model.Password{
            Service:   service.Text,
            Username:  username.Text,
            Link:      link.Text,
            Password:  passwordEntry.Text,
            Category:  category.Text,
            CreatedAt: time.Now().Format(time.RFC3339),
        }
        encrypted, err := appInstance.Crypto.Encrypt(p.Password)
        if err != nil {
            dialog.ShowError(err, w)
            return
        }
        p.Password = encrypted
        _, _, err = appInstance.DB.CreatePassword(p)
        if err != nil {
            dialog.ShowError(err, w)
            return
        }
        w.Close()
    })

    content := container.NewVBox(
        form,
        submitBtn,
        statusLabel,
    )

    scroll := container.NewVScroll(content)
    scroll.SetMinSize(factory.WindowSize())

    w.SetContent(container.NewPadded(scroll))
    w.Show()
}

func ShowFilterWindow(a fyne.App, appInstance *app.App) {
	factory := CurrentFactory()
	a.Settings().SetTheme(factory.Theme())

	w := a.NewWindow(i18n.T("Filter_Passwords"))
	w.Resize(factory.WindowSize())
	w.CenterOnScreen()

	passwords, _ := appInstance.DB.GetAllPasswords()
	services, usernames, categories, _ := extractSuggestions(passwords)

	service := widget.NewSelectEntry(services)
	service.PlaceHolder = i18n.T("Any")
	username := widget.NewSelectEntry(usernames)
	username.PlaceHolder = i18n.T("Any")
	category := widget.NewSelectEntry(categories)
	category.PlaceHolder = i18n.T("Any")

	resultBox := container.NewVBox(widget.NewLabel(i18n.T("No_results_yet")))

	form := widget.NewForm(
		widget.NewFormItem(i18n.T("Service"), service),
		widget.NewFormItem(i18n.T("Username"), username),
		widget.NewFormItem(i18n.T("Category"), category),
	)
	form.SubmitText = i18n.T("Filter")

	form.OnSubmit = func() {
		list, err := appInstance.DB.GetFilteredPasswords(service.Text, username.Text, category.Text)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if len(list) == 0 {
			resultBox.Objects = []fyne.CanvasObject{widget.NewLabel(i18n.T("No_matching_entries"))}
			resultBox.Refresh()
			return
		}

		statusLabel := widget.NewLabel("")
		cryptoSvc := appInstance.Crypto
		_, tableContainer := buildPasswordTable(&list, statusLabel, w, cryptoSvc)

		resultBox.Objects = []fyne.CanvasObject{
			tableContainer,
			statusLabel,
		}
		resultBox.Refresh()
	}

	w.SetContent(container.NewVBox(form, resultBox))
	w.Show()
}

func ShowUpdateWindow(a fyne.App, appInstance *app.App) {
	factory := CurrentFactory()
	a.Settings().SetTheme(factory.Theme())

	w := a.NewWindow(i18n.T("Update_Password"))
	w.Resize(factory.SmallWindowSize())
	w.CenterOnScreen()

	passwords, _ := appInstance.DB.GetAllPasswords()
	services, usernames, categories, links := extractSuggestions(passwords)

	idEntry := widget.NewEntry()
	service := widget.NewSelectEntry(services)
	username := widget.NewSelectEntry(usernames)
	link := widget.NewSelectEntry(links)
	password := widget.NewPasswordEntry()
	category := widget.NewSelectEntry(categories)

	form := widget.NewForm(
		widget.NewFormItem(i18n.T("ID"), idEntry),
		widget.NewFormItem(i18n.T("Service"), service),
		widget.NewFormItem(i18n.T("Username"), username),
		widget.NewFormItem(i18n.T("Link"), link),
		widget.NewFormItem(i18n.T("Password"), password),
		widget.NewFormItem(i18n.T("Category"), category),
	)

	form.OnSubmit = func() {
		p := model.Password{
			Service:  service.Text,
			Username: username.Text,
			Link:     link.Text,
			Password: password.Text,
			Category: category.Text,
		}
		encrypted, err := appInstance.Crypto.Encrypt(p.Password)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		p.Password = encrypted
		if err := appInstance.DB.UpdatePassword(idEntry.Text, p); err != nil {
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation(i18n.T("Updated"), i18n.T("Password_updated"), w)
		w.Close()
	}

	w.SetContent(form)
	w.Show()
}

func ShowDeleteWindow(a fyne.App, appInstance *app.App) {
	factory := CurrentFactory()
	a.Settings().SetTheme(factory.Theme())

	w := a.NewWindow(i18n.T("Delete_Password"))
	w.Resize(factory.SmallWindowSize())
	w.CenterOnScreen()

	idEntry := widget.NewEntry()
	idEntry.SetPlaceHolder(i18n.T("Enter_ID_to_delete"))

	deleteBtn := widget.NewButton(i18n.T("Delete"), func() {
		id := idEntry.Text
		if err := appInstance.DB.DeletePassword(id); err != nil {
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation(i18n.T("Deleted"), i18n.T("Password_deleted"), w)
		w.Close()
	})

	w.SetContent(container.NewVBox(
		idEntry,
		deleteBtn,
	))
	w.Show()
}
