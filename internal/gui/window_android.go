//go:build android

package gui

import "fyne.io/fyne/v2"

func configureWindow(w fyne.Window) {
    w.SetMaster()
    w.SetFullScreen(true)
    w.SetTitle("") // пустой заголовок → не будет панели с крестиком
}
