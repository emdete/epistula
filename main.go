package main

import (
	"log"
	"os"
	"strings"
	"os/user"
	"path/filepath"
	// see ~/go/pkg/mod/github.com/gdamore/tcell/v2@v2.4.1-0.20210905002822-f057f0a857a1/
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
	"github.com/mattn/go-runewidth"
)

const (
	MAIN_MAILDIR = "Maildir"
)

var frames Frames
var status Status
var query Query
var threads Threads
var mails Mails
var NotMuchDatabasePath string

func updateScreen(s tcell.Screen) {
	w, h := s.Size()
	if frames.dirty { // if frames.dirty a complete redraw is requested
		frames.SetSize(0, 0, w, h)
		frames.Draw(s)
		frames.dirty = false
	}
	if frames.dirty || status.dirty {
		status.SetSize(0, 0, w, 1)
		status.Draw(s)
		status.dirty = false
	}
	if frames.dirty || query.dirty {
		query.SetSize(0, 1, w, 1)
		query.Draw(s)
		query.dirty = false
	}
	if frames.dirty || threads.dirty {
		threads.SetSize(0, 3, frames.pos_vertical_bar, h-3)
		threads.Draw(s)
		threads.dirty = false
	}
	if frames.dirty || mails.dirty {
		mails.SetSize(frames.pos_vertical_bar+1, 3, w-frames.pos_vertical_bar-1, h-3)
		mails.Draw(s)
		mails.dirty = false
	}
	s.Show()
}

func _log() {
	log.SetPrefix("email ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC | log.Lshortfile)
	log.SetOutput(os.Stderr)
}

func main() {
	// log
	_log()
	log.Printf("main")
	// tcell
	encoding.Register()
	if s, err := tcell.NewScreen(); err != nil {
		panic(err)
	} else {
		defer s.Fini()
		if usr, err := user.Current(); err != nil {
			panic(err)
		} else {
			NotMuchDatabasePath = filepath.Join(usr.HomeDir, MAIN_MAILDIR)
		}
		if err := s.Init(); err != nil {
			panic(err)
		}
		s.EnableMouse()
		s.EnablePaste()
		s.Clear()
		// Frames
		frames = NewFrames(s, 61)
		// Status
		status = NewStatus(s)
		// Query
		query = NewQuery(s)
		// Threads
		threads = NewThreads(s)
		// Mails
		mails = NewMails(s)
		//
		running := true
		for running {
			updateScreen(s)
			event := s.PollEvent()
			switch ev := event.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyRune, tcell.KeyLeft, tcell.KeyRight,
					tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyEnter,
					tcell.KeyDelete, tcell.KeyHome, tcell.KeyEnd, tcell.KeyTab:
					query.EventHandler(s, event)
				case tcell.KeyUp, tcell.KeyDown:
					threads.EventHandler(s, event)
				case tcell.KeyPgUp, tcell.KeyPgDn:
					mails.EventHandler(s, event)
				case tcell.KeyEscape:
					running = false
				case tcell.KeyCtrlR:
					notifyRefresh(s)
				case tcell.KeyCtrlB:
					s.Beep()
				case tcell.KeyCtrlL:
					s.Sync()
				}
			case *tcell.EventMouse:
				if threads.IsEventIn(ev) {
					threads.EventHandler(s, event)
				} else if mails.IsEventIn(ev) {
					mails.EventHandler(s, event)
				} else if query.IsEventIn(ev) {
					query.EventHandler(s, event)
				}
			case *tcell.EventPaste:
				query.EventHandler(s, event)
			case *tcell.EventResize:
				frames.EventHandler(s, event)
			case *EventMainRefresh: // request to re-issue query
				threads.EventHandler(s, event)
			case *EventQuery: // query input reports new querystring -> threads
				threads.EventHandler(s, event)
			case *EventThreadsStatus: // threads report new thread list / stats -> status
				status.EventHandler(s, event)
			case *EventThreadsThread: // threads report new selected thread -> threads
				mails.EventHandler(s, event)
			}
		}
	}
}

// event telling to refresh the threads cause mails arrived / db changed
type EventMainRefresh struct {
	tcell.EventTime
	query string
}

func notifyRefresh(s tcell.Screen) {
	ev := &EventMainRefresh{}
	ev.SetEventNow()
	if err := s.PostEvent(ev); err != nil {
		panic(err)
	}
}

type Area struct {
	px, py, // position on screen
	dx, dy int // size on screen
	dirty bool
}

func (this *Area) SetSize(px, py, dx, dy int) {
	this.px = px
	this.py = py
	this.dx = dx
	this.dy = dy
}

func (this *Area) IsEventIn(ev *tcell.EventMouse) bool {
	x, y := ev.Position()
	return this.px <= x && x < this.px + this.dx && this.py <= y && y < this.py + this.dy
}

func (this *Area) ClearArea(s tcell.Screen) {
	for x:=this.px;x<this.px+this.dx;x++ {
		for y:=this.py;y<this.py+this.dy;y++ {
			s.SetContent(x, y, ' ', nil, tcell.StyleDefault)
		}
	}
}

func (this *Area) SetContent(s tcell.Screen, x int, y int, mainc rune, combc []rune, style tcell.Style) {
	if x < this.dx && y < this.dy {
		s.SetContent(x+this.px, y+this.py, mainc, combc, style)
	} else {
		//log.Printf("SetContent off screen %#v x=%d y=%d %c", this, x, y, mainc)
	}
}

func (this *Area) SetString(s tcell.Screen, x, y int, style tcell.Style, str string, width int) int {
	px := 0
	for _, c := range str {
		if c < ' ' {
			c = ' '
		}
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		this.SetContent(s, x+px, y, c, comb, style)
		px += w
		width -= w
		if width <= 0 {
			break
		}
	}
	for width > 0 {
		this.SetContent(s, x+px, y, ' ', nil, style)
		px++
		width--
	}
	return px
}

func (this *Area) SetParagraph(s tcell.Screen, x, y int, style tcell.Style, paragraphprefix, paragraph string, width int) (int, int) {
	if true {
		px := 0
		for _, word := range strings.Split(paragraph, " ") {
			if len(word) > 0 { // remove double space
				if px + len(word) > width {
					this.SetString(s, x+px, y, style, "", width-px)
					y++
					px = 0
				}
				if px == 0 {
					px += this.SetString(s, x+px, y, style, paragraphprefix, len(paragraphprefix))
				}
				// if len(word) > width // TODO
				px += this.SetString(s, x+px, y, style, word, len(word)+1)
			}
		}
		this.SetString(s, x+px, y, style, "", width+px)
		y++
	} else {
		this.SetString(s, x, y, style, paragraph, width)
		y++
	}
	log.Printf("SetParagraph y=%d '%s'", y, paragraph)
	return x, y
}

