package gui

import (
	"image/color"
	"strconv"
	"strings"
	"time"

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

func ShowMainWindow(a fyne.App, appInstance *app.App) {
	w := a.NewWindow(i18n.T("Password_Manager"))
	// при закрытии главного окна завершаем всё приложение
	w.SetOnClosed(func() {
		a.Quit()
	})

	passwords, err := appInstance.DB.GetAllPasswords()
	if err != nil {
		dialog.ShowError(err, w)
		return
	}

	currentList := passwords
	statusLabel := widget.NewLabel("")
	table, tableContainer := buildPasswordTable(&currentList, statusLabel)

	// переменные для блокировки
	lastActivity := time.Now()
	isLocked := false
	const idleTimeout = 2 * time.Minute

	updateActivity := func() {
		if !isLocked {
			lastActivity = time.Now()
		}
	}

	// горутина для блокировки по таймауту
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
					spacer.SetMinSize(fyne.NewSize(400, 0))

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

					dialogWindow.Resize(fyne.NewSize(420, 200))
					dialogWindow.Show()
				})
			}
		}
	}()

	// элементы интерфейса
	actionsLabel := widget.NewLabel("📁 " + i18n.T("Actions"))
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
		// код фильтрации
	})
	headerLabel := widget.NewLabel(i18n.T("Your_Passwords"))

	// функция обновления UI при смене языка
	refreshUI := func() {
		w.SetTitle(i18n.T("Password_Manager"))
		actionsLabel.SetText("📁 " + i18n.T("Actions"))
		addBtn.SetText("➕ " + i18n.T("Add"))
		updateBtn.SetText("✏️ " + i18n.T("Update"))
		deleteBtn.SetText("❌ " + i18n.T("Delete"))
		refreshBtn.SetText("🔄 " + i18n.T("Refresh"))
		filterBtn.SetText("🔍 " + i18n.T("Show_Filters"))
		headerLabel.SetText(i18n.T("Your_Passwords"))

		// обновляем заголовки таблицы
		tableColumns[0] = i18n.T("ID")
		tableColumns[1] = i18n.T("Service")
		tableColumns[2] = i18n.T("Username")
		tableColumns[3] = i18n.T("Category")
		tableColumns[4] = i18n.T("Created_At")
		tableColumns[5] = i18n.T("Link")
		tableColumns[6] = i18n.T("Copy_Password")

		table.Refresh()
	}

	langSelect := widget.NewSelect([]string{"en", "ru", "be"}, func(lang string) {
		if err := i18n.LoadLocale(lang); err != nil {
			dialog.ShowError(err, w)
			return
		}
		refreshUI()
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
	sidebarBox.Resize(fyne.NewSize(200, 500))

	// основной контент: заголовок, фильтры и таблица
	mainContent := container.NewBorder(
		container.NewVBox(headerLabel, filterBtn),
		nil, nil, nil,
		tableContainer,
	)

	// разделитель: слева меню, справа таблица
	split := container.NewHSplit(sidebarBox, mainContent)
	split.Offset = 0.25

	w.SetContent(split)
	w.Resize(fyne.NewSize(1225, 600))
	w.CenterOnScreen()
	w.Show()
}

// clearStatusLater очищает статус через 3 секунды
func clearStatusLater(label *widget.Label) {
	go func() {
		time.Sleep(3 * time.Second)
		fyne.Do(func() {
			label.SetText("")
		})
	}()
}

// глобальные заголовки таблицы
var tableColumns = []string{
	i18n.T("ID"),
	i18n.T("Service"),
	i18n.T("Username"),
	i18n.T("Category"),
	i18n.T("Created_At"),
	i18n.T("Link"),
	i18n.T("Copy_Password"),
}

// таблица с заголовками и данными (заголовки = первая строка)
func buildPasswordTable(currentList *[]model.PasswordListItem, statusLabel *widget.Label) (*widget.Table, fyne.CanvasObject) {
	table := widget.NewTable(
		func() (int, int) { return len(*currentList) + 1, len(tableColumns) },
		func() fyne.CanvasObject {
			return container.NewMax(widget.NewLabel(""), widget.NewButton("", nil))
		},
		func(cell widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*fyne.Container).Objects[0].(*widget.Label)
			button := o.(*fyne.Container).Objects[1].(*widget.Button)
			label.Hide()
			button.Hide()

			if cell.Row == 0 {
				// заголовки
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

			switch cell.Col {
			case 0:
				label.SetText(strconv.Itoa(row.ID))
				label.Show()
			case 1:
				label.SetText(row.Service)
				label.Show()
			case 2:
				label.SetText(row.Username)
				label.Show()
			case 3:
				label.SetText(row.Category)
				label.Show()
			case 4:
				t, err := time.Parse(time.RFC3339, row.CreatedAt)
				if err != nil {
					label.SetText(row.CreatedAt)
				} else {
					label.SetText(t.Local().Format("02 January 2006, 15:04"))
				}
				label.Show()
			case 5:
				link := row.Link
				button.SetText(link)
				button.OnTapped = func() {
					utils.CopyToClipboard(link)
					statusLabel.SetText(i18n.T("Link_copied"))
					clearStatusLater(statusLabel)
				}
				button.Show()
			case 6:
				button.SetText(i18n.T("Copy_Password"))
				button.OnTapped = func() {
					utils.CopyToClipboard(row.Password)
					statusLabel.SetText(i18n.T("Password_copied"))
					clearStatusLater(statusLabel)
				}
				button.Show()
			}
		},
	)

	widths := []float32{30, 110, 140, 100, 190, 150, 150}
	for i, w := range widths {
		table.SetColumnWidth(i, w)
	}

	scroll := container.NewScroll(table)
	scroll.SetMinSize(fyne.NewSize(800, 300))

	statusBox := container.NewVBox(statusLabel)
	statusBox.Resize(fyne.NewSize(800, 30))

	content := container.NewBorder(nil, statusBox, nil, nil, scroll)
	return table, content
}

func ShowCreateForm(a fyne.App, appInstance *app.App) {
	w := a.NewWindow(i18n.T("Create_Password"))

	passwords, _ := appInstance.DB.GetAllPasswords()
	services, usernames, categories, links := extractSuggestions(passwords)

	// элементы формы
	service := widget.NewSelectEntry(services)
	username := widget.NewSelectEntry(usernames)
	link := widget.NewSelectEntry(links)
	category := widget.NewSelectEntry(categories)
	passwordEntry := widget.NewPasswordEntry()
	statusLabel := widget.NewLabel("")

	// подсказка о силе пароля
	strengthLabel := widget.NewLabel("")
	strengthLabel.TextStyle = fyne.TextStyle{Bold: true}
	strengthLabel.Wrapping = fyne.TextWrapWord
	strengthLabel.Resize(fyne.NewSize(480, 60))

	// обработчик изменения пароля
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
	lengthLabel := widget.NewLabel(i18n.T("Length"))
	lengthEntry.SetPlaceHolder(i18n.T("Length"))

	excludeEntry := widget.NewEntry()
	excludeLabel := widget.NewLabel(i18n.T("Exclude"))
	excludeEntry.SetPlaceHolder(i18n.T("Exclude_chars"))

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

	// компоновка
	passwordRow := container.NewGridWithColumns(2,
		container.NewVBox(passwordEntry),
		container.NewVBox(generateBtn),
	)

	optionsGrid := container.NewGridWithColumns(2,
		container.NewVBox(lengthLabel, lengthEntry),
		container.NewVBox(excludeLabel, excludeEntry),
	)

	checkboxGrid := container.NewGridWithColumns(4, useUpper, useLower, useDigits, useSymbols)

	passwordSection := container.NewVBox(
		passwordRow,
		optionsGrid,
		checkboxGrid,
		strengthLabel,
	)

	form := widget.NewForm(
		widget.NewFormItem(i18n.T("Service"), service),
		widget.NewFormItem(i18n.T("Username"), username),
		widget.NewFormItem(i18n.T("Link"), link),
		widget.NewFormItem(i18n.T("Password"), passwordSection),
		widget.NewFormItem(i18n.T("Category"), category),
	)

	// сохранение пароля без обязательных требований
	form.OnSubmit = func() {
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
	}

	// язык: селектор + обновление UI
	langLabel := widget.NewLabel("🌐 " + i18n.T("Language"))
	langSelect := widget.NewSelect([]string{"en", "ru", "be"}, func(lang string) {
		if err := i18n.LoadLocale(lang); err != nil {
			dialog.ShowError(err, w)
			return
		}
		refreshUI := func() {
			w.SetTitle(i18n.T("Create_Password"))
			form.Items[0].Text = i18n.T("Service")
			form.Items[1].Text = i18n.T("Username")
			form.Items[2].Text = i18n.T("Link")
			form.Items[3].Text = i18n.T("Password")
			form.Items[4].Text = i18n.T("Category")
			form.Refresh()

			lengthLabel.SetText(i18n.T("Length"))
			lengthEntry.SetPlaceHolder(i18n.T("Length"))
			excludeLabel.SetText(i18n.T("Exclude"))
			excludeEntry.SetPlaceHolder(i18n.T("Exclude_chars"))

			generateBtn.SetText("🔁 " + i18n.T("Generate"))
			langLabel.SetText("🌐 " + i18n.T("Language"))
		}
		refreshUI()
	})
	langSelect.SetSelected(i18n.CurrentLang())

	content := container.NewVBox(
		form,
		container.NewHBox(langLabel, langSelect),
		statusLabel,
	)
	w.SetContent(container.NewPadded(content))
	w.Resize(fyne.NewSize(560, 500))
	w.CenterOnScreen()
	w.Show()
}

func ShowFilterWindow(a fyne.App, appInstance *app.App) {
	w := a.NewWindow(i18n.T("Filter_Passwords"))

	passwords, _ := appInstance.DB.GetAllPasswords()
	services, usernames, categories, _ := extractSuggestions(passwords)

	service := widget.NewSelectEntry(services)
	username := widget.NewSelectEntry(usernames)
	category := widget.NewSelectEntry(categories)

	result := widget.NewLabel("")

	form := widget.NewForm(
		widget.NewFormItem(i18n.T("Service"), service),
		widget.NewFormItem(i18n.T("Username"), username),
		widget.NewFormItem(i18n.T("Category"), category),
	)

	form.OnSubmit = func() {
		list, err := appInstance.DB.GetFilteredPasswords(service.Text, username.Text, category.Text)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if len(list) == 0 {
			result.SetText(i18n.T("No_matching_entries"))
			return
		}
		var output string
		for _, p := range list {
			output += p.Service + " | " + p.Username + " | " + p.Category + "\n"
		}
		result.SetText(output)
	}

	w.SetContent(container.NewVBox(form, result))
	w.Resize(fyne.NewSize(400, 300))
	w.CenterOnScreen()
	w.Show()
}

func ShowUpdateWindow(a fyne.App, appInstance *app.App) {
	w := a.NewWindow(i18n.T("Update_Password"))

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
	w.Resize(fyne.NewSize(400, 300))
	w.CenterOnScreen()
	w.Show()
}

func ShowDeleteWindow(a fyne.App, appInstance *app.App) {
	w := a.NewWindow(i18n.T("Delete_Password"))

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
	w.Resize(fyne.NewSize(300, 150))
	w.CenterOnScreen()
	w.Show()
}
