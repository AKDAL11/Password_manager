//go:build !android && !ios

package gui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/theme"
)

type DesktopFactory struct{}

func (DesktopFactory) WindowSize() fyne.Size        { return fyne.NewSize(1100, 700) }
func (DesktopFactory) SmallWindowSize() fyne.Size   { return fyne.NewSize(640, 420) }
func (DesktopFactory) SidebarRatio() float64        { return 0.25 }
func (DesktopFactory) SidebarWidth() float32        { return 300 }
func (DesktopFactory) TableColumnRatios() []float32 { return []float32{5, 15, 20, 10, 20, 15, 15} } // проценты
func (DesktopFactory) HeaderFontSize() int          { return 18 }
func (DesktopFactory) Theme() fyne.Theme            { return theme.DarkTheme() }

func CurrentFactory() UIFactory { return DesktopFactory{} }
