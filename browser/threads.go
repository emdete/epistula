package main

import (
	"log"
	"fmt"
	"errors"
	"time"
	// see ~/go/pkg/mod/github.com/gdamore/tcell/v2@v2.4.1-0.20210905002822-f057f0a857a1/
	"github.com/gdamore/tcell/v2"
	// see ~/go/pkg/mod/github.com/zenhack/go.notmuch@v0.0.0-20211022191430-4d57e8ad2a8b/
	"github.com/zenhack/go.notmuch"
)

type Threads struct {
	Area // where do we write
	last_query string // what was the last query (to not repeat it unnesesarily)
	filtered_t int // how many threads where found by notmuch
	threadEntries [100](*ThreadEntry) // list of found threads // TODO add dynamic+pageing
	offset int // offset into the list, whats on top
	selected_index int // which thread is selected
}

func NewThreads(s tcell.Screen) (this Threads) {
	log.Printf("NewThreads")
	this = Threads{}
	return
}

const (
	THREADS_PAGE = 13
	THREADS_ATTACHMENT = 0x1F4CE
)

func (this *Threads) Close() {
}

func (this *Threads) Draw(s tcell.Screen) (ret bool) {
	this.ClearArea(s)
	selected_style := tcell.StyleDefault.Foreground(tcell.GetColor("#333333")).Background(tcell.GetColor("#cc7711"))
	for i, threadEntry := range this.threadEntries[this.offset:] {
		//log.Printf("%d: %v", i, threadEntry)
		var cs1, cs2 tcell.Style
		if this.offset+i == this.selected_index {
			cs1 = selected_style.Bold(true)
			cs2 = cs1.Bold(false)
		} else {
			cs1 = tcell.StyleDefault.Bold(true)
			cs2 = cs1.Bold(false).Foreground(tcell.GetColor("#999999"))
		}
		if threadEntry != nil {
			dx := this.dx
			if threadEntry.unread {
				cs3 := tcell.StyleDefault.Reverse(true)
				dx--
				this.SetContent(s, dx, i*2, ' ', nil, cs3)
				this.SetContent(s, dx, i*2+1, ' ', nil, cs3)
			}
			this.SetString(s, 0, i*2, cs1, "🙂 " + threadEntry.author, dx)
			this.SetString(s, 0, i*2+1, cs2, threadEntry.subject, dx)
			var ts string
			if threadEntry.attachment {
				ts = fmt.Sprintf("%c [%d]", THREADS_ATTACHMENT, threadEntry.count)
				dx = this.dx - len(ts)
			} else {
				ts = fmt.Sprintf("[%d]", threadEntry.count)
				dx = this.dx - len(ts) - 2
			}
			this.SetString(s, dx, i*2, cs1, ts, this.dx - dx - 1)
		} else {
			this.SetString(s, 0, i*2, tcell.StyleDefault, "", this.dx-1)
			this.SetString(s, 0, i*2+1, tcell.StyleDefault, "", this.dx-1)
		}
		i++
	}
	return true
}

func (this *Threads) doDown(down bool) bool {
	if down {
		if this.selected_index+1 < this.filtered_t {
			this.selected_index++
			if this.selected_index-this.offset > this.dy/2-2 {
				this.offset++
			}
		}
	} else {
		if this.selected_index > 0 {
			this.selected_index--
			if this.selected_index <= this.offset && this.offset > 0 {
				this.offset--
			}
		}
	}
	return true
}

func (this *Threads) EventHandler(s tcell.Screen, event tcell.Event) {
	// log.Printf("Threads.EventHandler %v", event)
	old_index := this.selected_index
	switch ev := event.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlA:
			ThreadRemoveTag(this.threadEntries[this.selected_index].id, "inbox")
		case tcell.KeyCtrlS:
			ThreadRemoveTag(this.threadEntries[this.selected_index].id, "inbox")
			ThreadAddTag(this.threadEntries[this.selected_index].id, "spam")
		case tcell.KeyDown:
			if ev.Modifiers()&tcell.ModCtrl != 0 {
				for i:=0;i<THREADS_PAGE;i++ {
					this.dirty = this.doDown(true)
				}

			} else {
				this.dirty = this.doDown(true)
			}
		case tcell.KeyUp:
			if ev.Modifiers()&tcell.ModCtrl != 0 {
				for i:=0;i<THREADS_PAGE;i++ {
					this.dirty = this.doDown(false)
				}
			} else {
				this.dirty = this.doDown(false)
			}
		}
	case *tcell.EventMouse:
		button := ev.Buttons()
		switch button {
		case tcell.Button1:
			x, y := ev.Position()
			x -= this.px
			y -= this.py
			this.selected_index = y/2 + this.offset
			this.dirty = true
		case tcell.WheelUp:
			this.dirty = this.doDown(false)
		case tcell.WheelDown:
			this.dirty = this.doDown(true)
		}
	case *EventQuery:
		this.do_query(s, ev.query, ev.refresh)
		old_index = -1
		this.dirty = true
	}
	if old_index != this.selected_index {
		if old_index >= 0 {
			ThreadRemoveTag(this.threadEntries[old_index].id, "unread")
		}
		this.notifyThreadsThread(s)
	}
}

