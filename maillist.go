package main

import (
	"log"
	// +tcell
	"github.com/gdamore/tcell/v2"
	// +pgpme
	"github.com/proglottis/gpgme"
	// +gmime3
	_ "github.com/sendgrid/go-gmime"
)
// Threads
// keys: PgUp PgDn
type Threads struct {
}

func NewThreads(s tcell.Screen) (this Threads) {
	this = Threads{
	}
	// gpgme
	_gpgme()
	// gmime3
	_gmime3()
	return
}

func (this *Threads) EventHandler(s tcell.Screen, ev tcell.Event) (ret bool) {
	ret = false
	return
}

func _gpgme() {
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

func _gmime3() {
	//_, _ = gmime3.Parse("")
}

