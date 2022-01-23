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

type Threads struct {
}

func NewThreads(s tcell.Screen) (this Threads) {
	log.Printf("NewThreads")
	this = Threads{
	}
	// gpgme
	this._gpgme()
	// gmime3
	this._gmime3()
	return
}

func (this *Threads) EventHandler(s tcell.Screen, ev tcell.Event) (ret bool) {
	ret = false
	return
}

func (this *Threads) _gpgme() {
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

func (this *Threads) _gmime3() {
	//_, _ = gmime3.Parse("")
}

