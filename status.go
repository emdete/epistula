package main

import (
	"log"
	"fmt"
	// +tcell
	"github.com/gdamore/tcell/v2"
)

var TEMPLATE = "Filtered %d of %d overall mails"
type Status struct {
	line string
}

func NewStatus(s tcell.Screen) (ret Status) {
	log.Printf("NewStatus")
	ret = Status{
		fmt.Sprintf(TEMPLATE, 0, 0),
	}
	return
}

func (this *Status) Draw(s tcell.Screen, px, py, w, h int) (ret bool) {
	style := tcell.StyleDefault.Foreground(tcell.GetColor("#333333")).Background(tcell.GetColor("#ee9900")).Bold(true)
	emitStr(s, 0, 0, style, this.line, w)
	return true
}

func (this *Status) EventHandler(s tcell.Screen, event tcell.Event) (ret bool) {
	ret = false
	switch ev := event.(type) {
	case *EventThreadsStatus:
		this.line = fmt.Sprintf(TEMPLATE, ev.filtered, ev.overall)
		ret = true
	}
	return
}
