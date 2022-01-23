package main

import (
	"log"
	"os/user"
	"path/filepath"
	"github.com/gdamore/tcell/v2"
	"github.com/zenhack/go.notmuch"
)

type Enumeration struct {
	// see ~/go/pkg/mod/github.com/zenhack/go.notmuch@v0.0.0-20211022191430-4d57e8ad2a8b/
	db *notmuch.DB
	threads *notmuch.Threads
	filtered_t int
	mailthreads [100](*ThreadEntry)
	pos_mail int
}

func NewEnumeration(s tcell.Screen) (this Enumeration) {
	log.Printf("NewEnumeration")
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
	if this.threads != nil {
		this.threads.Close()
	}
	if this.db != nil {
		this.db.Close()
	}
}

func (this *Enumeration) Draw(s tcell.Screen, px, py, w, h int) (ret bool) {
	style := tcell.StyleDefault.Foreground(tcell.GetColor("#333333")).Background(tcell.GetColor("#ee9900")).Bold(true)
	if this.pos_mail >= h {
		this.pos_mail = h-1
	}
	for i, thread := range this.mailthreads {
		log.Printf("%d: %v", i, thread)
		var cs tcell.Style
		if i == this.pos_mail {
			cs = style
		} else {
			cs = tcell.StyleDefault
		}
		if thread != nil {
			emitStr(s, px, py+i, cs, thread.subject, frames.pos_vertical_bar)
			log.Printf("%d: s=%v", i, thread.subject)
		} else {
			emitStr(s, px, py+i, cs, "           ", frames.pos_vertical_bar)
		}
		i++
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
			if this.pos_mail+1 < this.filtered_t {
				this.pos_mail++
				ret = true
			}
		}
	case *tcell.EventMouse:
		x, y := ev.Position()
		button := ev.Buttons()
		if button != tcell.ButtonNone && x < frames.pos_vertical_bar && y >= 3 {
			this.pos_mail = y - 3 // TODO offset
		}
		ret = true
	case *EventQuery:
		this.do_query(s, ev.query)
	}
	return
}

func (this *Enumeration) do_query(s tcell.Screen, query string) {
	// this.db.FindMessage(id)
	// this.db.Tags()
	if this.threads != nil {
		this.threads.Close()
	}
	nmq := this.db.NewQuery(query)
	nmq.SetSortScheme(notmuch.SORT_NEWEST_FIRST)
	nmq.SetExcludeScheme(notmuch.EXCLUDE_ALL)
	nmq.AddTagExclude("spam")
	this.filtered_t = nmq.CountThreads()
	log.Printf("filtered_t=%d", this.filtered_t)
	filtered_m := nmq.CountMessages()
	if threads, err := nmq.Threads(); err != nil {
		panic(err)
	} else {
		this.threads = threads
	}
	count := 0
	var thread *notmuch.Thread
	for this.threads.Next(&thread) {
		defer thread.Close()
		log.Printf("%d: %v", count, thread.Subject())
		this.mailthreads[count] = newThreadEntry(thread)
		if count >= this.filtered_t {
			panic("more threads than reported")
		}
		count++
		if count > 50 {
			break
		}
	}
	if count < this.filtered_t {
		panic("less threads than reported")
	}
	for count<len(this.mailthreads) {
		this.mailthreads[count] = nil
		count++
	}
	if this.pos_mail >= this.filtered_t {
		this.pos_mail = this.filtered_t-1
	}
	st := this.db.NewQuery("*")
	overall_t := 0 // too expensive: st.CountThreads()
	overall_m := st.CountMessages()
	this.notifyStatus(s, overall_t, overall_m, this.filtered_t, filtered_m)
	this.notifyMail(s)
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
	overall_t int
	overall_m int
	filtered_m int
	filtered_t int
}

func (this *Enumeration) notifyStatus(s tcell.Screen, overall_t, overall_m, filtered_t, filtered_m int) {
	ev := &EventThreadsStatus{}
	ev.SetEventNow()
	ev.overall_t = overall_t
	ev.overall_m = overall_m
	ev.filtered_m = filtered_m
	ev.filtered_t = filtered_t
	if err := s.PostEvent(ev); err != nil {
		panic(err)
	}
}

type ThreadEntry struct {
	id string
	author string
	subject string
}

func newThreadEntry(thread *notmuch.Thread) (*ThreadEntry) {
	// matched, unmatched := thread.Authors()
	// log.Printf("matched=%v, unmatched=%v", matched, unmatched)
	this := ThreadEntry {
		thread.ID(),
		"",//matched[0],
		thread.Subject(),
	}
	// thread.Count()
	// thread.CountMatched()
	// thread.OldestDate()
	// thread.NewestDate()
	// tags := thread.Tags()
	// var tag *notmuch.Tag
	// for tags.Next(&tag) {
	// log.Printf("tag=%s", tag)
	// }
	// thread.Messages()
	// messages := thread.Messages()
	// var message *notmuch.Message
	// for messages.Next(&message) {
	// defer message.Close()
	// log.Printf("filename=%s", message.Filename())
	// }
	return &this
}

