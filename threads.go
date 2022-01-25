package main

import (
	"log"
	"time"
	// see ~/go/pkg/mod/github.com/gdamore/tcell/v2@v2.4.1-0.20210905002822-f057f0a857a1/
	"github.com/gdamore/tcell/v2"
	// see ~/go/pkg/mod/github.com/zenhack/go.notmuch@v0.0.0-20211022191430-4d57e8ad2a8b/
	"github.com/zenhack/go.notmuch"
)

type Threads struct {
	Area
	last_query string
	filtered_t int
	threadEntries [100](*ThreadEntry)
	offset int
	area_height int
	area_width int
	area_px int
	area_py int
	selected_index int
}

func NewThreads(s tcell.Screen) (this Threads) {
	log.Printf("NewThreads")
	this = Threads{}
	return
}

func (this *Threads) Close() {
}

func (this *Threads) Draw(s tcell.Screen, px, py, w, h int) (ret bool) {
	this.area_height = h
	this.area_width = w
	this.area_px = px
	this.area_py = py
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
			emitStr(s, px, py+i*2, cs1, "🙂 " + threadEntry.author, w)
			emitStr(s, px, py+i*2+1, cs2, threadEntry.subject, w)
		} else {
			emitStr(s, px, py+i*2, tcell.StyleDefault, "", w)
			emitStr(s, px, py+i*2+1, tcell.StyleDefault, "", w)
		}
		i++
	}
	return true
}

func (this *Threads) doDown(down bool) bool {
	if down {
		if this.selected_index+1 < this.filtered_t {
			this.selected_index++
			if this.selected_index-this.offset > this.area_height/2-2 {
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
	log.Printf("Threads.EventHandler %v", event)
	old_index := this.selected_index
	switch ev := event.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyDown:
			if ev.Modifiers()&tcell.ModCtrl != 0 {
				for i:=0;i<7;i++ {
					this.dirty = this.doDown(true)
				}

			} else {
				this.dirty = this.doDown(true)
			}
		case tcell.KeyUp:
			if ev.Modifiers()&tcell.ModCtrl != 0 {
				this.selected_index = 0
				this.offset = 0
				this.dirty = true
			} else {
				this.dirty = this.doDown(false)
			}
		}
	case *tcell.EventMouse:
		button := ev.Buttons()
		x, y := ev.Position()
		if x >= this.area_px && y >= this.area_py {
			x -= this.area_px
			y -= this.area_py
			if x < this.area_width && y < this.area_height {
				switch button {
				case tcell.Button1:
					this.selected_index = y/2 + this.offset
					this.dirty = true
				case tcell.WheelUp:
					this.dirty = this.doDown(false)
				case tcell.WheelDown:
					this.dirty = this.doDown(true)
				}
			}
		}
	case *EventQuery:
		this.do_query(s, ev.query)
		this.dirty = true
	}
	if old_index != this.selected_index && this.selected_index >= 0 {
		this.notifyThreadsThread(s)
	}
}

func (this *Threads) do_query(s tcell.Screen, query string) {
	if this.last_query != query {
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
					log.Printf("%d: %v", count, thread.Subject())
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
		thread.Count(),
		thread.NewestDate(),
	}
	// thread.CountMatched()
	// thread.OldestDate()
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

type EventThreadsThread struct {
	tcell.EventTime
	ThreadEntry
}

func (this *Threads) notifyThreadsThread(s tcell.Screen) {
	ev := &EventThreadsThread{}
	ev.SetEventNow()
	ev.ThreadEntry = *this.threadEntries[this.selected_index]
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

func MessageHasTag(message *notmuch.Message, search string) bool {
	tags := message.Tags()
	var tag *notmuch.Tag
	for tags.Next(&tag) {
		if tag.Value == search {
			return true
		}
	}
	return false
}

