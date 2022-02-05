package main

import (
	"log"
	"os/exec"
	"os"
	"strings"
	// see ~/go/pkg/mod/github.com/gdamore/tcell/v2@v2.4.1-0.20210905002822-f057f0a857a1/
	"github.com/gdamore/tcell/v2"
	// see ~/go/pkg/mod/github.com/sendgrid/go-gmime@v0.0.0-20211124164648-4c44cbd981d8/
	"github.com/sendgrid/go-gmime/gmime"
	// see ~/go/pkg/mod/github.com/zenhack/go.notmuch@v0.0.0-20211022191430-4d57e8ad2a8b/
	"github.com/zenhack/go.notmuch"
)

type Mails struct {
	Area // where do we write
	ThreadEntry // thread to show
	paged_y int // offset into mails
	textlinelimit int // count of lines of text initially shown
	selected_index_message int // which message is selected (default: first unread)
	selected_message_filename string // cache the filename of the selected message (for reply)
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
				//} else if part.IsAttachment() {
				//
				} else {
					// if we cannot display this part, increase
					// selected_index_part, we are at index_message_part ==
					// this.selected_index_part already so if we cant be
					// displayd lets try the next one
					this.selected_index_part++
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
							envelope := parseMessage(message.Filename())
							defer envelope.Close()
							isencrypted := MessageHasTag(message, "encrypted")
							selected := index_message == this.selected_index_message
							if selected {
								this.selected_message_filename = message.Filename()
							}
							py = this.drawMessage(s, px, py, envelope, decryptMessage(message.Filename(), selected && isencrypted), index_message, isencrypted, selected)
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
	// log.Printf("Mails.EventHandler %#v", event)
	switch ev := event.(type) {
	case *EventThreadsThread:
		this.ThreadEntry = ev.ThreadEntry
		this.paged_y = 0
		this.selected_index_message = this.count-1
		this.textlinelimit = MAILS_TEXTLINELIMIT
		this.selected_index_part = 0
		this.cache = make(map[IntPair]IntPair)
		this.dirty = true
log.Printf("Mails.Draw after clear, if=%s", this.id)
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlC: // compose new email
			this.compose()
		case tcell.KeyCtrlR: // reply to selected email
			if mailfilename := mails.GetSelectedMailFilename(); mailfilename != "" {
				mails.reply(mailfilename)
			}
		case tcell.KeyCtrlJ:
			if this.selected_index_message+1 < this.count {
				this.selected_index_message++
				this.selected_index_part = 0
				this.dirty = true
			}
		case tcell.KeyCtrlK:
			if this.selected_index_message > 0 {
				this.selected_index_message--
				this.selected_index_part = 0
				this.dirty = true
			}
		case tcell.KeyCtrlO:
			this.textlinelimit += MAILS_TEXTLINELIMIT
			this.dirty = true
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
	// log.Printf("Mails.EventHandler %#v", this)
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

func (this *Mails) compose() {
	cwd,_ := os.Getwd()
	cmd := exec.Command("gnome-terminal",
		"--wait",
		"--hide-menubar",
		"--working-directory=" + cwd,
		"--",
		"../composer/epistula-composer",
		)
	go cmd.Run()
}

func (this *Mails) reply(message_filename string) {
	log.Printf("reply %s", message_filename)
	envelope := parseMessage(message_filename)
	defer envelope.Close()
	cc := envelope.Header("Cc")
	from := envelope.Header("From")
	message_id := envelope.Header("Message-ID")
	references := envelope.Header("References")
	in_reply_to := envelope.Header("In-Reply-To")
	reply_to := envelope.Header("Reply-To")
	subject := envelope.Subject()
	to := envelope.Header("To")
	var text string
	if e := decryptMessage(message_filename, true); e != nil {
		envelope = e
	}
	index_message_part := 0
	if err := envelope.Walk(func (part *gmime.Part) error {
		log.Printf("index_message_part=%s %s", index_message_part, part.ContentType())
		if index_message_part == this.selected_index_part {
			if part.IsText() {
				if part.ContentType() == "text/plain" {
					text = part.Text()
				} else if part.ContentType() == "text/html" {
					text, _ = HtmlToPlaintext(part.Text())
				} else if part.ContentType() == "text/calendar" {
					text, _ = ICalToPlaintext(part.Text())
				} else {
					log.Printf("unknown text type %s", part.ContentType())
				}
			}
		}
		index_message_part++
		return nil
	}); err != nil {
		panic(nil)
	}
	var tempfilename string
	if f, err := os.CreateTemp("", "epistula-browser-"); err != nil {
		log.Fatal(err)
	} else {
		if _, err := f.Write([]byte(text)); err != nil {
			log.Fatal(err)
		}
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
		tempfilename = f.Name()
	}
	//log.Printf("cc=%s,from=%s,message_id=%s,reply_to=%s,subject=%s,to=%s, tempfilename=%s", cc, from, message_id, reply_to, subject, to, tempfilename)
	cwd,_ := os.Getwd()
	cmd := exec.Command("gnome-terminal",
		"--wait",
		"--hide-menubar",
		"--working-directory=" + cwd,
		"--",
		"epistula-composer",
			"--bcc=",
			"--cc=" + cc,
			"--from=" + from,
			"--message-id=" + message_id,
			"--references=" + references,
			"--reply-to=" + reply_to,
			"--in-reply-to=" + in_reply_to,
			"--subject=" + subject,
			"--text=" + tempfilename ,
			"--to=" + to,
		)
	go cmd.Run()
}