func (this *Threads) do_query(s tcell.Screen, query string, refresh bool) {
	log.Printf("Threads.do_query %v", query)
	if this.last_query != query || refresh {
		this.last_query = query
		if db, err := notmuch.Open(NotMuchDatabasePath, notmuch.DBReadOnly); err != nil {
			panic(err)
		} else {
			defer db.Close()
			// db.Tags()
			// db.FindMessage(id)
			nmq := db.NewQuery(query)
			defer nmq.Close()
			nmq.SetSortScheme(notmuch.SORT_NEWEST_FIRST)
			nmq.SetExcludeScheme(notmuch.EXCLUDE_ALL)
			nmq.AddTagExclude("spam")
			this.filtered_t = nmq.CountThreads()
			log.Printf("filtered_t=%d", this.filtered_t)
			filtered_m := nmq.CountMessages()
			count := 0
			if threads, err := nmq.Threads(); err == nil {
				defer threads.Close()
				threads = threads
				var thread *notmuch.Thread
				for threads.Next(&thread) {
					defer thread.Close()
					// log.Printf("%d: %v", count, thread.Subject())
					this.threadEntries[count] = newThreadEntry(thread)
					if count >= this.filtered_t { // assertion
						panic("more threads than reported")
					}
					count++
					if count >= len(this.threadEntries) {
						this.filtered_t = len(this.threadEntries)
						break
					}
				}
				if count < this.filtered_t && count != len(this.threadEntries) { // assertion
					panic("less threads than reported")
				}
			}
			for count < len(this.threadEntries) {
				this.threadEntries[count] = nil
				count++
			}
			if this.filtered_t <= 0 {
				this.selected_index = -1 // -1 being valid for empty results
			} else {
				this.selected_index = 0
			}
			this.offset = 0
			st := db.NewQuery("*")
			overall_t := 0 // too expensive: st.CountThreads()
			overall_m := st.CountMessages()
			this.notifyThreadsStatus(s, overall_t, overall_m, this.filtered_t, filtered_m)
			if this.filtered_t > 0 {
				this.notifyThreadsThread(s)
			}
		}
	}
}

type ThreadEntry struct {
	id string
	author string
	subject string
	unread, attachment bool
	count int
	newest time.Time
}

func newThreadEntry(thread *notmuch.Thread) *ThreadEntry {
	author := "-"
	matched, unmatched := thread.Authors()
	if len(matched) > 0 {
		author = matched[0]
	} else if len(unmatched) > 0 {
		author = unmatched[0]
	}
	// log.Printf("matched=%v, unmatched=%v", matched, unmatched)
	this := ThreadEntry{
		thread.ID(),
		author,
		thread.Subject(),
		ThreadHasTag(thread, "unread"),
		ThreadHasTag(thread, "attachment"),
		thread.Count(),
		thread.NewestDate(),
	}
	return &this
}

type EventThreadsThread struct {
	tcell.EventTime
	ThreadEntry
}

func (this *Threads) notifyThreadsThread(s tcell.Screen) {
	ev := &EventThreadsThread{}
	ev.SetEventNow()
	if this.selected_index >= 0 {
		ev.ThreadEntry = *this.threadEntries[this.selected_index]
	}
	if err := s.PostEvent(ev); err != nil {
		panic(err)
	}
	//x(this.db, ev.id)
}

type EventThreadsStatus struct {
	tcell.EventTime
	overall_t int
	overall_m int
	filtered_m int
	filtered_t int
}

func (this *Threads) notifyThreadsStatus(s tcell.Screen, overall_t, overall_m, filtered_t, filtered_m int) {
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

func ThreadHasTag(thread *notmuch.Thread, search string) bool {
	tags := thread.Tags()
	var tag *notmuch.Tag
	for tags.Next(&tag) {
		if tag.Value == search {
			return true
		}
	}
	return false
}

func ThreadAddTag(id, tag string) error {
	if db, err := notmuch.Open(NotMuchDatabasePath, notmuch.DBReadWrite); err != nil {
		return err
	} else {
		defer db.Close()
		query := db.NewQuery("thread:" + id)
		defer query.Close()
		if 1 != query.CountThreads() { return errors.New("not uniq") }
		if threads, err := query.Threads(); err != nil {
			return err
		} else {
			var thread *notmuch.Thread
			if threads.Next(&thread) {
				defer thread.Close()
			}
			messages := thread.Messages()
			var message *notmuch.Message
			for messages.Next(&message) {
				if err := message.AddTag(tag); err != nil { return err }
			}
			if threads.Next(&thread) { return errors.New("additional thread") }
		}
	}
	return nil
}


func ThreadRemoveTag(id, tag string) error {
	if db, err := notmuch.Open(NotMuchDatabasePath, notmuch.DBReadWrite); err != nil {
		return err
	} else {
		defer db.Close()
		query := db.NewQuery("thread:" + id)
		defer query.Close()
		if 1 != query.CountThreads() { return errors.New("not uniq") }
		if threads, err := query.Threads(); err != nil {
			return err
		} else {
			var thread *notmuch.Thread
			if threads.Next(&thread) {
				defer thread.Close()
			}
			messages := thread.Messages()
			var message *notmuch.Message
			for messages.Next(&message) {
				if err := message.RemoveTag(tag); err != nil { return err }
			}
			if threads.Next(&thread) { return errors.New("additional thread") }
		}
	}
	return nil
}

