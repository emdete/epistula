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
	return
}

// RuneULCorner = '┌' // RuneTTee  = '┬' // RuneURCorner = '┐'
// RuneLTee     = '├' // RuneHLine = '─' // RuneRTee     = '┤'
// RuneLLCorner = '└' // RuneBTee  = '┴' // RuneLRCorner = '┘'
// RuneVLine =    '│' 

func (this *Mails) drawMessage(s tcell.Screen, px, py, w, h int, message *notmuch.Message) int {
	cs1 := tcell.StyleDefault.Background(tcell.ColorLightGray)
	isencrypted := false
	tags := message.Tags()
	var tag *notmuch.Tag
	for tags.Next(&tag) {
		if tag.Value == "encrypted" {
			isencrypted = true
		}
	}
	if isencrypted {
		cs1 = cs1.Foreground(tcell.ColorDarkGreen)
	} else {
		cs1 = cs1.Foreground(tcell.ColorDarkRed)
	}
	cs2 := cs1.Reverse(true)
	cs3 := tcell.StyleDefault.Foreground(tcell.ColorLightGray)
	y := 0
	emitStr(s, px, py+y, cs2, " " + message.Header("Subject"), w)
	y++
	s.SetCell(px, py+y, cs3, tcell.RuneVLine)
	emitStr(s, px+1, py+y, cs1, message.Date().String(), w)
	y++
	s.SetCell(px, py+y, cs3, tcell.RuneVLine)
	emitStr(s, px+1, py+y, cs1, "From: " + message.Header("From"), w)
	y++
	s.SetCell(px, py+y, cs3, tcell.RuneVLine)
	emitStr(s, px+1, py+y, cs1, "To: " + message.Header("To"), w)
	y++
	if message.Header("CC") != "" {
		s.SetCell(px, py+y, cs3, tcell.RuneVLine)
		emitStr(s, px+1, py+y, cs1, "CC: " + message.Header("CC"), w)
		y++
	}
	s.SetCell(px, py+y, cs3, tcell.RuneLLCorner)
	y++
	return y
}

func (this *Mails) Draw(s tcell.Screen, px, py, w, h int) (ret bool) {
	if this.id == "" {
		return true
	}
	log.Printf("Mails.Draw '%v'", this.id)
	query := NotMuchDataBase.NewQuery(this.id)
	defer query.Close()
	if 1 != query.CountThreads() {
		return
	}
	if threads, err := query.Threads(); err != nil {
		return
	} else {
		var thread *notmuch.Thread
		for threads.Next(&thread) {
			defer thread.Close()
			message := &notmuch.Message{}
			messages := thread.Messages()
			for messages.Next(&message) {
				defer message.Close()
				py += this.drawMessage(s, px, py, w, h, message)
				py--
				px++
			}
		}
	}
	return true
}

func (this *Mails) EventHandler(s tcell.Screen, event tcell.Event) (ret bool) {
	ret = false
	log.Printf("Mails.EventHandler %v", event)
	switch ev := event.(type) {
	case *EventThreadsThread:
		this.ThreadEntry = ev.ThreadEntry
		//retrieveMail(NotMuchDataBase, this.id)
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

// retrieveMail over the mails of a thread
func retrieveMail(db *notmuch.DB, id string) {
	log.Printf("-- [ retrieveMail %s", id)
	query := db.NewQuery(id)
	defer query.Close()
	if 1 != query.CountThreads() {
		panic("thread not found")
	}
	if threads, err := query.Threads(); err != nil {
		panic(err)
	} else {
		var thread *notmuch.Thread
		for threads.Next(&thread) {
			defer thread.Close()
			recurse(thread.Messages())
		}
	}
	log.Printf("-- ] retrieveMail")
}

// recurse through the mails of a thread
func recurse(messages *notmuch.Messages) {
	log.Printf("-- [ recurse")
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
	log.Printf("-- ] recurse")
}

// descrypt the mails of a thread
func decr(filename string) {
	log.Printf("-- [ decr")
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
	log.Printf("-- ] decr")
}

