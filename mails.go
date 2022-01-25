package main

import (
	"log"
	"os"
	"io/ioutil"
	"bufio"
	// see ~/go/pkg/mod/github.com/gdamore/tcell/v2@v2.4.1-0.20210905002822-f057f0a857a1/
	"github.com/gdamore/tcell/v2"
	// see ~/go/pkg/mod/github.com/proglottis/gpgme@v0.1.1
	"github.com/proglottis/gpgme"
	// see ~/go/pkg/mod/github.com/sendgrid/go-gmime@v0.0.0-20211124164648-4c44cbd981d8/
	"github.com/sendgrid/go-gmime/gmime"
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

func (this *Mails) drawMessage(s tcell.Screen, px, py, w, h int, envelope *gmime.Envelope, isencrypted, show bool) int {
	cs1 := tcell.StyleDefault
	if show {
		cs1 = cs1.Background(tcell.ColorLightGray)
	} else {
		cs1 = cs1.Background(tcell.ColorGray)
	}
	if isencrypted {
		cs1 = cs1.Foreground(tcell.ColorDarkGreen)
	} else {
		cs1 = cs1.Foreground(tcell.ColorDarkRed)
	}
	cs2 := cs1.Reverse(true)
	cs3 := tcell.StyleDefault.Foreground(tcell.ColorLightGray)
	y := 0
	emitStr(s, px, py+y, cs2, " " + envelope.Header("Subject"), w)
	y++
	// from now of we have a RuneVLine, on the left, so text is indented
	w-- // indent reduced width
	s.SetCell(px, py+y, cs3, tcell.RuneVLine)
	emitStr(s, px+1, py+y, cs1, envelope.Header("Date"), w)
	y++
	s.SetCell(px, py+y, cs3, tcell.RuneVLine)
	emitStr(s, px+1, py+y, cs1, "From: " + envelope.Header("From"), w)
	y++
	s.SetCell(px, py+y, cs3, tcell.RuneVLine)
	emitStr(s, px+1, py+y, cs1, "To: " + envelope.Header("To"), w)
	y++
	if envelope.Header("CC") != "" {
		s.SetCell(px, py+y, cs3, tcell.RuneVLine)
		emitStr(s, px+1, py+y, cs1, "CC: " + envelope.Header("CC"), w)
		y++
	}
	envelope.Walk(func (part *gmime.Part) error {
		s.SetCell(px, py+y, cs3, tcell.RuneVLine)
		emitStr(s, px+1, py+y, cs1, part.ContentType(), w)
		y++
		log.Printf("contentype=%s", part.ContentType())
		if part.ContentType() == "plain/text" {
			log.Printf("text=%s", part.Text())

		} else if part.IsText() {
			log.Printf("text=%s", part.Text())
		} else if part.IsAttachment() {
			log.Printf("filename=%s", part.Filename())
		}
		return nil
	})
	s.SetCell(px, py+y, cs3, tcell.RuneLLCorner)
	y++
	return y
}

func (this *Mails) Draw(s tcell.Screen, px, py, w, h int) (ret bool) {
	if this.id == "" {
		return true
	}
	log.Printf("Mails.Draw '%v'", this.id)
	if db, err := notmuch.Open(NotMuchDatabasePath, notmuch.DBReadOnly); err != nil {
		panic(err)
	} else {
		defer db.Close()
		query := db.NewQuery(this.id)
		defer query.Close()
		if 1 != query.CountThreads() {
			return
		}
		if threads, err := query.Threads(); err != nil {
			log.Printf("Mails: thread id=%s not found", this.id)
			return
		} else {
			var thread *notmuch.Thread
			for threads.Next(&thread) {
				defer thread.Close()
				message := &notmuch.Message{}
				messages := thread.Messages()
				show := true
				for messages.Next(&message) {
					defer message.Close()
					py += this.drawMessage(s, px, py, w, h, parseMessageFile(message), MessageHasTag(message, "encrypted"), show)
					py-- // put subject right of the RuneLLCorner, last line of last message
					px++ // indent next
					w-- // indent reduces width
					show = false
				}
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
		ret = true
	}
	return
}

func (this *Mails) hasAdressKey(address string) bool {
	if context, err := gpgme.New(); err != nil {
		log.Printf("error %v on getting context", err)
	} else {
		defer context.Release()
		if keys, err := gpgme.FindKeys(address, false); err != nil {
			log.Printf("error %v on finding keys for %s", err, address)
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
			return true
		}
	}
	return false
}

func parseMessageFile(nmm *notmuch.Message) *gmime.Envelope {
	if fh, err := os.Open(nmm.Filename()); err != nil {
		panic(err)
	} else {
		defer fh.Close()
		if data, err := ioutil.ReadAll(bufio.NewReader(fh)); err != nil {
			panic(err)
		} else {
			if envelope, err := gmime.Parse(string(data)); err != nil {
				panic(err)
			} else {
				/*
				defer envelope.Close()
				if b, err := envelope.Export(); err != nil {
					log.Printf("%v\n", err)
				} else {
					log.Printf("%v\n", b)
				}
				*/
				return envelope
			}
		}
	}
	return nil
}

