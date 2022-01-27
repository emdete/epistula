package main

import (
	"log"
	"os"
	"strings"
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
	Area // where do we write
	ThreadEntry // thread to show
	selected_index_message int // which message is selected (default: first unread)
	selected_index_part int // which part in that message is selected (default: first plain text)
}

func NewMails(s tcell.Screen) (this Mails) {
	log.Printf("NewMails")
	this = Mails{}
	return
}

const (
	MAILS_OPEN = '▼'
	MAILS_CLOSE = '▶'
	MAILS_MORE = '+'
)


// RuneULCorner = '┌' // RuneTTee  = '┬' // RuneURCorner = '┐'
// RuneLTee     = '├' // RuneHLine = '─' // RuneRTee     = '┤'
// RuneLLCorner = '└' // RuneBTee  = '┴' // RuneLRCorner = '┘'
// RuneVLine =    '│' 

func (this *Mails) drawMessage(s tcell.Screen, px, py int, envelope, decrypted *gmime.Envelope, isencrypted, show bool) int {
	style_normal := tcell.StyleDefault.Background(tcell.ColorLightGray)
	if isencrypted {
		style_normal = style_normal.Foreground(tcell.ColorDarkGreen)
	} else {
		style_normal = style_normal.Foreground(tcell.ColorDarkRed)
	}
	style_header := style_normal.Reverse(true)
	if !show {
		style_normal = style_normal.Background(tcell.ColorDarkGray)
	}
	style_frame := tcell.StyleDefault.Foreground(tcell.ColorLightGray)
	y := 0
	w := this.dx
	this.SetString(s, px, py+y, style_header, " " + envelope.Header("Subject"), w)
	if show {
		this.SetContent(s, px+w-1, py+y, MAILS_OPEN, nil, style_header)
	} else {
		this.SetContent(s, px+w-1, py+y, MAILS_CLOSE, nil, style_header)
	}
	y++
	// from now of we have a RuneVLine, on the left, so text is indented
	w-- // indent reduced width
	this.SetContent(s, px, py+y, tcell.RuneVLine, nil, style_frame)
	this.SetString(s, px+1, py+y, style_normal, envelope.Header("Date"), w)
	y++
	this.SetContent(s, px, py+y, tcell.RuneVLine, nil, style_frame)
	this.SetString(s, px+1, py+y, style_normal, "From: " + envelope.Header("From"), w)
	y++
	this.SetContent(s, px, py+y, tcell.RuneVLine, nil, style_frame)
	this.SetString(s, px+1, py+y, style_normal, "To: " + envelope.Header("To"), w)
	y++
	if envelope.Header("CC") != "" {
		this.SetContent(s, px, py+y, tcell.RuneVLine, nil, style_frame)
		this.SetString(s, px+1, py+y, style_normal, "CC: " + envelope.Header("CC"), w)
		y++
	}
	if decrypted != nil {
		envelope = decrypted
	}
	if show {
		even := true
		style_normal_dim := style_normal.Background(tcell.ColorDarkGray)
		if err := envelope.Walk(func (part *gmime.Part) error {
			style := style_normal
			if even {
				style = style_normal_dim
			}
			this.SetContent(s, px, py+y, tcell.RuneVLine, nil, style_frame)
			this.SetString(s, px+1, py+y, style, part.ContentType() + " " + part.Filename(), w)
			this.SetContent(s, px+w-1, py+y, MAILS_CLOSE, nil, style)
			y++
			log.Printf("contentype=%s", part.ContentType())
			if part.ContentType() == "message/rfc822" {
				if envlp, err := gmime.Parse(part.Text()); err != nil {
					log.Printf("inner message parsing error=%s", err)
				} else {
					log.Printf("inner from=%s", envlp.Header("From"))
				}
			} else if part.IsText() {
				//log.Printf("text=%s", part.Text())
				if part.ContentType() == "text/plain" {
					this.SetContent(s, px+w-1, py+y-1, MAILS_OPEN, nil, style)
					c := 0
					paragraphprefix := ""
					lastparagraphempty := true
					for _, paragraph := range strings.Split(part.Text(), "\n") {
						oy := y
						paragraph = strings.TrimSpace(paragraph)
						if !lastparagraphempty || len(paragraph) > 0 {
							_, y = this.SetParagraph(s, px+1, py+y, style, paragraphprefix, paragraph, w)
							for oy < y {
								this.SetContent(s, px, py+oy, tcell.RuneVLine, nil, style_frame)
								oy++
							}
						}
						lastparagraphempty = len(paragraph) == 0
						c++
						if c > 12 {
							this.SetContent(s, px+w-1, py+y-1, MAILS_MORE, nil, style)
							break
						}
					}
				}
			} else if part.IsAttachment() {
				log.Printf("attachment filename=%s", part.Filename())
			}
			even = !even
			return nil
		}); err != nil {
			panic(nil)
		}
	}
	this.SetContent(s, px, py+y, tcell.RuneLLCorner, nil, style_frame)
	y++
	return y
}

func (this *Mails) Draw(s tcell.Screen) (ret bool) {
	this.ClearArea(s)
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
				px := 0
				dx := 0
				py := 0
				for messages.Next(&message) {
					defer message.Close()
					envelope := parseMessage(message)
					defer envelope.Close()
					isencrypted := MessageHasTag(message, "encrypted")
					py += this.drawMessage(s, px, py, envelope, decryptMessage(message, show && isencrypted), isencrypted, show)
					py-- // put subject right of the RuneLLCorner, last line of last message
					px++ // indent next
					dx-- // indent reduces width
					if false { // TODO core
						if replies, err := message.Replies(); err != nil {
							log.Printf("error %v on getting replies", err)
						} else {
							defer replies.Close()
							reply := &notmuch.Message{}
							for replies.Next(&reply) {
								defer reply.Close()

							}
						}
					}
					show = false
				}
			}
		}
	}
	return true
}

func (this *Mails) EventHandler(s tcell.Screen, event tcell.Event) {
	log.Printf("Mails.EventHandler %v", event)
	switch ev := event.(type) {
	case *EventThreadsThread:
		this.ThreadEntry = ev.ThreadEntry
		this.dirty = true
	}
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

func parseMessage(message *notmuch.Message) *gmime.Envelope {
	if fh, err := os.Open(message.Filename()); err != nil {
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

func decryptMessage(message *notmuch.Message, decrypt bool) *gmime.Envelope {
	if !decrypt {
		return nil
	}
	if stream, err := os.Open(message.Filename()); err != nil {
		panic(err)
	} else {
		if fh, err := gpgme.Decrypt(stream); err != nil {
			log.Printf("decryptMessage error %v", err)
		} else {
			defer fh.Close()
			if data, err := ioutil.ReadAll(bufio.NewReader(fh)); err != nil {
				panic(err)
			} else {
				if envelope, err := gmime.Parse(string(data)); err != nil {
					panic(err)
				} else {
					return envelope
				}
			}
		}
	}
	return nil
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

