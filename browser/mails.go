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
	w := this.dx
	this.SetString(s, px, py, style_header, " " + envelope.Header("Subject"), w)
	if show {
		this.SetContent(s, px+w-1, py, MAILS_OPEN, nil, style_header)
	} else {
		this.SetContent(s, px+w-1, py, MAILS_CLOSE, nil, style_header)
	}
	py++
	// from now of we have a RuneVLine, on the left, so text is indented
	w-- // indent reduced width
	this.SetContent(s, px, py, tcell.RuneVLine, nil, style_frame)
	this.SetString(s, px+1, py, style_normal, envelope.Header("Date"), w)
	py++
	this.SetContent(s, px, py, tcell.RuneVLine, nil, style_frame)
	this.SetString(s, px+1, py, style_normal, "From: " + envelope.Header("From"), w)
	py++
	this.SetContent(s, px, py, tcell.RuneVLine, nil, style_frame)
	this.SetString(s, px+1, py, style_normal, "To: " + envelope.Header("To"), w)
	py++
	if envelope.Header("CC") != "" {
		this.SetContent(s, px, py, tcell.RuneVLine, nil, style_frame)
		this.SetString(s, px+1, py, style_normal, "CC: " + envelope.Header("CC"), w)
		py++
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
			this.SetContent(s, px, py, tcell.RuneVLine, nil, style_frame)
			this.SetString(s, px+1, py, style, part.ContentType() + " " + part.Filename(), w)
			this.SetContent(s, px+w-1, py, MAILS_CLOSE, nil, style)
			py++
			//if part.ContentType() == "message/rfc822" { if envlp, err := gmime.Parse(part.Text()); err != nil {}}
			//if part.IsAttachment() {}
			if part.IsText() && part.ContentType() == "text/plain" {
				this.SetContent(s, px+w-1, py-1, MAILS_OPEN, nil, style)
				c := 0
				paragraphprefix := ""
				lastparagraphempty := true
				for _, paragraph := range strings.Split(part.Text(), "\n") {
					oy := py
					paragraph = strings.TrimSpace(paragraph)
					if !lastparagraphempty || len(paragraph) > 0 {
						_, py = this.SetParagraph(s, px+1, py, style, paragraphprefix, paragraph, w)
						for oy < py {
							this.SetContent(s, px, oy, tcell.RuneVLine, nil, style_frame)
							oy++
						}
					}
					lastparagraphempty = len(paragraph) == 0
					c++
					if c > 12 {
						this.SetContent(s, px+w-1, py-1, MAILS_MORE, nil, style)
						break
					}
				}
			}
			even = !even
			return nil
		}); err != nil {
			panic(nil)
		}
	}
	this.SetContent(s, px, py, tcell.RuneLLCorner, nil, style_frame)
	return py
}

func (this *Mails) Draw(s tcell.Screen) (ret bool) {
	this.ClearArea(s)
	if this.id == "" {
		return true
	}
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
			log.Printf("Mails.Draw thread id=%s not found", this.id)
			return
		} else {
			var thread *notmuch.Thread
			if threads.Next(&thread) {
				defer thread.Close()
				messages := thread.TopLevelMessages()
				show := true
				var recurse func (messages *notmuch.Messages, px, py int) int
				recurse = func (messages *notmuch.Messages, px, py int) int {
					message := &notmuch.Message{}
					for messages.Next(&message) {
						defer message.Close()
						{
							envelope := parseMessage(message)
							defer envelope.Close()
							isencrypted := MessageHasTag(message, "encrypted")
							py = this.drawMessage(s, px, py, envelope, decryptMessage(message, show && isencrypted), isencrypted, show)
						}
						if replies, err := message.Replies(); err == nil {
							defer replies.Close()
							// put subject right of the RuneLLCorner, last line of last message
							// indent next
							py = recurse(replies, px+1, py)
						}
						show = false
					}
					return py
				}
				recurse(messages, 0, 0)
			} else {
				log.Printf("Mails.Draw not found: %s", this.id)
			}
			if threads.Next(&thread) {
				log.Printf("Mails.Draw not uniq: %s", this.id)
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

