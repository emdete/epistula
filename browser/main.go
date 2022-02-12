package main

import (
	"log"
	"os"
	"fmt"
	"os/signal"
	"syscall"
	"strings"
	"os/user"
	"path/filepath"
	// see ~/go/pkg/mod/github.com/gdamore/tcell/v2@v2.4.1-0.20210905002822-f057f0a857a1/
	"github.com/gdamore/tcell/v2"
	//
	"github.com/gdamore/tcell/v2/encoding"
	// see ~/go/pkg/mod/github.com/mattn/go-runewidth@v0.0.14-0.20210830053702-dc8fe66265af/
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
var from string

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
	log.SetPrefix("epistula ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC | log.Lshortfile)
	if f, err := os.OpenFile("/tmp/epistula-browser.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		log.Fatal(err)
	} else {
		os.Stderr = f
	}
	log.SetOutput(os.Stderr)
}

func main() {
	config := NewConfig()
	// terminal title
	title := "Epistula Browser: " + config.user_name + " <" + config.user_primary_email + ">"
	os.Stdout.Write([]byte("\x1b]1;"+title+"\a\x1b]2;"+title+"\a"))
	// log
	_log()
	log.Printf("main")
	// search the required executables
	if err := CheckEnv(); err != nil {
		panic(err)
	}
	//
	for i:=1;i<len(os.Args);i++ {
		if strings.HasPrefix(os.Args[i], "--") {
			x := strings.Split(os.Args[i][2:], "=")
			switch x[0] {
			case "from":
				from = x[1]
			default:
				log.Fatal(fmt.Sprintf("wrong arg: %s", os.Args[i]))
			}
		} else {
			log.Fatal(fmt.Sprintf("wrong arg: %s", os.Args[i]))
		}
	}
	// tcell
	encoding.Register()
	if s, err := tcell.NewScreen(); err != nil {
		panic(err)
	} else {
		defer s.Fini()
		Signal(s)
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
			if !s.HasPendingEvent() {
				updateScreen(s)
			}
			event := s.PollEvent()
			switch ev := event.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyRune, // all input goes to the query editing
						tcell.KeyLeft,
						tcell.KeyRight,
						tcell.KeyBackspace,
						tcell.KeyBackspace2,
						tcell.KeyEnter,
						tcell.KeyDelete,
						tcell.KeyHome,
						tcell.KeyEnd,
						tcell.KeyTab:
					query.EventHandler(s, ev)
				case tcell.KeyUp, // navigate through the result set
						tcell.KeyDown:
					threads.EventHandler(s, ev)
				case tcell.KeyCtrlA, // archive thread
						tcell.KeyCtrlS: // mark thread as spam
					threads.EventHandler(s, ev) // we have to do this separate cause we need to
					query.notify(s, true) // notify via query which isnt available in threads
				case tcell.KeyPgUp, // page through the selected thread in view
						tcell.KeyPgDn, //
						tcell.KeyCtrlO, // open more lines of the selected mail
						tcell.KeyCtrlJ, // next mail in thread
						tcell.KeyCtrlK, // previous mail in thread
						tcell.KeyCtrlR, // reply to selected email
						tcell.KeyCtrlD, // save selected part of selected email
						tcell.KeyCtrlC: // compose new email
					mails.EventHandler(s, ev)
				case tcell.KeyEscape: // terminate epistula
					running = false
				case tcell.KeyCtrlB: // test alert
					s.Beep()
				case tcell.KeyCtrlL: // refresh screen
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
			case *EventQuery: // query input reports new querystring -> threads
				threads.EventHandler(s, event)
			case *EventThreadsStatus: // threads report new thread list / stats -> status
				status.EventHandler(s, event)
			case *EventThreadsThread: // threads report new selected thread -> threads
				mails.EventHandler(s, event)
			case *EventSignal: // handle signals
				switch ev.sig {
				case syscall.SIGHUP: // reread config
					config = NewConfig()
				case syscall.SIGUSR1: // refresh result set
					query.notify(s, true)
				case syscall.SIGUSR2: //
					//
				}
			}
		}
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
	if x >= 0 && x < this.dx && y >= 0 && y < this.dy {
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
	px := 0
	for _, word := range strings.Split(paragraph, " ") {
		if len(word) > 0 { // remove double space
			if px + len(word) > width {
				this.SetString(s, x+px, y, style, "", width-px) // wipe rest of line
				y++
				px = 0
			}
			if px == 0 { // put prefix
				px += this.SetString(s, x+px, y, style, paragraphprefix, len(paragraphprefix))
			}
			// if len(word) > width // TODO
			px += this.SetString(s, x+px, y, style, word, len(word)+1)
		}
	}
	this.SetString(s, x+px, y, style, "", width+px)
	y++
	return x, y
}

type EventSignal struct {
	tcell.EventTime
	sig os.Signal
}

func Signal(s tcell.Screen) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		for true {
			sig := <-sigs
			log.Printf("signal %v", sig)
			ev := &EventSignal{}
			ev.SetEventNow()
			ev.sig = sig
			if err := s.PostEvent(ev); err != nil {
				panic(err)
			}
		}
	}()
}

