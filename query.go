package main

import (
	"log"
	"fmt"
	"strings"
	// see ~/go/pkg/mod/github.com/gdamore/tcell/v2@v2.4.1-0.20210905002822-f057f0a857a1/
	"github.com/gdamore/tcell/v2"
)


var QUERY_DEFAULT = "tag:inbox"
var QUERY_PREFIX = "search "
var QUERY_SUFFIX = " AND "
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
		len(QUERY_DEFAULT) + len(QUERY_SUFFIX),
		QUERY_DEFAULT + QUERY_SUFFIX,
		false,
	}
	this.notify(s)
	return
}

func (this *Query) Draw(s tcell.Screen, px, py, w, h int) (ret bool) {
	emitStr(s, px, py, tcell.StyleDefault, QUERY_PREFIX + this.query, w)
	// Cursor in query line
	s.ShowCursor(px+len(QUERY_PREFIX)+this.pos_cur, py)
	return true
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

func (this *Query) EventHandler(s tcell.Screen, event tcell.Event) (ret bool) {
	ret = false
	switch ev := event.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			r := ev.Rune()
			this.query = fmt.Sprintf("%s%c%s", this.query[:this.pos_cur], r, this.query[this.pos_cur:])
			this.pos_cur++
			ret = true
		case tcell.KeyLeft:
			if this.pos_cur > 0 {
				this.pos_cur--
				ret = true
			}
		case tcell.KeyRight:
			if this.pos_cur < len(this.query) {
				this.pos_cur++
				ret = true
			}
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if this.pos_cur > 0 {
				this.pos_cur--
				this.query = this.query[:this.pos_cur] + this.query[this.pos_cur+1:]
				ret = true
			}
		case tcell.KeyDelete:
			if this.pos_cur > 0 {
				this.query = this.query[:this.pos_cur] + this.query[this.pos_cur+1:]
				ret = true
			}
		case tcell.KeyHome:
			this.pos_cur = 0
			ret = true
		case tcell.KeyEnd:
			this.pos_cur = len(this.query)
			ret = true
		case tcell.KeyTab:
			// autocompletion
		case tcell.KeyEnter:
			this.notify(s)
		}
		case *tcell.EventPaste:
			this.pasting = ev.Start()
	}
	return
}
