package gui

import (
	"password-manager/internal/app"
	"password-manager/internal/app/model"
	"password-manager/pkg/utils"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/theme"
)

func ShowMainWindow(a fyne.App, appInstance *app.App) {
	w := a.NewWindow("Password Manager")

	passwords, err := appInstance.DB.GetAllPasswords()
	if err != nil {
		dialog.ShowError(err, w)
		return
	}

	currentList := passwords
	statusLabel := widget.NewLabel("")
	table, tableContainer := buildPasswordTable(currentList, statusLabel)

	var filterDialog dialog.Dialog

	// Кнопка фильтра
	filterBtn := widget.NewButton("🔍 Show Filters", func() {
		// Получаем подсказки
		passwords, _ := appInstance.DB.GetAllPasswords()
		services, usernames, categories, _ := extractSuggestions(passwords)

		serviceFilter := widget.NewSelectEntry(services)
		serviceFilter.SetPlaceHolder("Service")
		serviceFilter.Resize(fyne.NewSize(300, 40))

		usernameFilter := widget.NewSelectEntry(usernames)
		usernameFilter.SetPlaceHolder("Username")
		usernameFilter.Resize(fyne.NewSize(300, 40))

		categoryFilter := widget.NewSelectEntry(categories)
		categoryFilter.SetPlaceHolder("Category")
		categoryFilter.Resize(fyne.NewSize(300, 40))

		form := widget.NewForm(
			widget.NewFormItem("Service", serviceFilter),
			widget.NewFormItem("Username", usernameFilter),
			widget.NewFormItem("Category", categoryFilter),
		)

		applyBtn := widget.NewButtonWithIcon("Apply", theme.ConfirmIcon(), func() {
			filtered, err := appInstance.DB.GetFilteredPasswords(serviceFilter.Text, usernameFilter.Text, categoryFilter.Text)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			currentList = filtered
			table.Length = func() (int, int) { return len(currentList), 7 }
			table.Refresh()
			filterDialog.Hide()
		})

		buttons := container.NewHBox(applyBtn)
		content := container.NewVBox(form, buttons)

		filterDialog = dialog.NewCustom("Filter Passwords", "Close", content, w)
		filterDialog.Resize(fyne.NewSize(500, 300))
		filterDialog.Show()
	})

	// Кнопка обновления
	refreshBtn := widget.NewButton("🔄 Refresh", func() {
		newList, err := appInstance.DB.GetAllPasswords()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		currentList = newList
		table.Length = func() (int, int) { return len(currentList), 7 }
		table.Refresh()
	})

	// Обновление ячеек таблицы
	table.UpdateCell = func(cell widget.TableCellID, o fyne.CanvasObject) {
		if cell.Row >= len(currentList) || cell.Col >= 7 {
			return
		}

		i, j := cell.Row, cell.Col
		label := o.(*fyne.Container).Objects[0].(*widget.Label)
		button := o.(*fyne.Container).Objects[1].(*widget.Button)

		label.Hide()
		button.Hide()

		switch j {
		case 0:
			label.SetText(strconv.Itoa(currentList[i].ID))
			label.Show()
		case 1:
			label.SetText(currentList[i].Service)
			label.Show()
		case 2:
			label.SetText(currentList[i].Username)
			label.Show()
		case 3:
			label.SetText(currentList[i].Category)
			label.Show()
		case 4:
			t, err := time.Parse(time.RFC3339, currentList[i].CreatedAt)
			if err != nil {
				label.SetText(currentList[i].CreatedAt)
			} else {
				label.SetText(t.Local().Format("02 January 2006, 15:04"))
			}
			label.Show()
		case 5:
			link := currentList[i].Link
			button.SetText(link)
			button.OnTapped = func() {
				utils.CopyToClipboard(link)
				statusLabel.SetText("Link copied to clipboard")
				clearStatusLater(statusLabel)
			}
			button.Show()
		case 6:
			button.SetText("Copy Password")
			button.OnTapped = func() {
				utils.CopyToClipboard(currentList[i].Password)
				statusLabel.SetText("Password copied to clipboard")
				clearStatusLater(statusLabel)
			}
			button.Show()
		}
	}

	// Боковая панель
	sidebar := container.NewVBox(
		widget.NewLabel("📁 Actions"),
		widget.NewButton("➕ Add", func() { ShowCreateForm(a, appInstance) }),
		widget.NewButton("✏️ Update", func() { ShowUpdateWindow(a, appInstance) }),
		widget.NewButton("❌ Delete", func() { ShowDeleteWindow(a, appInstance) }),
		refreshBtn,
	)

	sidebarBox := container.NewVBox(sidebar)
	sidebarBox.Resize(fyne.NewSize(200, 500))

	// Основное содержимое
	mainContent := container.NewBorder(
		container.NewVBox(widget.NewLabel("Your Passwords"), filterBtn),
		nil, nil, nil,
		tableContainer,
	)

	split := container.NewHSplit(sidebarBox, mainContent)
	split.Offset = 0.25

	w.SetContent(split)
	w.Resize(fyne.NewSize(1225, 600))
	w.CenterOnScreen()
	w.Show()
}

