package main

import (
	_ "log"
	// +tcell
	"github.com/gdamore/tcell/v2"
)


var PREFIX = "search "
var SUFFIX = " AND "
// Query
// keys: Left Right chars enter tab
type Query struct {
	pos_cur int
	query string
}

type EventQuery struct {
	tcell.EventTime
	query string
}

func NewQuery(s tcell.Screen) (this Query) {
	this = Query{
		38,
		"tag:inbox AND NOT tag:spam",
	}
	this.notify(s)
	return
}

func (this *Query) Draw(s tcell.Screen, px, py, w, h int) (ret bool) {
	emitStr(s, px, py, tcell.StyleDefault, PREFIX + this.query + SUFFIX, w)
	// Cursor in query line
	s.ShowCursor(px+this.pos_cur, py)
	return true
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
			if this.pos_cur > len(PREFIX) {
				this.pos_cur--
			}
			ret = true
		case tcell.KeyRight:
			if this.pos_cur < len(PREFIX) + len(this.query) + len(SUFFIX) {
				this.pos_cur++
			}
			ret = true
		case tcell.KeyEnter:
			this.notify(s)
		}
		case *tcell.EventPaste:
			_ = ev.Start()
	}
	return
}
