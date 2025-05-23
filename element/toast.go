package element

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"time"
)

/*

  File:    toast.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

*/
/*
  Description: Provide a simple temporary message / notification.
*/

func InfoColor() color.NRGBA { return color.NRGBA{B: 0xff} }

func GOColor() color.NRGBA { return color.NRGBA{G: 0xff} }

func FailColor() color.NRGBA { return color.NRGBA{R: 0xff} }

func Toast(can fyne.Canvas, pos fyne.Position, txt string, c color.NRGBA, d time.Duration) {
	// light gray
	rect := canvas.NewRectangle(color.NRGBA{R: 0xd3, G: 0xd3, B: 0xd3, A: 0xcc})
	lr := c
	lr.A = 0x11
	dr := c
	dr.A = 0xff
	t := canvas.NewText(txt, lr)
	t.TextStyle.Bold = true
	a := canvas.NewColorRGBAAnimation(lr, dr,
		d, func(c color.Color) {
			t.Color = c
			canvas.Refresh(t)
		})
	a.RepeatCount = 0
	a.AutoReverse = true
	// size in pixels of an icon
	h := theme.IconInlineSize()
	w := t.MinSize().Width + h/2
	rect.Resize(fyne.NewSize(w, h))
	go func() {
		pos.X -= w + 10
		pos.Y -= h + 10
		pop := widget.NewPopUp(container.NewWithoutLayout(rect, t), can)
		pop.ShowAtPosition(pos)
		a.Start()
		pop.Show()
		time.Sleep(2 * d)
		pop.Hide()
	}()
}
