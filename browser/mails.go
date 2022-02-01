package main

import (
	"log"
	"fmt"
	"os/exec"
	"io"
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
	// see ~/go/pkg/mod/github.com/arran4/golang-ical@v0.0.0-20220115055431-e3ae8290e7b8/
	"github.com/arran4/golang-ical"
)

type Mails struct {
	Area // where do we write
	ThreadEntry // thread to show
	paged_y int // offset into mails
	textlinelimit int // count of lines of text initially shown
	selected_index_message int // which message is selected (default: first unread)
	selected_index_part int // which part in that message is selected (default: first plain text)
	count_of_lines int
	cache map[IntPair]IntPair // cache of positions of open/close controls
}

func NewMails(s tcell.Screen) (this Mails) {
	log.Printf("NewMails")
	this = Mails{}
	this.cache = make(map[IntPair]IntPair)
	return
}

const (
	MAILS_OPEN = '▼'
	MAILS_CLOSE = '▶'
	MAILS_MORE = '+'
	MAILS_TEXTLINELIMIT = 12
)

// this pair caches positions of the controls like MAILS_MORE on the screen so
// we know where to look at if the user klicks one. it maps a screen position
// (or better area position) to a mail idx/part idx in Mails.cache.
type IntPair struct {
	a,b int
}

// RuneULCorner = '┌' // RuneTTee  = '┬' // RuneURCorner = '┐'
// RuneLTee     = '├' // RuneHLine = '─' // RuneRTee     = '┤'
// RuneLLCorner = '└' // RuneBTee  = '┴' // RuneLRCorner = '┘'
// RuneVLine =    '│' 

