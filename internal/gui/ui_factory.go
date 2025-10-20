package gui

import "fyne.io/fyne/v2"

type UIFactory interface {
    WindowSize() fyne.Size
    SmallWindowSize() fyne.Size
    SidebarRatio() float64
    TableColumnRatios() []float32
    HeaderFontSize() int
    Theme() fyne.Theme
    SidebarWidth() float32
}
