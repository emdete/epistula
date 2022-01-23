package main

import (
	"log"
	// +tcell
	"github.com/gdamore/tcell/v2"
)

type Frames struct {
	pos_vertical_bar int
}

func NewFrames(s tcell.Screen, pos_vertical_bar int) (ret Frames) {
	log.Printf("NewFrames")
	ret = Frames{
		pos_vertical_bar,
	}
	return
}

func (this *Frames) Draw(s tcell.Screen, px, py, w, h int) (ret bool) {
	for x:=0;x<w;x++ {
		s.SetContent(x, 2, tcell.RuneHLine, nil, tcell.StyleDefault)
	}
	s.SetContent(this.pos_vertical_bar, 2, tcell.RuneTTee, nil, tcell.StyleDefault)
	for y:=3;y<h;y++ {
		s.SetContent(this.pos_vertical_bar, y, tcell.RuneVLine, nil, tcell.StyleDefault)
	}
	return true
}

func (this *Frames) EventHandler(s tcell.Screen, event tcell.Event) (ret bool) {
	ret = false
	return
}