func (this *Mails) drawMessage(s tcell.Screen, px, py int, envelope, decrypted *gmime.Envelope, index_message int, isencrypted, selected bool) int {
	style_normal := tcell.StyleDefault.Background(tcell.ColorLightGray)
	selected_style := tcell.StyleDefault.Background(tcell.GetColor("#eeeeee"))
	if isencrypted {
		style_normal = style_normal.Foreground(tcell.ColorDarkGreen)
	} else {
		style_normal = style_normal.Foreground(tcell.ColorDarkRed)
	}
	style_header := style_normal.Reverse(true)
	if !selected {
		style_normal = style_normal.Background(tcell.ColorDarkGray)
	}
	style_frame := tcell.StyleDefault.Foreground(tcell.ColorLightGray)
	w := this.dx - px
	this.SetString(s, px, py, style_header, " " + envelope.Header("Subject"), w)
	if selected {
		this.SetContent(s, px+w-2, py, MAILS_OPEN, nil, style_header)
		this.SetContent(s, px+w-1, py, ' ', nil, selected_style)
	} else {
		this.SetContent(s, px+w-2, py, MAILS_CLOSE, nil, style_header)
	}
	this.cache[IntPair{px+w-2,py}] = IntPair{index_message,-1}
	py++
	// from now on we have a RuneVLine, on the left, so text is indented
	w-- // indent reduced width
	this.SetContent(s, px, py, tcell.RuneVLine, nil, style_frame)
	this.SetString(s, px+1, py, style_normal, envelope.Header("Date"), w)
	if selected { this.SetContent(s, px+w, py, ' ', nil, selected_style) }
	py++
	this.SetContent(s, px, py, tcell.RuneVLine, nil, style_frame)
	this.SetString(s, px+1, py, style_normal, "From: " + envelope.Header("From"), w)
	if selected { this.SetContent(s, px+w, py, ' ', nil, selected_style) }
	py++
	this.SetContent(s, px, py, tcell.RuneVLine, nil, style_frame)
	this.SetString(s, px+1, py, style_normal, "To: " + envelope.Header("To"), w)
	if selected { this.SetContent(s, px+w, py, ' ', nil, selected_style) }
	py++
	if envelope.Header("CC") != "" {
		this.SetContent(s, px, py, tcell.RuneVLine, nil, style_frame)
		this.SetString(s, px+1, py, style_normal, "CC: " + envelope.Header("CC"), w)
		if selected { this.SetContent(s, px+w, py, ' ', nil, selected_style) }
		py++
	}
	if decrypted != nil {
		envelope = decrypted
	}
	if selected {
		style_normal_dim := style_normal.Background(tcell.ColorDarkGray)
		index_message_part := 0
		if err := envelope.Walk(func (part *gmime.Part) error {
			this.SetContent(s, px, py, tcell.RuneVLine, nil, style_frame)
			this.SetString(s, px+1, py, style_normal_dim, part.ContentType() + " " + part.Filename(), w)
			this.SetContent(s, px+w-1, py, MAILS_CLOSE, nil, style_normal_dim)
			this.cache[IntPair{px+w-1,py}] = IntPair{index_message,index_message_part}
			if selected { this.SetContent(s, px+w, py, ' ', nil, selected_style) }
			py++
			if index_message_part == this.selected_index_part {
				//if part.ContentType() == "message/rfc822" { if envlp, err := gmime.Parse(part.Text()); err != nil {}}
				if part.IsText() {
					this.SetContent(s, px+w-1, py-1, MAILS_OPEN, nil, style_normal_dim)
					textline := 0
					paragraphprefix := ""
					lastparagraphempty := true
					text := ""
					if part.ContentType() == "text/plain" {
						text = part.Text()
					} else if part.ContentType() == "text/html" {
						text, _ = HtmlToPlaintext(part.Text())
					} else if part.ContentType() == "text/calendar" {
						text, _ = ICalToPlaintext(part.Text())
					} else {
						log.Printf("unknown text type %s", part.ContentType())
					}
					for _, paragraph := range strings.Split(text, "\n") {
						oy := py
						paragraph = strings.TrimSpace(paragraph)
						if !lastparagraphempty || len(paragraph) > 0 {
							_, py = this.SetParagraph(s, px+1, py, style_normal, paragraphprefix, paragraph, w)
							textline += py-oy
							for oy < py {
								this.SetContent(s, px, oy, tcell.RuneVLine, nil, style_frame)
								if selected { this.SetContent(s, px+w, oy, ' ', nil, selected_style) }
								oy++
							}
						}
						lastparagraphempty = len(paragraph) == 0
						if textline > this.textlinelimit {
							this.SetContent(s, px+w-1, py-1, MAILS_MORE, nil, style_normal)
							this.cache[IntPair{px+w-1,py-1}] = IntPair{index_message,-2}
							break
						}
					}
				} else if part.IsAttachment() {
					//
				}
			}
			index_message_part++
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
		query := db.NewQuery("thread:" + this.id)
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
				index_message := 0
				var recurse func (messages *notmuch.Messages, px, py int) int
				recurse = func (messages *notmuch.Messages, px, py int) int {
					message := &notmuch.Message{}
					for messages.Next(&message) {
						defer message.Close()
						{
							envelope := parseMessage(message)
							defer envelope.Close()
							isencrypted := MessageHasTag(message, "encrypted")
							selected := index_message == this.selected_index_message
							py = this.drawMessage(s, px, py, envelope, decryptMessage(message, selected && isencrypted), index_message, isencrypted, selected)
						}
						index_message++
						if replies, err := message.Replies(); err == nil {
							defer replies.Close()
							// put subject right of the RuneLLCorner, last line of last message
							// indent next
							py = recurse(replies, px+1, py)
						}
					}
					return py
				}
				this.count_of_lines = recurse(messages, 0, -this.paged_y)+this.paged_y
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
	log.Printf("Mails.EventHandler %#v", event)
	switch ev := event.(type) {
	case *EventThreadsThread:
		this.ThreadEntry = ev.ThreadEntry
		this.paged_y = 0
		this.selected_index_message = 0
		this.textlinelimit = MAILS_TEXTLINELIMIT
		this.selected_index_part = 0
		this.cache = make(map[IntPair]IntPair)
		this.dirty = true
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyPgUp:
			if this.paged_y > this.dy-1 {
				this.paged_y -= this.dy - 1
			} else {
				this.paged_y = 0
			}
			this.dirty = true
		case tcell.KeyPgDn:
			if this.paged_y + this.dy < this.count_of_lines {
				this.paged_y += this.dy - 1
				this.dirty = true
			}
		}
	case *tcell.EventMouse:
		button := ev.Buttons()
		switch button {
		case tcell.Button1:
			x, y := ev.Position()
			x -= this.px
			y -= this.py
			if m, found := this.cache[IntPair{x, y}]; found {
				log.Printf("x=%d, y=%d, m=%v:%#v", x, y, found, m)
				switch m.b {
				case -2:
					// unlimit text line count
					this.textlinelimit += MAILS_TEXTLINELIMIT
				case -1:
					// select mail
					this.selected_index_message = m.a
					this.selected_index_part = 0
				default:
					// select, open part
					this.selected_index_message = m.a
					this.selected_index_part = m.b
				}
				this.dirty = true
			}
		case tcell.WheelUp:
			if this.paged_y > 0 {
				this.paged_y--
				this.dirty = true
			}
		case tcell.WheelDown:
			this.paged_y++
			this.dirty = true
		}
	}
	log.Printf("Mails.EventHandler %#v", this)
}

func (this *Mails) GetSelectedMailFilename() string {
	if this.id == "" {
		return ""
	}
	if db, err := notmuch.Open(NotMuchDatabasePath, notmuch.DBReadOnly); err != nil {
		return ""
	} else {
		defer db.Close()
		query := db.NewQuery("thread:" + this.id)
		defer query.Close()
		if 1 != query.CountThreads() {
			return ""
		}
		if threads, err := query.Threads(); err != nil {
			return ""
		} else {
			var thread *notmuch.Thread
			if threads.Next(&thread) {
				defer thread.Close()
				index_message := 0
				var recurse func (messages *notmuch.Messages) string
				recurse = func (messages *notmuch.Messages) string {
					message := &notmuch.Message{}
					for messages.Next(&message) {
						defer message.Close()
						if this.selected_index_message == index_message {
							return message.Filename()
						}
						index_message++
						if replies, err := message.Replies(); err == nil {
							defer replies.Close()
							n := recurse(replies)
							if n != "" {
								return n
							}
						}
					}
					return ""
				}
				return recurse(thread.TopLevelMessages())
			} else {
				return ""
			}
			if threads.Next(&thread) {
				return ""
			}
		}
	}
	return ""
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

func HtmlToPlaintext(content string) (string, error) {
	result := ""
	cmd := exec.Command(
		"elinks", "-force-html", "-dump-charset", "utf-8", "-dump",
		//"w3m", "-T", "text/html", "-I", "utf-8", "-O", "utf-8",
		//"pandoc", "--reference-links", "-f", "html", "-t", "plain",
		)
	if stdin, err := cmd.StdinPipe(); err != nil {
		return "", err
	} else {
		go func() {
			defer stdin.Close()
			if _, err := io.WriteString(stdin, content); err != nil {
				panic(err)
			}
		}()
	}
	if stdout, err := cmd.StdoutPipe(); err != nil {
		return "", err
	} else {
		go func() {
			defer stdout.Close()
			if data, err := ioutil.ReadAll(bufio.NewReader(stdout)); err != nil {
				panic(err)
			} else {
				result = string(data)
			}
		}()
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	if err := cmd.Wait(); err != nil {
		return "", err
	}
	return result, nil
}

func ICalToPlaintext(content string) (string, error) {
	ioutil.WriteFile("/tmp/m.ical", []byte(content), 0644)
	result := ""
	if cal, err := ics.ParseCalendar(strings.NewReader(content)); err != nil {
		return "", err
	} else {
		log.Printf("%#v", cal)
		for _, component := range cal.Components {
			//t,_ := component.GetSummary()
			//t,_ := component.GetOrganizer()
			//t,_ := component.GetDescription()
			//t,_ := component.GetAttendee()
			//t,_ := component.GetLocation()
			//result = result + fmt.Sprintf("%s\n", t)
			log.Printf("%#v", component)
		}
		for _, event := range cal.Events() {
			s,_ := event.GetStartAt()
			e,_ := event.GetEndAt()
			result = result + fmt.Sprintf("%s - %s\n", s.String(), e.String())
		}
	}
	return result, nil
}

