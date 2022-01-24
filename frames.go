package main

import (
	"log"
	// see ~/go/pkg/mod/github.com/gdamore/tcell/v2@v2.4.1-0.20210905002822-f057f0a857a1/
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
	s.SetStyle(tcell.StyleDefault) //.Background(tcell.ColorBlack).Background(tcell.GetColor("#000000")).Foreground(tcell.ColorWhite))
	return
}

func (this *Frames) Draw(s tcell.Screen, px, py, w, h int) (ret bool) {
	this.pos_vertical_bar = w / 3
	for x := 0; x < w; x++ {
		s.SetCell(x, 2, tcell.StyleDefault, tcell.RuneHLine)
	}
	s.SetCell(this.pos_vertical_bar, 2, tcell.StyleDefault, tcell.RuneTTee)
	for y := 3; y < h; y++ {
		s.SetCell(this.pos_vertical_bar, y, tcell.StyleDefault, tcell.RuneVLine)
	}
	return true
}

func (this *Frames) EventHandler(s tcell.Screen, event tcell.Event) (ret bool) {
	ret = false
	switch ev := event.(type) {
	case *tcell.EventResize:
		log.Printf("event=%v", ev)
		s.Sync()
		ret = true
	}
	return
}
