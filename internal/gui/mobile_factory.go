//go:build android || ios

package gui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/theme"
)

type MobileFactory struct{}

func (MobileFactory) WindowSize() fyne.Size        { return fyne.NewSize(400, 700) }
func (MobileFactory) SmallWindowSize() fyne.Size   { return fyne.NewSize(360, 640) }
func (MobileFactory) SidebarRatio() float64        { return 0.35 }
func (MobileFactory) SidebarWidth() float32        { return 180 }
func (MobileFactory) TableColumnRatios() []float32 { return []float32{10, 20, 20, 10, 15, 15, 10} } // проценты
func (MobileFactory) HeaderFontSize() int          { return 14 }
func (MobileFactory) Theme() fyne.Theme            { return theme.DarkTheme() }

func CurrentFactory() UIFactory { return MobileFactory{} }
