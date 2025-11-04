//go:build android

package main

import (
    "fyne.io/fyne/v2/app"
    "password-manager/internal/gui"
    "password-manager/internal/i18n"
)

func main() {
    a := app.New()

    // Загружаем язык по умолчанию
    _ = i18n.LoadLocale("en")

    
    gui.LaunchWithUnlock(a)

    // Главный цикл событий
    a.Run()
}