package main

import (
	_ "log"
	// +tcell
	"github.com/gdamore/tcell/v2"
)

type Status struct {
}

func NewStatus(s tcell.Screen) (ret Status) {
	ret = Status{
	}
	return
}

func (this *Status) Draw(s tcell.Screen, px, py, w, h int) (ret bool) {
	style := tcell.StyleDefault.Foreground(tcell.GetColor("#333333")).Background(tcell.GetColor("#ee9900")).Bold(true)
	emitStr(s, 0, 0, style, "Mails 12 unread, 24 in inbox of 310032 overall", w)
	return true
}
