package main

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
	// see ~/go/pkg/mod/github.com/gdamore/tcell/v2@v2.4.1-0.20210905002822-f057f0a857a1/
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
	"github.com/mattn/go-runewidth"
)

var frames Frames
var status Status
var query Query
var threads Threads
var mails Mails
var NotMuchDatabasePath string

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string, width int) {
	for _, c := range str {
		if width > 0 {
			var comb []rune
			w := runewidth.RuneWidth(c)
			if w == 0 {
				comb = []rune{c}
				c = ' '
				w = 1
			}
			s.SetContent(x, y, c, comb, style)
			x += w
			width -= w
		} else {
			break
		}
	}
	for width > 0 {
		s.SetContent(x, y, ' ', nil, style)
		x++
		width--
	}
}

func updateScreen(s tcell.Screen) {
	w, h := s.Size()
	s.Clear()
	frames.Draw(s, 0, 0, w, h)
	status.Draw(s, 0, 0, w, 1)
	query.Draw(s, 0, 1, w, 1)
	threads.Draw(s, 0, 3, frames.pos_vertical_bar, h-3)
	mails.Draw(s, frames.pos_vertical_bar+1, 3, w-frames.pos_vertical_bar-1, h-3)
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
			NotMuchDatabasePath = filepath.Join(usr.HomeDir, "Maildir")
		}
		if err := s.Init(); err != nil {
			panic(err)
		}
		s.EnableMouse()
		s.EnablePaste()
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
		update := true
		for running {
			if update {
				updateScreen(s)
				update = false
			}
			event := s.PollEvent()
			switch ev := event.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyRune, tcell.KeyLeft, tcell.KeyRight,
					tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyEnter,
					tcell.KeyDelete, tcell.KeyHome, tcell.KeyEnd, tcell.KeyTab:
					update = update || query.EventHandler(s, event)
				case tcell.KeyUp, tcell.KeyDown:
					update = update || threads.EventHandler(s, event)
				case tcell.KeyPgUp, tcell.KeyPgDn:
					update = update || mails.EventHandler(s, event)
				case tcell.KeyEscape:
					running = false
				case tcell.KeyCtrlB:
					s.Beep()
				}
			case *tcell.EventResize:
				update = update || frames.EventHandler(s, event)
			case *tcell.EventPaste:
				update = update || query.EventHandler(s, event)
			case *tcell.EventMouse:
				update = update ||
					threads.EventHandler(s, event) ||
					mails.EventHandler(s, event) ||
					query.EventHandler(s, event)
			case *EventQuery: // query input reports new querystring -> threads
				update = update || threads.EventHandler(s, event)
			case *EventThreadsStatus: // threads report new thread list / stats -> status
				update = update || status.EventHandler(s, event)
			case *EventThreadsThread: // threads report new selected thread -> threads
				update = update || mails.EventHandler(s, event)
			}
		}
	}
}

// event telling to refresh the threads cause mails arrived / db changed
type EventMainRefresh struct {
	tcell.EventTime
	query string
}

func notify(s tcell.Screen) {
	ev := &EventMainRefresh{}
	ev.SetEventNow()
	if err := s.PostEvent(ev); err != nil {
		panic(err)
	}
}

type Area struct {
	px, py, dx, dy int
}

