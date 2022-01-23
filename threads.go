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
	selected_index int
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
	selected_style := tcell.StyleDefault.Foreground(tcell.GetColor("#333333")).Background(tcell.GetColor("#ee9900"))
	if this.selected_index >= h {
		this.selected_index = h-1
	}
	for i, thread := range this.mailthreads {
		log.Printf("%d: %v", i, thread)
		var cs1, cs2 tcell.Style
		if i == this.selected_index {
			cs1 = selected_style.Bold(true)
			cs2 = cs1.Bold(false)
		} else {
			cs1 = tcell.StyleDefault.Bold(true)
			cs2 = cs1.Bold(false).Foreground(tcell.GetColor("#999999"))
		}
		if thread != nil {
			emitStr(s, px, py+i*2, cs1, thread.author, w)
			emitStr(s, px, py+i*2+1, cs2, thread.subject, w)
			//emitStr(s, px, py+i*3+2, tcell.StyleDefault, "", w)
		} else {
			emitStr(s, px, py+i*2, tcell.StyleDefault, "", w)
			emitStr(s, px, py+i*2+1, tcell.StyleDefault, "", w)
			//emitStr(s, px, py+i*3+2, tcell.StyleDefault, "", w)
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
			if this.selected_index > 0 {
				this.selected_index--
				ret = true
			}
		case tcell.KeyDown:
			if this.selected_index+1 < this.filtered_t {
				this.selected_index++
				ret = true
			}
		}
	case *tcell.EventMouse:
		x, y := ev.Position()
		button := ev.Buttons()
		if button != tcell.ButtonNone && x < frames.pos_vertical_bar && y >= 3 {
			this.selected_index = y - 3 // TODO offset
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
	count := 0
	if threads, err := nmq.Threads(); err == nil {
		this.threads = threads
		var thread *notmuch.Thread
		for this.threads.Next(&thread) {
			defer thread.Close()
			log.Printf("%d: %v", count, thread.Subject())
			this.mailthreads[count] = newThreadEntry(thread)
			if count >= this.filtered_t { // assertion
				panic("more threads than reported")
			}
			count++
			if count >= len(this.mailthreads) {
				break
			}
		}
		if count < this.filtered_t && count != len(this.mailthreads) { // assertion
			panic("less threads than reported")
		}
		if this.selected_index < 0 {
			// after empty results move the selection into the result
			this.selected_index = 0
		}
	}
	for count<len(this.mailthreads) {
		this.mailthreads[count] = nil
		count++
	}
	if this.selected_index >= this.filtered_t {
		this.selected_index = this.filtered_t-1
		// -1 being valid for empty results
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
	author := "-"
	matched, unmatched := thread.Authors()
	if len(matched) > 0 {
		author = matched[0]
	} else if len(unmatched) > 0 {
		author = unmatched[0]
	}
	// log.Printf("matched=%v, unmatched=%v", matched, unmatched)
	this := ThreadEntry {
		thread.ID(),
		"🙂 " + author,
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

