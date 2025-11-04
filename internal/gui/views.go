package gui

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"password-manager/internal/app"
	"password-manager/internal/app/db"
	"password-manager/internal/app/model"
	"password-manager/internal/i18n"
	"password-manager/pkg/utils"
)

func ShowMainWindow(a fyne.App, appInstance *app.App) {
	factory := CurrentFactory()
	a.Settings().SetTheme(factory.Theme())

	w := a.NewWindow(i18n.T("Password_Manager"))
	configureWindow(w)
	w.SetOnClosed(func() { a.Quit() })
	w.Resize(factory.WindowSize())
	w.CenterOnScreen()

	// --- Автоблокировка при бездействии ---
	const idleTimeout = 2 * time.Minute
	var idleTimer *time.Timer

	resetIdleTimer := func() {
		if idleTimer != nil {
			idleTimer.Stop()
		}
		idleTimer = time.AfterFunc(idleTimeout, func() {
			fyne.Do(func() {

				w.Hide()

				LaunchWithUnlock(a)

			})
		})
	}

	canvas := w.Canvas()
	canvas.SetOnTypedKey(func(ev *fyne.KeyEvent) { resetIdleTimer() })
	canvas.SetOnTypedRune(func(r rune) { resetIdleTimer() })

	// запускаем таймер при старте
	resetIdleTimer()

	// Загружаем список
	passwords, err := appInstance.DB.GetAllPasswords()
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	currentList := &passwords

	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextTruncate
	statusLabel.Alignment = fyne.TextAlignLeading

	cryptoSvc := appInstance.Crypto

	var table *widget.Table
	table, tableContent := buildPasswordTable(currentList, statusLabel, w, cryptoSvc, appInstance.DB)

	// Сохраняем ссылки на элементы, чтобы обновлять при смене языка
	welcomeLabel := widget.NewLabel("🔐 " + i18n.T("Welcome_to_Manager"))
	welcomeLabel.Alignment = fyne.TextAlignCenter
	welcomeLabel.TextStyle = fyne.TextStyle{Bold: true}

	headerLabel := widget.NewLabel("🔑 " + i18n.T("Your_Passwords"))
	headerLabel.TextStyle = fyne.TextStyle{Bold: true}

	addBtn := widget.NewButtonWithIcon(i18n.T("Add"), theme.ContentAddIcon(), func() {
		ShowCreateForm(a, appInstance, func() {
			newList, _ := appInstance.DB.GetAllPasswords()
			*currentList = newList
			table.Refresh()
		})
	})
	updateBtn := widget.NewButtonWithIcon(i18n.T("Update"), theme.DocumentCreateIcon(), func() {
		ShowUpdateWindow(a, appInstance, func() {
			newList, _ := appInstance.DB.GetAllPasswords()
			*currentList = newList
			table.Refresh()
		})
	})
	deleteBtn := widget.NewButtonWithIcon(i18n.T("Delete"), theme.DeleteIcon(), func() {
		ShowDeleteWindow(a, appInstance, func() {
			newList, _ := appInstance.DB.GetAllPasswords()
			*currentList = newList
			table.Refresh()
		})
	})
	filterBtn := widget.NewButtonWithIcon(i18n.T("Show_Filters"), theme.SearchIcon(), func() {
		ShowFilterWindow(a, appInstance)
	})

	mainContent := container.NewBorder(
		container.NewVBox(welcomeLabel, headerLabel, widget.NewSeparator()),
		nil, nil, nil,
		tableContent,
	)

	if fyne.CurrentDevice().IsMobile() {
		sidebarTop := container.NewVBox(addBtn, updateBtn, deleteBtn, filterBtn)
		sidebarContent := container.NewBorder(nil, nil, nil, nil, sidebarTop)

		tabs := container.NewAppTabs(
			container.NewTabItem(i18n.T("Menu"), sidebarContent),
			container.NewTabItem(i18n.T("Passwords"), mainContent),
		)
		tabs.SetTabLocation(container.TabLocationBottom)

		// обновление при смене языка
		langSelect := widget.NewSelect([]string{"en", "ru", "be"}, func(lang string) {
			if err := i18n.LoadLocale(lang); err != nil {
				dialog.ShowError(err, w)
				return
			}
			w.SetTitle(i18n.T("Password_Manager"))
			welcomeLabel.SetText("🔐 " + i18n.T("Welcome_to_Manager"))
			headerLabel.SetText("🔑 " + i18n.T("Your_Passwords"))
			addBtn.SetText(i18n.T("Add"))
			updateBtn.SetText(i18n.T("Update"))
			deleteBtn.SetText(i18n.T("Delete"))
			filterBtn.SetText(i18n.T("Show_Filters"))
			tabs.Items[0].Text = i18n.T("Menu")
			tabs.Items[1].Text = i18n.T("Passwords")
			tabs.Refresh()
		})
		langSelect.SetSelected(i18n.CurrentLang())

		w.SetContent(container.NewBorder(nil, langSelect, nil, nil, tabs))
	} else {
		langSelect := widget.NewSelect([]string{"en", "ru", "be"}, nil)
		langSelect.SetSelected(i18n.CurrentLang())

		sidebarTop := container.NewVBox(addBtn, updateBtn, deleteBtn, filterBtn)
		sidebarBottom := container.NewVBox(widget.NewSeparator(), langSelect)
		sidebarContent := container.NewBorder(nil, sidebarBottom, nil, nil, sidebarTop)

		split := container.NewHSplit(sidebarContent, mainContent)
		split.Offset = 0.2
		w.SetContent(split)

		// обновление при смене языка
		langSelect.OnChanged = func(lang string) {
			if err := i18n.LoadLocale(lang); err != nil {
				dialog.ShowError(err, w)
				return
			}
			w.SetTitle(i18n.T("Password_Manager"))
			welcomeLabel.SetText("🔐 " + i18n.T("Welcome_to_Manager"))
			headerLabel.SetText("🔑 " + i18n.T("Your_Passwords"))
			addBtn.SetText(i18n.T("Add"))
			updateBtn.SetText(i18n.T("Update"))
			deleteBtn.SetText(i18n.T("Delete"))
			filterBtn.SetText(i18n.T("Show_Filters"))
			table.Refresh()
			split.Refresh()
		}
	}
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
	storage db.Storage,
) (*widget.Table, fyne.CanvasObject) {

	columnWidths := []float32{60, 180, 180, 140, 160, 220, 120}
	rowHeights := make(map[int]float32)

	// объявляем table заранее
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
			tap.onTap = nil

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
					text = t.Local().Format("02 Jan 2006, 15:04")
				} else {
					text = row.CreatedAt
				}
			case 5:
				text = row.Link
			case 6:
				text = i18n.T("Copy")
			}

			label.SetText(text)
			label.Alignment = fyne.TextAlignLeading
			label.Show()

			if cell.Col == 6 {
				tap.onTap = func() {
					sqlStore, ok := storage.(*db.SQLStorage)
					if !ok {
						dialog.ShowError(errors.New("storage backend does not support direct access"), w)
						return
					}

					encB64, err := sqlStore.GetEncryptedPasswordByID(row.ID)
					if err != nil {
						dialog.ShowError(err, w)
						return
					}

					if err := utils.CopyToClipboard(encB64, cryptoSvc); err != nil {
						dialog.ShowError(err, w)
						return
					}

					statusLabel.SetText(i18n.T("Password_copied"))
					clearStatusLater(statusLabel)

					go func() {
						time.Sleep(15 * time.Second)
						fyne.Do(func() {
							fyne.CurrentApp().Clipboard().SetContent("")
						})
					}()
				}
				return
			}

			if cell.Col == 5 {
				tap.onTap = func() {
					if strings.TrimSpace(row.Link) != "" {
						fyne.CurrentApp().Clipboard().SetContent(row.Link)
						statusLabel.SetText(i18n.T("Link_copied"))
						clearStatusLater(statusLabel)
					}
				}
				return
			}

			tap.onTap = func() {
				if rowHeights[cell.Row] == 0 {
					rowHeights[cell.Row] = 72
					table.SetRowHeight(cell.Row, 72)
					label.Wrapping = fyne.TextWrapWord
					label.Truncation = fyne.TextTruncateOff
				} else {
					rowHeights[cell.Row] = 0
					table.SetRowHeight(cell.Row, 32)
					label.Wrapping = fyne.TextWrapOff
					label.Truncation = fyne.TextTruncateEllipsis
				}
				table.Refresh()
			}
		},
	)

	for i := range tableColumns {
		if i < len(columnWidths) {
			table.SetColumnWidth(i, columnWidths[i])
		}
	}
	for r := 0; r < len(*currentList)+1; r++ {
		table.SetRowHeight(r, 32)
	}

	scroll := container.NewScroll(table)
	scroll.SetMinSize(fyne.NewSize(w.Canvas().Size().Width, w.Canvas().Size().Height*0.6))

	statusBox := container.NewVBox(widget.NewSeparator(), statusLabel)
	content := container.NewBorder(nil, statusBox, nil, nil, scroll)
	return table, content
}

