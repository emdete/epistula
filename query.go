package main

import (
	"fmt"
	"log"
	"strings"
	// see ~/go/pkg/mod/github.com/gdamore/tcell/v2@v2.4.1-0.20210905002822-f057f0a857a1/
	"github.com/gdamore/tcell/v2"
)

type Query struct {
	Area
	pos_cur int
	query string
	pasting bool
}

func NewQuery(s tcell.Screen) (this Query) {
	log.Printf("NewQuery")
	this = Query{}
	this.pos_cur = len(QUERY_DEFAULT) + len(QUERY_SUFFIX)
	this.query = QUERY_DEFAULT + QUERY_SUFFIX
	this.pasting = false
	this.dirty = true
	this.notify(s)
	return
}

const (
	QUERY_DEFAULT = "tag:spam"
	//QUERY_DEFAULT = "tag:inbox"
	QUERY_PREFIX = "search "
	QUERY_SUFFIX = " AND "
)

func (this *Query) Draw(s tcell.Screen, px, py, w, h int) (ret bool) {
	this.SetString(s, px, py, tcell.StyleDefault, QUERY_PREFIX+this.query, w)
	// Cursor in query line
	s.ShowCursor(px+len(QUERY_PREFIX)+this.pos_cur, py)
	return true
}

func (this *Query) EventHandler(s tcell.Screen, event tcell.Event) {
	log.Printf("Query.EventHandler %v", event)
	switch ev := event.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			r := ev.Rune()
			this.query = fmt.Sprintf("%s%c%s", this.query[:this.pos_cur], r, this.query[this.pos_cur:])
			this.pos_cur++
			this.dirty = true
		case tcell.KeyLeft:
			if this.pos_cur > 0 {
				this.pos_cur--
				this.dirty = true
			}
		case tcell.KeyRight:
			if this.pos_cur < len(this.query) {
				this.pos_cur++
				this.dirty = true
			}
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if this.pos_cur > 0 {
				this.pos_cur--
				this.query = this.query[:this.pos_cur] + this.query[this.pos_cur+1:]
				this.dirty = true
			}
		case tcell.KeyDelete:
			if this.pos_cur > 0 {
				this.query = this.query[:this.pos_cur] + this.query[this.pos_cur+1:]
				this.dirty = true
			}
		case tcell.KeyHome:
			this.pos_cur = 0
			this.dirty = true
		case tcell.KeyEnd:
			this.pos_cur = len(this.query)
			this.dirty = true
		case tcell.KeyEnter:
			this.notify(s)
		case tcell.KeyTab:
			this.pos_cur = len(QUERY_DEFAULT) + len(QUERY_SUFFIX)
			this.query = QUERY_DEFAULT + QUERY_SUFFIX
			this.dirty = true
		}
	case *tcell.EventPaste:
		this.pasting = ev.Start()
	}
}

type EventQuery struct {
	tcell.EventTime
	query string
}

func (this *Query) notify(s tcell.Screen) {
	ev := &EventQuery{}
	ev.SetEventNow()
	if strings.HasSuffix(this.query, QUERY_SUFFIX) {
		// if the user did not change the QUERY_SUFFIX, she doesnt want that considered, cut it
		ev.query = this.query[:len(this.query)-len(QUERY_SUFFIX)]
	} else {
		ev.query = this.query
	}
	if err := s.PostEvent(ev); err != nil {
		panic(err)
	}
}