func buildPasswordTable(passwords []model.PasswordListItem, statusLabel *widget.Label) (*widget.Table, fyne.CanvasObject) {
	columns := []string{"ID", "Service", "Username", "Category", "Link", "Created At", "Copy Password"}

	headerTable := widget.NewTable(
		func() (int, int) { return 1, len(columns) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(cell widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(columns[cell.Col])
			o.(*widget.Label).TextStyle = fyne.TextStyle{Bold: true}
			o.(*widget.Label).Alignment = fyne.TextAlignCenter
		},
	)

	dataTable := widget.NewTable(
		func() (int, int) { return len(passwords), len(columns) },
		func() fyne.CanvasObject {
			return container.NewMax(widget.NewLabel(""), widget.NewButton("", nil))
		},
		func(cell widget.TableCellID, o fyne.CanvasObject) {
			i, j := cell.Row, cell.Col
			label := o.(*fyne.Container).Objects[0].(*widget.Label)
			button := o.(*fyne.Container).Objects[1].(*widget.Button)

			label.Hide()
			button.Hide()

			switch j {
			case 0:
				label.SetText(strconv.Itoa(passwords[i].ID))
				label.Show()
			case 1:
				label.SetText(passwords[i].Service)
				label.Show()
			case 2:
				label.SetText(passwords[i].Username)
				label.Show()
			case 3:
				label.SetText(passwords[i].Category)
				label.Show()
			case 4:
				t, err := time.Parse(time.RFC3339, passwords[i].CreatedAt)
				if err != nil {
					label.SetText(passwords[i].CreatedAt)
				} else {
					label.SetText(t.Local().Format("02 January 2006, 15:04"))
				}
				label.Show()
			case 5:
				link := passwords[i].Link
				button.SetText(link)
				button.OnTapped = func() {
					utils.CopyToClipboard(link)
					statusLabel.SetText("Link copied to clipboard")
					clearStatusLater(statusLabel)
				}
				button.Show()
			case 6:
				button.SetText("Copy Password")
				button.OnTapped = func() {
					utils.CopyToClipboard(passwords[i].Password)
					statusLabel.SetText("Password copied to clipboard")
					clearStatusLater(statusLabel)
				}
				button.Show()
			}
		},
	)

	widths := []float32{30, 110, 110, 100, 190, 190, 140}
	for i, w := range widths {
		headerTable.SetColumnWidth(i, w)
		dataTable.SetColumnWidth(i, w)
	}

	scroll := container.NewScroll(dataTable)
	scroll.SetMinSize(fyne.NewSize(800, 300))

	statusBox := container.NewVBox(statusLabel)
	statusBox.Resize(fyne.NewSize(800, 30))

	content := container.NewBorder(headerTable, statusBox, nil, nil, scroll)

	return dataTable, content
}

func clearStatusLater(label *widget.Label) {
	go func() {
		time.Sleep(3 * time.Second)
		label.SetText("")
	}()
}

func ShowCreateForm(a fyne.App, appInstance *app.App) {
	w := a.NewWindow("Create Password")

	passwords, _ := appInstance.DB.GetAllPasswords()
	services, usernames, categories, links := extractSuggestions(passwords)

	// Поля ввода
	service := widget.NewSelectEntry(services)
	username := widget.NewSelectEntry(usernames)
	link := widget.NewSelectEntry(links)
	passwordEntry := widget.NewPasswordEntry()
	category := widget.NewSelectEntry(categories)
	statusLabel := widget.NewLabel("")

	// Настройки генерации
	lengthEntry := widget.NewEntry()
	lengthEntry.SetText("16")
	lengthEntry.SetPlaceHolder("Length")

	excludeEntry := widget.NewEntry()
	excludeEntry.SetPlaceHolder("Exclude chars")

	useUpper := widget.NewCheck("A-Z", nil)
	useUpper.SetChecked(true)
	useLower := widget.NewCheck("a-z", nil)
	useLower.SetChecked(true)
	useDigits := widget.NewCheck("0-9", nil)
	useDigits.SetChecked(true)
	useSymbols := widget.NewCheck("!@#", nil)
	useSymbols.SetChecked(true)

	// Кнопка генерации
	generateBtn := widget.NewButton("🔁 Generate", func() {
		length, err := strconv.Atoi(lengthEntry.Text)
		if err != nil || length <= 0 {
			statusLabel.SetText("Invalid length")
			clearStatusLater(statusLabel)
			return
		}
		password, err := utils.GeneratePassword(length, useUpper.Checked, useLower.Checked, useDigits.Checked, useSymbols.Checked, excludeEntry.Text)
		if err != nil {
			statusLabel.SetText("Generation error: " + err.Error())
			clearStatusLater(statusLabel)
			return
		}
		passwordEntry.SetText(password)
		statusLabel.SetText("Generated password inserted")
		clearStatusLater(statusLabel)
	})

	// Сетка генерации
	passwordRow := container.NewGridWithColumns(2,
		container.NewVBox(passwordEntry),
		container.NewVBox(generateBtn),
	)

	optionsGrid := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("Length"), lengthEntry),
		container.NewVBox(widget.NewLabel("Exclude"), excludeEntry),
	)

	checkboxGrid := container.NewGridWithColumns(4, useUpper, useLower, useDigits, useSymbols)

	passwordSection := container.NewVBox(
		passwordRow,
		optionsGrid,
		checkboxGrid,
	)

	// Финальная форма
	form := widget.NewForm(
		widget.NewFormItem("Service", service),
		widget.NewFormItem("Username", username),
		widget.NewFormItem("Link", link),
		widget.NewFormItem("Password", passwordSection),
		widget.NewFormItem("Category", category),
	)

	form.OnSubmit = func() {
		p := model.Password{
			Service:  service.Text,
			Username: username.Text,
			Link:     link.Text,
			Password: passwordEntry.Text,
			Category: category.Text,
		}
		if err := utils.ValidatePasswordStrength(p.Password, 60); err != nil {
			dialog.ShowError(err, w)
			return
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

	content := container.NewVBox(form, statusLabel)
	w.SetContent(container.NewPadded(content))
	w.Resize(fyne.NewSize(520, 420))
	w.Show()
}

func ShowFilterWindow(a fyne.App, appInstance *app.App) {
	w := a.NewWindow("Filter Passwords")

	passwords, _ := appInstance.DB.GetAllPasswords()
	services, usernames, categories, _ := extractSuggestions(passwords)

	service := widget.NewSelectEntry(services)
	username := widget.NewSelectEntry(usernames)
	category := widget.NewSelectEntry(categories)

	result := widget.NewLabel("")

	form := widget.NewForm(
		widget.NewFormItem("Service", service),
		widget.NewFormItem("Username", username),
		widget.NewFormItem("Category", category),
	)

	form.OnSubmit = func() {
		list, err := appInstance.DB.GetFilteredPasswords(service.Text, username.Text, category.Text)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if len(list) == 0 {
			result.SetText("No matching entries found.")
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
	w.Show()
}

func ShowUpdateWindow(a fyne.App, appInstance *app.App) {
	w := a.NewWindow("Update Password")

	passwords, _ := appInstance.DB.GetAllPasswords()
	services, usernames, categories, links := extractSuggestions(passwords)

	idEntry := widget.NewEntry()
	service := widget.NewSelectEntry(services)
	username := widget.NewSelectEntry(usernames)
	link := widget.NewSelectEntry(links)
	password := widget.NewPasswordEntry()
	category := widget.NewSelectEntry(categories)

	form := widget.NewForm(
		widget.NewFormItem("ID", idEntry),
		widget.NewFormItem("Service", service),
		widget.NewFormItem("Username", username),
		widget.NewFormItem("Link", link),
		widget.NewFormItem("Password", password),
		widget.NewFormItem("Category", category),
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
		dialog.ShowInformation("Updated", "Password updated successfully", w)
		w.Close()
	}

	w.SetContent(form)
	w.Resize(fyne.NewSize(400, 300))
	w.Show()
}

func ShowDeleteWindow(a fyne.App, appInstance *app.App) {
	w := a.NewWindow("Delete Password")

	idEntry := widget.NewEntry()
	idEntry.SetPlaceHolder("Enter ID to delete")

	deleteBtn := widget.NewButton("Delete", func() {
		id := idEntry.Text
		if err := appInstance.DB.DeletePassword(id); err != nil {
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation("Deleted", "Password deleted successfully", w)
		w.Close()
	})

	w.SetContent(container.NewVBox(
		idEntry,
		deleteBtn,
	))
	w.Resize(fyne.NewSize(300, 150))
	w.Show()
}
