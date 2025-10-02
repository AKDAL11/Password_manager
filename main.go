// main.go
//go:build android

package main

import (
    fyneapp "fyne.io/fyne/v2/app"
    "password-manager/internal/gui"
)

func main() {
    a := fyneapp.New()
    gui.LaunchWithUnlock(a)
}
