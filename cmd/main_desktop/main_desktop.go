//go:build !android

package main

import (
    "log"

    _ "github.com/mattn/go-sqlite3"
    fyneapp "fyne.io/fyne/v2/app"

    "password-manager/internal/gui"
    "password-manager/internal/i18n"
)

func main() {
    // Загружаем локаль
    if err := i18n.LoadLocale("en"); err != nil {
        log.Fatal(err)
    }

    // Создаём приложение Fyne
    a := fyneapp.New()

    // Запускаем через экран разблокировки/создания мастер‑пароля
    gui.LaunchWithUnlock(a)

    // Запускаем цикл приложения
    a.Run()
}
