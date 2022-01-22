package main

import (
	"log"
	"os"
	// +tcell
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
	"github.com/mattn/go-runewidth"
)

var pos_vertical_bar = 60
var status Status
var queryEdit QueryEdit
var frames Frames

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
			width--
		}
	}
	for width > 0 {
		var comb []rune
		s.SetContent(x, y, ' ', comb, style)
		x += 1
		width--
	}
}

func updateScreen(s tcell.Screen) {
	w, h := s.Size()
	s.Clear()
	frames.Draw(s, 0, 0, w, h)
	status.Draw(s, 0, 0, w, 1)
	queryEdit.Draw(s, 0, 1, w, 1)
	// TODO:
	style := tcell.StyleDefault.Foreground(tcell.GetColor("#333333")).Background(tcell.GetColor("#ee9900")).Bold(true)
	if pos_mail_y <= 3 {
		pos_mail_y = 3
	} else if pos_mail_y >= h {
		pos_mail_y = h-1
	}
	y := 3
	for _,value := range subjects {
		cs := tcell.StyleDefault
		if y == pos_mail_y {
			cs = style
		}
		emitStr(s, 0, y, cs, value, pos_vertical_bar)
		y++
	}
	s.Show()
	//s.Sync()
}

func _log() {
	log.SetPrefix("email ")
	log.SetFlags(log.Ldate|log.Lmicroseconds|log.LUTC|log.Lshortfile)
	log.SetOutput(os.Stderr)
}

func main() {
	// log
	_log()
	// Status
	status = NewStatus()
	// Frames
	frames = NewFrames(pos_vertical_bar)
	// Query
	queryEdit = NewQueryEdit()
	// notmuch
	_notmuch()
	// gpgme
	_gpgme()
	// gmime3
	_gmime3()
	// tcell
	// see ~/go/pkg/mod/github.com/gdamore/tcell/v2@v2.4.1-0.20210905002822-f057f0a857a1/
	encoding.Register()
	if s, err := tcell.NewScreen(); err != nil {
		panic(err)
	} else {
		if err := s.Init(); err != nil {
			panic(err)
		}
		s.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
		s.EnableMouse()
		s.EnablePaste()
		updateScreen(s)
		for {
			switch ev := s.PollEvent().(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				// queryedit //
				case tcell.KeyRune, tcell.KeyLeft, tcell.KeyRight,
				tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyEnter,
				tcell.KeyDelete, tcell.KeyHome, tcell.KeyEnd, tcell.KeyTab:
					if queryEdit.Feed(s, ev) {
						updateScreen(s)
					}
				// enumeration //
				case tcell.KeyUp, tcell.KeyDown:
					if EnumerationFeed(s, ev) {
						updateScreen(s)
					}
					// threaddisplay
				case tcell.KeyPgUp, tcell.KeyPgDn:
					if ThreadDisplayFeed(s, ev) {
						updateScreen(s)
					}
					//
				case tcell.KeyEscape:
					s.Fini()
					os.Exit(0)
				case tcell.KeyCtrlB:
					s.Beep()
					s.Sync()
				}
			case *tcell.EventResize:
				s.Sync()
				updateScreen(s)
			case *tcell.EventPaste:
				_ = ev.Start()
			case *tcell.EventMouse:
				x, y := ev.Position()
				button := ev.Buttons()
				if button != tcell.ButtonNone && x < pos_vertical_bar && y >= 3 {
					pos_mail_y = y
					updateScreen(s)
				}
			}
		}
	}
}
