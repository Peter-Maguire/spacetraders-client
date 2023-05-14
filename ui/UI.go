package ui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var logView *tview.TextView
var shipView *tview.TextView

var app *tview.Application

func Init() {
	app = tview.NewApplication()

	logView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)

	logView.SetTextColor(tcell.ColorWhite).SetBorder(true).SetTitle("LogOutput")

	shipView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(false)

	shipView.SetBorder(true).SetTitle("Ships")

	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(logView, 0, 3, true), 0, 2, false).
		AddItem(shipView, 0, 1, false)

	app.SetRoot(flex, true)

	if err := app.Run(); err != nil {
		panic(err)
	}
}

var drawUpdateWaiting = false

func MainLog(str string) {
	if logView != nil {
		drawUpdateWaiting = true
		app.QueueUpdateDraw(func() {
			drawUpdateWaiting = false
			_, _ = fmt.Fprint(logView, str)
		})
	} else {
		fmt.Print(str)
	}
}

func WriteShipState(shipStates string) {
	if drawUpdateWaiting || shipView == nil {
		return
	}
	drawUpdateWaiting = true
	app.QueueUpdateDraw(func() {
		drawUpdateWaiting = false
		shipView.Clear()
		_, _ = fmt.Fprint(shipView, shipStates)
	})
}
