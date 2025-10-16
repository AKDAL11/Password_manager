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

    // Запускаем окно разблокировки
    gui.LaunchWithUnlock(a)

    // Запускаем главный цикл событий
    a.Run()
}
