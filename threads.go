package main

import (
	"log"
	"os/user"
	"path/filepath"
	// +tcell
	"github.com/gdamore/tcell/v2"
	// +notmuch
	"github.com/zenhack/go.notmuch"
)
// List
// KeyUp KeyDown EventQuery
type Enumeration struct {
	db *notmuch.DB
	subjects []string
	pos_mail int
}

func NewEnumeration(s tcell.Screen) (this Enumeration) {
	this = Enumeration{
	}
	if user, err := user.Current(); err != nil {
		panic(err)
	} else if db, err := notmuch.Open(filepath.Join(user.HomeDir, "Maildir"), notmuch.DBReadOnly); err != nil {
			panic(err)
	} else {
		this.db = db
	}
	return
}

func (this *Enumeration) Close() {
	if this.db != nil {
		this.db.Close()
	}
}

func (this *Enumeration) Draw(s tcell.Screen, px, py, w, h int) (ret bool) {
	style := tcell.StyleDefault.Foreground(tcell.GetColor("#333333")).Background(tcell.GetColor("#ee9900")).Bold(true)
	if this.pos_mail >= h {
		this.pos_mail = h-1
	}
	y := 0
	for _,value := range this.subjects {
		cs := tcell.StyleDefault
		if y == this.pos_mail {
			cs = style
		}
		emitStr(s, px, py+y, cs, value, frames.pos_vertical_bar)
		y++
	}
	return true
}

func (this *Enumeration) EventHandler(s tcell.Screen, event tcell.Event) (ret bool) {
	ret = false
	switch ev := event.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyUp:
			if this.pos_mail > 0 {
				this.pos_mail--
				ret = true
			}
		case tcell.KeyDown:
			this.pos_mail++
			ret = true
		}
	case *tcell.EventMouse:
		x, y := ev.Position()
		button := ev.Buttons()
		if button != tcell.ButtonNone && x < frames.pos_vertical_bar && y >= 3 {
			this.pos_mail = y - 3 // TODO offset
			updateScreen(s)
		}
	case *EventQuery:
		this.do_query()
	}
	return
}

type EventThreadsMail struct {
	tcell.EventTime
}

func (this *Enumeration) notifyMail(s tcell.Screen) {
	ev := &EventThreadsMail{}
	ev.SetEventNow()
	if err := s.PostEvent(ev); err != nil {
		panic(err)
	}
}

type EventThreadsStatus struct {
	tcell.EventTime
	overall int
	filtered int
}

func (this *Enumeration) notifyStatus(s tcell.Screen) {
	ev := &EventThreadsStatus{}
	ev.SetEventNow()
	if err := s.PostEvent(ev); err != nil {
		panic(err)
	}
}

func (this *Enumeration) do_query() {
	// see ~/go/pkg/mod/github.com/zenhack/go.notmuch@v0.0.0-20211022191430-4d57e8ad2a8b/
	{
		{
			// this.db.FindMessage(id)
			// if _, err := this.db.Tags(); err != nil {
			//
			if threads, err := this.db.NewQuery("tag:inbox AND NOT tag:spam").Threads(); err != nil {
				panic(err)
			} else {
				defer threads.Close()
				var thread *notmuch.Thread
				for threads.Next(&thread) {
					defer thread.Close()
					//this.subjects[thread.ID()] = thread.Subject()
					this.subjects = append(this.subjects, thread.Subject())
					// matched, unmatched := thread.Authors()
					// thread.Count()
					// thread.CountMatched()
					// thread.OldestDate()
					// thread.NewestDate()
					tags := thread.Tags()
					var tag *notmuch.Tag
					for tags.Next(&tag) {
						log.Printf("tag=%s", tag)
					}
					// thread.Messages()
					messages := thread.Messages()
					var message *notmuch.Message
					for messages.Next(&message) {
						defer message.Close()
						log.Printf("filename=%s", message.Filename())
					}
				}
			}
		}
	}
}

