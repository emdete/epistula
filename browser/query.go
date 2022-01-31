package main

import (
	"log"
	"strings"
	// see ~/go/pkg/mod/github.com/gdamore/tcell/v2@v2.4.1-0.20210905002822-f057f0a857a1/
	"github.com/gdamore/tcell/v2"
)

type Query struct {
	Area
	pos_cur int
	query []rune
	pasting bool
}

func NewQuery(s tcell.Screen) (this Query) {
	log.Printf("NewQuery")
	this = Query{}
	this.query = append(QUERY_DEFAULT, QUERY_SUFFIX...)
	this.pos_cur = len(this.query)
	this.pasting = false
	this.dirty = true
	this.notify(s, false)
	return
}

var (
	QUERY_PREFIX = []rune("search ")
	QUERY_DEFAULT = []rune("tag:inbox")
	QUERY_SUFFIX = []rune(" AND ")
)

func (this *Query) Draw(s tcell.Screen) (ret bool) {
	this.SetString(s, 0, 0, tcell.StyleDefault, string(append(QUERY_PREFIX, this.query...)), this.dx)
	// Cursor always in query line
	s.ShowCursor(len(QUERY_PREFIX)+this.pos_cur, this.py)
	return true
}

func (this *Query) EventHandler(s tcell.Screen, event tcell.Event) {
	log.Printf("Query.EventHandler %v", event)
	switch ev := event.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			if this.pos_cur < len(this.query) {
				this.query = append(this.query, ' ') // create empty space
				copy(this.query[this.pos_cur+1:], this.query[this.pos_cur:]) // copy suffix to the right
				this.query[this.pos_cur] = ev.Rune() // set the rune
			} else {
				this.query = append(this.query, ev.Rune()) // add rune to the end
			}
			this.pos_cur++ // increment cursor position
			this.dirty = true
		case tcell.KeyLeft:
			if ev.Modifiers()&tcell.ModCtrl != 0 {
				this.pos_cur = 0
			} else {
				if this.pos_cur > 0 {
					this.pos_cur--
				}
			}
			this.dirty = true
		case tcell.KeyRight:
			if ev.Modifiers()&tcell.ModCtrl != 0 {
				this.pos_cur = len(this.query) - 1
			} else {
				if this.pos_cur < len(this.query) {
					this.pos_cur++
				}
			}
			this.dirty = true
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if this.pos_cur > 0 {
				this.pos_cur--
				suff := this.query[this.pos_cur+1:]
				this.query = append(this.query[:this.pos_cur], suff...)
				this.dirty = true
			}
		case tcell.KeyDelete:
			if this.pos_cur < len(this.query) {
				suff := this.query[this.pos_cur+1:]
				this.query = append(this.query[:this.pos_cur], suff...)
				this.dirty = true
			}
		case tcell.KeyHome:
			this.pos_cur = 0
			this.dirty = true
		case tcell.KeyEnd:
			this.pos_cur = len(this.query)
			this.dirty = true
		case tcell.KeyEnter:
			this.notify(s, false)
		case tcell.KeyTab:
			this.query = append(QUERY_DEFAULT, QUERY_SUFFIX...)
			this.pos_cur = len(this.query)
			this.dirty = true
		}
	case *tcell.EventMouse:
		button := ev.Buttons()
		switch button {
		case tcell.Button1:
			x, _ := ev.Position()
			if x >= this.px + len(QUERY_PREFIX) {
				x -= this.px + len(QUERY_PREFIX)
				if x > len(this.query) {
					x = len(this.query)
				}
				this.pos_cur = x
				this.dirty = true
			}
		case tcell.WheelUp:
			if this.pos_cur > 0 {
				this.pos_cur--
				this.dirty = true
			}
		case tcell.WheelDown:
			if this.pos_cur < len(this.query) {
				this.pos_cur++
				this.dirty = true
			}
		}
	case *tcell.EventPaste:
		this.pasting = ev.Start()
	}
}

type EventQuery struct {
	tcell.EventTime
	query string
	refresh bool
}

func (this *Query) notify(s tcell.Screen, refresh bool) {
	ev := &EventQuery{}
	ev.SetEventNow()
	if strings.HasSuffix(string(this.query), string(QUERY_SUFFIX)) {
		// if the user did not change the QUERY_SUFFIX, she doesnt want that considered, cut it
		ev.query = string(this.query[:len(this.query)-len(QUERY_SUFFIX)])
	} else {
		ev.query = string(this.query)
	}
	ev.refresh = refresh
	if err := s.PostEvent(ev); err != nil {
		panic(err)
	}
}

