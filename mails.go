package main

import (
	"log"
	"os"
	// see ~/go/pkg/mod/github.com/gdamore/tcell/v2@v2.4.1-0.20210905002822-f057f0a857a1/
	"github.com/gdamore/tcell/v2"
	// see ~/go/pkg/mod/github.com/proglottis/gpgme@v0.1.1
	"github.com/proglottis/gpgme"
	// see ~/go/pkg/mod/github.com/sendgrid/go-gmime@v0.0.0-20211124164648-4c44cbd981d8/
	_ "github.com/sendgrid/go-gmime"
	// see ~/go/pkg/mod/github.com/zenhack/go.notmuch@v0.0.0-20211022191430-4d57e8ad2a8b/
	"github.com/zenhack/go.notmuch"
)

type Mails struct {
	ThreadEntry
}

func NewMails(s tcell.Screen) (this Mails) {
	log.Printf("NewMails")
	this = Mails{}
	// gpgme
	this._gpgme()
	// gmime3
	this._gmime3()
	return
}

func (this *Mails) Draw(s tcell.Screen, px, py, w, h int) (ret bool) {
	// RuneULCorner = '┌' // RuneTTee  = '┬' // RuneURCorner = '┐'
	// RuneLTee     = '├' // RuneHLine = '─' // RuneRTee     = '┤'
	// RuneLLCorner = '└' // RuneBTee  = '┴' // RuneLRCorner = '┘'
	// RuneVLine =    '│' 
	cs1 := tcell.StyleDefault.Foreground(tcell.GetColor("#11aa11"))
	//selected_style := tcell.StyleDefault.Foreground(tcell).Background(tcell.GetColor("#ee9900"))
	cs := cs1.Reverse(true)
	emitStr(s, px, py, cs, " " + this.subject, w)
	for row := 1; row < 24; row++ {
		s.SetCell(px, py+row, cs1, tcell.RuneVLine)
	}
	emitStr(s, px+1, py+1, cs1.Bold(true), "From: " + this.author, w)
	//emitStr(s, px+1, py+1, cs1, this.newest, w)
	s.SetCell(px, py+24, cs1, tcell.RuneLLCorner)
	emitStr(s, px+1, py+24, cs, " " + this.subject, w)
	return true
}

func (this *Mails) EventHandler(s tcell.Screen, event tcell.Event) (ret bool) {
	ret = false
	switch ev := event.(type) {
	case *EventThreadsThread:
		log.Printf("EventThreadsThread ev=%v", ev)
		iterate(NotMuchDataBase, ev.id)
		this.ThreadEntry = ev.ThreadEntry
		ret = true
	}
	return
}

func (this *Mails) _gpgme() {
	// see ~/go/pkg/mod/github.com/proglottis/gpgme@v0.1.1/gpgme.go
	if context, err := gpgme.New(); err != nil {
		panic(err)
	} else {
		defer context.Release()
		if keys, err := gpgme.FindKeys("mdt@emdete.de", false); err != nil {
			panic(err)
		} else {
			for _, key := range keys {
				userID := key.UserIDs()
				for userID != nil {
					log.Printf("userid email=%v, name=%v, comment=%v", userID.Email(), userID.Name(), userID.Comment())
					userID = userID.Next()
				}
				subKey := key.SubKeys()
				for subKey != nil {
					log.Printf("\tsubkey id=%v fp=%v", subKey.KeyID(), subKey.Fingerprint())
					subKey = subKey.Next()
				}
			}
		}
	}
}

func (this *Mails) _gmime3() {
	//_, _ = gmime3.Parse("")
}

// iterate over the mails of a thread
func iterate(db *notmuch.DB, id string) {
	log.Printf("-- BEGIN iterate %s", id)
	nmq := db.NewQuery(id)
	defer nmq.Close()
	if 1 != nmq.CountThreads() {
		panic("thread not found")
	}
	if threads, err := nmq.Threads(); err != nil {
		panic(err)
	} else {
		var thread *notmuch.Thread
		for threads.Next(&thread) {
			defer thread.Close()
			recurse(thread.Messages())
		}
	}
	log.Printf("-- END iterate")
}

// recurse through the mails of a thread
func recurse(messages *notmuch.Messages) {
	log.Printf("-- BEGIN recurse")
	message := &notmuch.Message{}
	for messages.Next(&message) {
		defer message.Close()
		//log.Printf("%v", message)
		log.Printf("%s: %v, '%s', '%s', '%s', '%s', '%s', '%s', '%s', ",
			message.Filename(), // string
			//message.Filenames(), // *Filenames
			message.Date(), // time.Time
			message.Header("From"), // string
			message.Header("To"), // string
			message.Header("CC"), // string
			message.Header("Subject"), // string
			message.Header("Return-Path"),
			message.Header("Delivered-To"),
			message.Header("Thread-Topic"),
			//message.Tags(), // *Tags
			//message.Properties(key string, exact bool), // *MessageProperties
		)
		if message.Header("Content-Type") == "multipart/encrypted" {
			decr(message.Filename())
		}
		if replies, err := message.Replies(); err == nil {
			defer replies.Close()
			//recurse(replies)
		}
	}
	log.Printf("-- END recurse")
}

// descrypt the mails of a thread
func decr(filename string) {
	if stream, err := os.Open(filename); err != nil {
		panic(err)
	} else {
		if data, err := gpgme.Decrypt(stream); err != nil {
			log.Printf("%v", err)
		} else {
			defer data.Close()
			buffer := make([]byte, 24)
			if count, err := data.Read(buffer); err != nil {
				panic(err)
			} else {
				log.Printf("%d, buffer=%v", count, buffer)
			}
		}
	}
}

