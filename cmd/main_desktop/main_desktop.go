//go:build !android

package main

import (
    fyneapp "fyne.io/fyne/v2/app"
    "password-manager/internal/gui"
    "password-manager/internal/i18n"
)

func main() {
    // Загружаем язык при старте
    if err := i18n.LoadLocale("en"); err != nil {
        panic(err)
    }

    a := fyneapp.New()
    gui.LaunchWithUnlock(a) // передаём реальный fyne.App
    a.Run()
}
