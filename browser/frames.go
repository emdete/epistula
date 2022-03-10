package main

import (
	"log"
	// see ~/go/pkg/mod/github.com/gdamore/tcell/v2@v2.4.1-0.20210905002822-f057f0a857a1/
	"github.com/gdamore/tcell/v2"
)

type Frames struct {
	Area
	pos_vertical_bar int
	style tcell.Style
}

func NewFrames(s tcell.Screen, pos_vertical_bar int) (this Frames) {
	log.Printf("NewFrames")
	this = Frames{}
	this.pos_vertical_bar = pos_vertical_bar
	//s.SetStyle(tcell.StyleDefault)
	this.style = tcell.StyleDefault.Foreground(tcell.GetColor("#cc7711"))
	return
}

func (this *Frames) Draw(s tcell.Screen) (ret bool) {
	this.pos_vertical_bar = this.dx / 3
	for x := 0; x < this.dx; x++ {
		this.SetContent(s, x, 2, tcell.RuneHLine, nil, this.style)
	}
	this.SetContent(s, this.pos_vertical_bar, 2, tcell.RuneTTee, nil, this.style)
	for y := 3; y < this.dy; y++ {
		this.SetContent(s, this.pos_vertical_bar, y, tcell.RuneVLine, nil, this.style)
	}
	return true
}

func (this *Frames) EventHandler(s tcell.Screen, event tcell.Event) {
	// log.Printf("Frames.EventHandler %v", event)
	switch event.(type) {
	case *tcell.EventResize:
		s.Sync()
		this.dirty = true
	}
}
