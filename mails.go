package main

import (
	"log"
	// see ~/go/pkg/mod/github.com/gdamore/tcell/v2@v2.4.1-0.20210905002822-f057f0a857a1/
	"github.com/gdamore/tcell/v2"
	// see
	"github.com/proglottis/gpgme"
	// see
	_ "github.com/sendgrid/go-gmime"
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
	// RuneBTee     = '┴'
	// RuneHLine    = '─'
	// RuneLLCorner = '└'
	// RuneLRCorner = '┘'
	// RuneLTee     = '├'
	// RuneRTee     = '┤'
	// RuneTTee     = '┬'
	// RuneULCorner = '┌'
	// RuneURCorner = '┐'
	// RuneVLine    = '│'
	cs := tcell.StyleDefault.Reverse(true)
	emitStr(s, px, py, cs, " " + this.subject, w)
	for row := 1; row < 24; row++ {
		s.SetCell(px, py+row, tcell.StyleDefault, tcell.RuneVLine)
	}
	emitStr(s, px+1, py+1, tcell.StyleDefault, "From: " + this.author, w)
	//emitStr(s, px+1, py+1, tcell.StyleDefault, this.newest, w)
	s.SetCell(px, py+24, tcell.StyleDefault, tcell.RuneLLCorner)
	emitStr(s, px+1, py+24, cs, " " + this.subject, w)
	return true
}

func (this *Mails) EventHandler(s tcell.Screen, event tcell.Event) (ret bool) {
	ret = false
	switch ev := event.(type) {
	case *EventThreadsThread:
		log.Printf("EventThreadsThread ev=%v", ev)
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
