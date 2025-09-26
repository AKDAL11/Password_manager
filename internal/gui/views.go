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

    var split *container.Split
    var toggleBtn *widget.Button

    refreshBtn := widget.NewButton("🔄 Refresh", func() {
        newList, err := appInstance.DB.GetAllPasswords()
        if err != nil {
            dialog.ShowError(err, w)
            return
        }
        currentList = newList
        table.Length = func() (int, int) { return len(currentList), 6 }
        table.UpdateCell = func(cell widget.TableCellID, o fyne.CanvasObject) {
            if cell.Row >= len(currentList) || cell.Col >= 6 {
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
                link := currentList[i].Link
                button.SetText(link)
                button.OnTapped = func() {
                    utils.CopyToClipboard(link)
                    statusLabel.SetText("Link copied to clipboard")
                    clearStatusLater(statusLabel)
                }
                button.Show()
            case 5:
                t, err := time.Parse(time.RFC3339, currentList[i].CreatedAt)
                if err != nil {
                    label.SetText(currentList[i].CreatedAt)
                } else {
                    label.SetText(t.Format("02 Jan 2006, 15:04"))
                }
                label.Show()
            }
        }

        table.Refresh()
    })

    sidebarVisible := true
    toggleBtn = widget.NewButton("⬅️ Collapse", func() {
        sidebarVisible = !sidebarVisible
        if sidebarVisible {
            split.SetOffset(0.25)
            toggleBtn.SetText("⬅️ Collapse")
        } else {
            split.SetOffset(0.0)
            toggleBtn.SetText("➡️ Expand")
        }
    })

    sidebar := container.NewVBox(
        widget.NewLabel("📁 Actions"),
        widget.NewButton("➕ Add", func() { ShowCreateForm(a, appInstance) }),
        widget.NewButton("🎲 Generate", func() { ShowGeneratorWindow(a) }),
        widget.NewButton("🔍 Filter", func() { ShowFilterWindow(a, appInstance) }),
        widget.NewButton("✏️ Update", func() { ShowUpdateWindow(a, appInstance) }),
        widget.NewButton("❌ Delete", func() { ShowDeleteWindow(a, appInstance) }),
        widget.NewButton("📋 Copy", func() { ShowCopyWindow(a, appInstance) }),
        widget.NewButton("📁 View One", func() { ShowSinglePasswordWindow(a, appInstance) }),
        refreshBtn,
    )

    sidebarBox := container.NewVBox(toggleBtn, sidebar)
    sidebarBox.Resize(fyne.NewSize(200, 500))

    mainContent := container.NewBorder(
        widget.NewLabel("Your Passwords"),
        nil, nil, nil,
        tableContainer,
    )

    split = container.NewHSplit(sidebarBox, mainContent)
    split.Offset = 0.25

    w.SetContent(split)
    w.Resize(fyne.NewSize(1150, 600))
    w.CenterOnScreen()
    w.Show()
}

func buildPasswordTable(passwords []model.PasswordListItem, statusLabel *widget.Label) (*widget.Table, fyne.CanvasObject) {
    columns := []string{"ID", "Service", "Username", "Category", "Link", "Created At"}

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
                link := passwords[i].Link
                button.SetText(link)
                button.OnTapped = func() {
                    utils.CopyToClipboard(link)
                    statusLabel.SetText("Link copied to clipboard")
                    clearStatusLater(statusLabel)
                }
                button.Show()
            case 5:
                t, err := time.Parse(time.RFC3339, passwords[i].CreatedAt)
                if err != nil {
                    label.SetText(passwords[i].CreatedAt)
                } else {
                    label.SetText(t.Format("02 Jan 2006, 15:04"))
                }
                label.Show()
            }
        },
    )

    widths := []float32{60, 140, 140, 120, 200, 160}
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

	service := widget.NewSelectEntry(services)
	username := widget.NewSelectEntry(usernames)
	link := widget.NewSelectEntry(links)
	password := widget.NewPasswordEntry()
	category := widget.NewSelectEntry(categories)

	form := widget.NewForm(
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

	w.SetContent(form)
	w.Resize(fyne.NewSize(400, 300))
	w.Show()
}

func ShowGeneratorWindow(a fyne.App) {
	w := a.NewWindow("Password Generator")

	lengthEntry := widget.NewEntry()
	lengthEntry.SetPlaceHolder("Length")

	excludeEntry := widget.NewEntry()
	excludeEntry.SetPlaceHolder("Exclude characters")

	upper := widget.NewCheck("Uppercase", nil)
	lower := widget.NewCheck("Lowercase", nil)
	digits := widget.NewCheck("Digits", nil)
	symbols := widget.NewCheck("Symbols", nil)

	result := widget.NewLabel("")

	generateBtn := widget.NewButton("Generate", func() {
		length, _ := strconv.Atoi(lengthEntry.Text)
		pass, err := utils.GeneratePassword(length, upper.Checked, lower.Checked, digits.Checked, symbols.Checked, excludeEntry.Text)
		if err != nil {
			result.SetText("Error: " + err.Error())
		} else {
			result.SetText("Generated: " + pass)
		}
	})

	w.SetContent(container.NewVBox(
		lengthEntry, excludeEntry,
		upper, lower, digits, symbols,
		generateBtn, result,
	))
	w.Resize(fyne.NewSize(400, 300))
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

func ShowCopyWindow(a fyne.App, appInstance *app.App) {
	w := a.NewWindow("Copy Password")

	idEntry := widget.NewEntry()
	idEntry.SetPlaceHolder("Enter ID to copy")

	copyBtn := widget.NewButton("Copy", func() {
		id := idEntry.Text
		encrypted, err := appInstance.DB.GetEncryptedPassword(id)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		password, err := appInstance.Crypto.Decrypt(encrypted)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if err := utils.CopyToClipboard(password); err != nil {
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation("Copied", "Password copied to clipboard", w)
	})

	w.SetContent(container.NewVBox(
		idEntry,
		copyBtn,
	))
	w.Resize(fyne.NewSize(300, 150))
	w.Show()
}

func ShowSinglePasswordWindow(a fyne.App, appInstance *app.App) {
	w := a.NewWindow("View Password by ID")

	idEntry := widget.NewEntry()
	idEntry.SetPlaceHolder("Enter ID")

	result := widget.NewLabel("")

	viewBtn := widget.NewButton("View", func() {
		id := idEntry.Text
		p, err := appInstance.DB.GetPasswordByID(id)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		output := "Service: " + p.Service + "\n" +
			"Username: " + p.Username + "\n" +
			"Link: " + p.Link + "\n" +
			"Category: " + p.Category + "\n" +
			"Created: " + p.CreatedAt
		result.SetText(output)
	})

	w.SetContent(container.NewVBox(
		idEntry,
		viewBtn,
		result,
	))
	w.Resize(fyne.NewSize(400, 250))
	w.Show()
}