func ShowCreateForm(a fyne.App, appInstance *app.App, onSuccess func()) {
	factory := CurrentFactory()
	a.Settings().SetTheme(factory.Theme())

	w := a.NewWindow(i18n.T("Create_Password"))
	w.Resize(factory.SmallWindowSize())
	w.CenterOnScreen()

	passwords, _ := appInstance.DB.GetAllPasswords()
	services, usernames, categories, links := extractSuggestions(passwords)

	// Поля ввода
	service := widget.NewSelectEntry(services)
	username := widget.NewSelectEntry(usernames)
	link := widget.NewSelectEntry(links)
	category := widget.NewSelectEntry(categories)
	passwordEntry := widget.NewPasswordEntry()
	localStatus := widget.NewLabel("")

	// Сила пароля
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
			strengthLabel.SetText("❌ " + i18n.T("Weak_password") + ": " + i18n.T("missing") + " " + strings.Join(missing, ", "))
		} else {
			if err := utils.ValidatePasswordStrength(p, 60); err != nil {
				strengthLabel.SetText("❌ " + i18n.T("Weak_password") + ": " + err.Error())
			} else {
				strengthLabel.SetText("✅ " + i18n.T("Strong_password"))
			}
		}
		strengthLabel.Refresh()
	}

	// Опции генерации
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

	generateBtn := widget.NewButtonWithIcon(i18n.T("Generate"), theme.ViewRefreshIcon(), func() {
		length, err := strconv.Atoi(lengthEntry.Text)
		if err != nil || length <= 0 {
			localStatus.SetText(i18n.T("Invalid_length"))
			clearStatusLater(localStatus)
			return
		}

		var password string
		for {
			password, err = utils.GeneratePassword(length, useUpper.Checked, useLower.Checked, useDigits.Checked, useSymbols.Checked, excludeEntry.Text)
			if err != nil {
				localStatus.SetText(i18n.T("Generation_error") + ": " + err.Error())
				clearStatusLater(localStatus)
				return
			}
			// Проверяем, что пароль проходит все условия
			if len(password) >= 8 &&
				strings.ContainsAny(password, "0123456789") &&
				strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") &&
				strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") &&
				strings.ContainsAny(password, "!@#$%^&*()-_=+[]{}<>?/") {
				break
			}
			// иначе генерируем заново
		}

		passwordEntry.SetText(password)
		localStatus.SetText(i18n.T("Generated_inserted"))
		clearStatusLater(localStatus)
	})
	generateBtn.Importance = widget.MediumImportance

	passwordRow := container.NewGridWithColumns(2,
		container.NewVBox(passwordEntry),
		container.NewVBox(generateBtn),
	)
	optionsGrid := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel(i18n.T("Length")), lengthEntry),
		container.NewVBox(widget.NewLabel(i18n.T("Exclude")), excludeEntry),
	)
	checkboxGrid := container.NewGridWithColumns(4, useUpper, useLower, useDigits, useSymbols)
	passwordSection := container.NewVBox(passwordRow, optionsGrid, checkboxGrid, strengthLabel)

	form := container.NewVBox(
		widget.NewLabelWithStyle("🔧 "+i18n.T("Service"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), service,
		widget.NewLabelWithStyle("👤 "+i18n.T("Username"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), username,
		widget.NewLabelWithStyle("🔗 "+i18n.T("Link"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), link,
		widget.NewLabelWithStyle("📂 "+i18n.T("Category"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), category,
		widget.NewLabelWithStyle("🔑 "+i18n.T("Password"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), passwordSection,
	)

	submitBtn := widget.NewButtonWithIcon(i18n.T("Save"), theme.ConfirmIcon(), func() {
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

		if _, _, err := appInstance.DB.CreatePassword(p); err != nil {
			dialog.ShowError(err, w)
			return
		}
		if onSuccess != nil {
			onSuccess()
		}
		w.Close()
	})
	submitBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		widget.NewLabelWithStyle("🆕 "+i18n.T("Create_Password"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		submitBtn,
		localStatus,
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(factory.SmallWindowSize())
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

	form := container.NewVBox(
		widget.NewLabelWithStyle("🔧 "+i18n.T("Service"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		service,
		widget.NewLabelWithStyle("👤 "+i18n.T("Username"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		username,
		widget.NewLabelWithStyle("📂 "+i18n.T("Category"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		category,
	)

	filterBtn := widget.NewButton("🔍 "+i18n.T("Filter"), func() {
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
		// используем единый CryptoService и DB
		_, content := buildPasswordTable(&list, statusLabel, w, appInstance.Crypto, appInstance.DB)

		resultBox.Objects = []fyne.CanvasObject{
			content,
			statusLabel,
		}
		resultBox.Refresh()
	})
	filterBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		widget.NewLabelWithStyle("🔍 "+i18n.T("Filter_Passwords"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		filterBtn,
		resultBox,
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(factory.WindowSize())

	w.SetContent(container.NewPadded(scroll))
	w.Show()
}

func ShowUpdateWindow(a fyne.App, appInstance *app.App, onSuccess func()) {
	factory := CurrentFactory()
	a.Settings().SetTheme(factory.Theme())

	w := a.NewWindow(i18n.T("Update_Password"))
	w.Resize(factory.SmallWindowSize())
	w.CenterOnScreen()

	passwords, _ := appInstance.DB.GetAllPasswords()
	services, usernames, categories, links := extractSuggestions(passwords)

	// Поля ввода
	idEntry := widget.NewEntry()
	service := widget.NewSelectEntry(services)
	username := widget.NewSelectEntry(usernames)
	link := widget.NewSelectEntry(links)
	category := widget.NewSelectEntry(categories)
	passwordEntry := widget.NewPasswordEntry()
	localStatus := widget.NewLabel("")

	// Сила пароля
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
			strengthLabel.SetText("❌ " + i18n.T("Weak_password") + ": " + i18n.T("missing") + " " + strings.Join(missing, ", "))
		} else {
			if err := utils.ValidatePasswordStrength(p, 60); err != nil {
				strengthLabel.SetText("❌ " + i18n.T("Weak_password") + ": " + err.Error())
			} else {
				strengthLabel.SetText("✅ " + i18n.T("Strong_password"))
			}
		}
		strengthLabel.Refresh()
	}

	// Генерация пароля — по умолчанию длина 16, все опции включены
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

	generateBtn := widget.NewButtonWithIcon(i18n.T("Generate"), theme.ViewRefreshIcon(), func() {
		length, err := strconv.Atoi(lengthEntry.Text)
		if err != nil || length <= 0 {
			localStatus.SetText(i18n.T("Invalid_length"))
			clearStatusLater(localStatus)
			return
		}
		password, err := utils.GeneratePassword(
			length,
			useUpper.Checked,
			useLower.Checked,
			useDigits.Checked,
			useSymbols.Checked,
			excludeEntry.Text,
		)
		if err != nil {
			localStatus.SetText(i18n.T("Generation_error") + ": " + err.Error())
			clearStatusLater(localStatus)
			return
		}
		passwordEntry.SetText(password)
		localStatus.SetText(i18n.T("Generated_inserted"))
		clearStatusLater(localStatus)
	})
	generateBtn.Importance = widget.MediumImportance

	// Разметка секции пароля
	passwordRow := container.NewGridWithColumns(2,
		container.NewVBox(passwordEntry),
		container.NewVBox(generateBtn),
	)
	optionsGrid := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel(i18n.T("Length")), lengthEntry),
		container.NewVBox(widget.NewLabel(i18n.T("Exclude")), excludeEntry),
	)
	checkboxGrid := container.NewGridWithColumns(4,
		useUpper, useLower, useDigits, useSymbols,
	)
	passwordSection := container.NewVBox(passwordRow, optionsGrid, checkboxGrid, strengthLabel)

	// Форма
	form := container.NewVBox(
		widget.NewLabelWithStyle("🆔 "+i18n.T("ID"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), idEntry,
		widget.NewLabelWithStyle("🔧 "+i18n.T("Service"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), service,
		widget.NewLabelWithStyle("👤 "+i18n.T("Username"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), username,
		widget.NewLabelWithStyle("🔗 "+i18n.T("Link"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), link,
		widget.NewLabelWithStyle("📂 "+i18n.T("Category"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), category,
		widget.NewLabelWithStyle("🔑 "+i18n.T("Password"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), passwordSection,
	)

	// Сабмит с onSuccess
	submitBtn := widget.NewButtonWithIcon(i18n.T("Update"), theme.ConfirmIcon(), func() {
		idStr := strings.TrimSpace(idEntry.Text)
		if idStr == "" {
			dialog.ShowInformation(i18n.T("Info"), i18n.T("Please_enter_ID"), w)
			return
		}
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

		if err := appInstance.DB.UpdatePassword(idStr, p); err != nil {
			dialog.ShowError(err, w)
			return
		}
		if onSuccess != nil {
			onSuccess()
		}
		w.Close()
	})
	submitBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		widget.NewLabelWithStyle("✏️ "+i18n.T("Update_Password"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		submitBtn,
		localStatus,
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(factory.SmallWindowSize())
	w.SetContent(container.NewPadded(scroll))
	w.Show()
}

func ShowDeleteWindow(a fyne.App, appInstance *app.App, onSuccess func()) {
	w := a.NewWindow(i18n.T("Delete"))
	w.Resize(fyne.NewSize(300, 150))
	w.CenterOnScreen()

	idEntry := widget.NewEntry()
	idEntry.SetPlaceHolder(i18n.T("Enter_ID"))

	deleteBtn := widget.NewButton(i18n.T("Delete"), func() {
		idStr := strings.TrimSpace(idEntry.Text)
		if idStr == "" {
			dialog.ShowInformation(i18n.T("Info"), i18n.T("Please_enter_ID"), w)
			return
		}
		if err := appInstance.DB.DeletePassword(idStr); err != nil {
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation(i18n.T("Success"), i18n.T("Deleted_successfully"), w)
		if onSuccess != nil {
			onSuccess()
		}
		w.Close()
	})

	content := container.NewVBox(
		widget.NewLabelWithStyle(i18n.T("Enter_ID_to_delete"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		idEntry,
		deleteBtn,
	)

	w.SetContent(content)
	w.Show()
}
