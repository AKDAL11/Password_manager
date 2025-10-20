package gui

import (
    "image/color"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/canvas"
    "fyne.io/fyne/v2/widget"
)

// tapOverlay — прозрачный виджет, реагирующий на клики
type tapOverlay struct {
    widget.BaseWidget
    onTap func()
}

func newTapOverlay(onTap func()) *tapOverlay {
    t := &tapOverlay{onTap: onTap}
    t.ExtendBaseWidget(t)
    return t
}

func (t *tapOverlay) CreateRenderer() fyne.WidgetRenderer {
    // полностью прозрачный прямоугольник
    r := canvas.NewRectangle(color.Transparent)
    return widget.NewSimpleRenderer(r)
}

func (t *tapOverlay) Tapped(*fyne.PointEvent) {
    if t.onTap != nil {
        t.onTap()
    }
}
