package main

import (
	_ "log"
	// +tcell
	"github.com/gdamore/tcell/v2"
)


// Query
// keys: Left Right chars enter tab
type Query struct {
	pos_cur int
	query string
}

func NewQuery() (ret Query) {
	ret = Query{
		38,
		"tag:inbox AND NOT tag:spam",
	}
	return
}

func (this *Query) Draw(s tcell.Screen, px, py, w, h int) (ret bool) {
	emitStr(s, px, py, tcell.StyleDefault, "search " + this.query + " AND ", w)
	// Cursor in query line
	s.ShowCursor(px+this.pos_cur, py)
	return true
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
			}
			ret = true
		case tcell.KeyRight:
			if this.pos_cur < 38 {
				this.pos_cur++
			}
			ret = true
		}
		case *tcell.EventPaste:
			_ = ev.Start()
	}
	return
}
