package main

import (
	"fmt"
	"log"
	// see ~/go/pkg/mod/github.com/gdamore/tcell/v2@v2.4.1-0.20210905002822-f057f0a857a1/
	"github.com/gdamore/tcell/v2"
)

const (
	STATUS_TEMPLATE = "Filtered %d messages from %d threads out of a total of %d messages"
)

type Status struct {
	line string
}

func NewStatus(s tcell.Screen) (ret Status) {
	log.Printf("NewStatus")
	ret = Status{
		fmt.Sprintf(STATUS_TEMPLATE, 0, 0, 0),
	}
	return
}

func (this *Status) Draw(s tcell.Screen, px, py, w, h int) (ret bool) {
	style := tcell.StyleDefault.Foreground(tcell.GetColor("#333333")).Background(tcell.GetColor("#cc7711")).Bold(true)
	emitStr(s, 0, 0, style, this.line, w)
	return true
}

func (this *Status) EventHandler(s tcell.Screen, event tcell.Event) (ret bool) {
	ret = false
	log.Printf("Status.EventHandler %v", event)
	switch ev := event.(type) {
	case *EventThreadsStatus:
		this.line = fmt.Sprintf(STATUS_TEMPLATE, ev.filtered_m, ev.filtered_t, ev.overall_m)
		ret = true
	}
	return
}
