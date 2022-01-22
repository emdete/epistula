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
// keys: Up Down
type Enumeration struct {
}
var pos_mail_y = 3
//var subjects = map[string]string{}
var subjects []string

func EnumerationFeed(s tcell.Screen, ev *tcell.EventKey) (ret bool) {
	ret = false
	switch ev.Key() {
	case tcell.KeyUp:
		pos_mail_y--
		ret = true
	case tcell.KeyDown:
		pos_mail_y++
		ret = true
	}
	return
}

func _notmuch() {
	// see ~/go/pkg/mod/github.com/zenhack/go.notmuch@v0.0.0-20211022191430-4d57e8ad2a8b/
	if user, err := user.Current(); err != nil {
		panic(err)
	} else {
		if db, err := notmuch.Open(filepath.Join(user.HomeDir, "Maildir"), notmuch.DBReadOnly); err != nil {
			panic(err)
		} else {
			defer db.Close()
			// db.FindMessage(id)
			// if _, err := db.Tags(); err != nil {
			//
			if threads, err := db.NewQuery("tag:inbox AND NOT tag:spam").Threads(); err != nil {
				panic(err)
			} else {
				defer threads.Close()
				var thread *notmuch.Thread
				for threads.Next(&thread) {
					defer thread.Close()
					//subjects[thread.ID()] = thread.Subject()
					subjects = append(subjects, thread.Subject())
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

