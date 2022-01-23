package main

import (
	"log"
	// +tcell
	"github.com/gdamore/tcell/v2"
)


var DEFAULT = "tag:inbox AND NOT tag:spam"
var PREFIX = "search "
var SUFFIX = " AND "
// Query
// keys: Left Right chars enter tab
type Query struct {
	pos_cur int
	query string
	pasting bool
}

func NewQuery(s tcell.Screen) (this Query) {
	log.Printf("NewQuery")
	this = Query{
		len(DEFAULT) + len(SUFFIX),
		DEFAULT,
		false,
	}
	this.notify(s)
	return
}

func (this *Query) Draw(s tcell.Screen, px, py, w, h int) (ret bool) {
	emitStr(s, px, py, tcell.StyleDefault, PREFIX + this.query + SUFFIX, w)
	// Cursor in query line
	s.ShowCursor(px+len(PREFIX)+this.pos_cur, py)
	return true
}

type EventQuery struct {
	tcell.EventTime
	query string
}

func (this *Query) notify(s tcell.Screen) {
	ev := &EventQuery{}
	ev.SetEventNow()
	ev.query = this.query
	if err := s.PostEvent(ev); err != nil {
		panic(err)
	}
}

func (this *Query) EventHandler(s tcell.Screen, event tcell.Event) (ret bool) {
	ret = false
	switch ev := event.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'b':
			}
		case tcell.KeyLeft:
			if this.pos_cur > 0 {
				this.pos_cur--
				ret = true
			}
		case tcell.KeyRight:
			if this.pos_cur < len(this.query) + len(SUFFIX) {
				this.pos_cur++
				ret = true
			}
		case tcell.KeyEnter:
			this.notify(s)
		}
		case *tcell.EventPaste:
			this.pasting = ev.Start()
	}
	return
}
