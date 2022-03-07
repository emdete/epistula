package main

import (
	"log"
	"math/rand"
	"os"
	"time"
	"fmt"
	"strings"
	"os/exec"
	"io/ioutil"
	"bufio"
	//
	"github.com/sendgrid/go-gmime/gmime"
	_ "github.com/proglottis/gpgme"
)

const (
	CRLF = "\r\n"
	EDITOR = "nvim"
)

func _log() {
	log.SetPrefix("epistula ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC | log.Lshortfile)
	if f, err := os.OpenFile("/tmp/epistula-composer.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		log.Fatal(err)
	} else {
		os.Stderr = f
	}
	log.SetOutput(os.Stderr)
}

func main() {
	// log
	_log()
	log.Printf("main")
	//
	config := NewConfig()
	rand.Seed(time.Now().UnixMilli())
	// The Idea is as follows: the composer
	// - is called with all information in its arguments like --from, --to, --subject, --cc, --bcc, ...
	var origin_to, origin_reply_to, origin_from, origin_cc, origin_bcc, origin_subject, origin_in_reply_to, origin_message_id, origin_references, content_text string
	for i:=1;i<len(os.Args);i++ {
		if strings.HasPrefix(os.Args[i], "--") {
			x := strings.Split(os.Args[i][2:], "=")
			switch x[0] {
			case "bcc": origin_bcc = x[1]
			case "cc": origin_cc = x[1]
			case "from": origin_from = x[1]
			case "message-id": origin_message_id = x[1]
			case "references": origin_references = x[1]
			case "reply-to": origin_reply_to = x[1]
			case "in-reply-to": origin_in_reply_to = x[1]
			case "subject": origin_subject = x[1]
			case "to": origin_to = x[1]
			case "text":
				if fh, err := os.Open(x[1]); err != nil {
					panic(err)
				} else {
					defer fh.Close()
					if data, err := ioutil.ReadAll(bufio.NewReader(fh)); err != nil {
						panic(err)
					} else {
						content_text = string(data)
					}
				}
				defer os.Remove(x[1]) // clean up
			default:
				panic(fmt.Sprintf("wrong arg: %s", os.Args[i]))
			}
		} else {
			origin_from = os.Args[i]
			const PREFIX = "mailto:"
			if strings.HasPrefix(os.Args[i], PREFIX) {
				origin_from = origin_from[len(PREFIX):]
			}
		}
	}
	if origin_reply_to == "" {
		origin_reply_to = origin_from
	}
	log.Printf("origin_to=%s, origin_reply_to=%s, origin_from=%s, origin_cc=%s, origin_bcc=%s, origin_subject=%s, origin_message_id=%s, origin_reply_to=%s, ",
		origin_to,
		origin_reply_to,
		origin_from,
		origin_cc,
		origin_bcc,
		origin_subject,
		origin_message_id,
		origin_in_reply_to,
	)
	// - composes an email via gmime
	var buffer []byte
	date_string := time.Now().Format(time.RFC1123Z)
	// go-gmime doesnt support creation of envelopes or parts in envelopes yet.
	// so we create an empty dummy email and modify the elements after parsing
	// that
	log.Printf("here")
	if message, err := gmime.Parse(
		"Date: " + date_string + CRLF +
		"From: " + config.user_name + " <" + config.user_primary_email + ">" + CRLF +
		CRLF +
		CRLF); err != nil {
		panic(err)
	} else {
		message.ClearAddress("From")
		message.ParseAndAppendAddresses("From", config.user_name + " <" + config.user_primary_email + ">")
		message.ParseAndAppendAddresses("To", origin_reply_to) // TODO how to add an empty "To:", .. ?
		message.ParseAndAppendAddresses("To", origin_to) // if multiple to: exist reply to all of them
		// TODO remove myself
		message.ParseAndAppendAddresses("Cc", origin_cc)
		message.ParseAndAppendAddresses("Bcc", origin_bcc)
		message.SetSubject(origin_subject)
		message.SetHeader("X-Epistula-Status", "I am not done")
		message.SetHeader("X-Epistula-Comment", "This is your MUA talking to you. Add attachments as headerfield like below. Dont destroy the mail structure, if the outcome cant be parsed you will thrown into your editor again to fix it. Change the Status to not contain 'not'. Add a 'abort' to abort sending (editings lost).")
		message.SetHeader("X-Epistula-Attachment", "#sample entry#")
		if content_text != "" {
			// TODO add from & date
			content_text = "> " + strings.ReplaceAll(content_text, "\n", "\n> ")
		}
		if err := message.Walk(func (part *gmime.Part) error {
			if part.IsText() && part.ContentType() == "text/plain" {
				part.SetText(content_text)
			}
			return nil
		}); err != nil {
			panic(err)
		}
		if b, err := message.Export(); err != nil {
			panic(err)
		} else {
			buffer = b
		}
	}
	// - exports it to a temp file
	var tempfilename string
	if f, err := os.CreateTemp("", "epistula-composer-"); err != nil {
		panic(err)
	} else {
		if _, err := f.Write(buffer); err != nil {
			panic(err)
		}
		if err := f.Close(); err != nil {
			panic(err)
		}
		tempfilename = f.Name()
	}
	defer os.Remove(tempfilename)
	// - execs the editor and waits for its termination
	// set terminal title
	title := "Epistula Composer: " + config.user_name + " <" + config.user_primary_email + ">" + " to " + origin_reply_to
	os.Stdout.Write([]byte("\x1b]1;"+title+"\a\x1b]2;"+title+"\a"))
	//
	var message *gmime.Envelope
	done := false
	abort := false
	for !done {
		if EDITOR, err := exec.LookPath(EDITOR); err == nil {
			var procAttr os.ProcAttr
			procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
			if proc, err := os.StartProcess(EDITOR, []string{EDITOR,
					"+set ft=mail", // switch to email syntax
					"+set fileencoding=utf-8", // use utf8
					"+set enc=utf-8", // use utf8
					"+set fo+=w", // do wsf
					"+set fo-=ro", // dont repeat ">.." on new lines
					// "+set ff=unix",
					tempfilename}, &procAttr); err == nil {
				proc.Wait()
			}
		} else {
			panic(err)
		}
	// - parses the file via gmime
		message = ParseFile(tempfilename)
		status := message.Header("X-Epistula-Status")
		done = !strings.Contains(status, "not")
		abort = strings.Contains(status, "abort")
		if done && !abort {
			// check To: field
			if to := message.Header("To"); to == "" {
				log.Printf("To: is empty")
				done = false
			// create message
			// rewrite temp file
			}
		}
	}
	if abort {
		// the user flagged the message to be aborted
		os.Exit(0)
	}
	message.RemoveHeader("Fcc") // we do not support fcc
	message.RemoveHeader("X-Epistula-Status")
	message.RemoveHeader("X-Epistula-Comment")
	attachments := strings.Split(message.Header("X-Epistula-Attachment"), " ")
	for i:=0;i<len(attachments);i++ {
		if attachments[i][0] != '#' {
			if _, err := ioutil.ReadFile(attachments[i]); err != nil {
				log.Printf("error %s with %s", err, attachments[i])
			} else {
				// TODO add attachment
				//message.Attach(content, attachments[i])
			}
		}
	}
	message.RemoveHeader("X-Epistula-Attachment")
	message.ParseAndAppendAddresses("Reply-To", config.user_primary_email)
	message.SetHeader("MIME-Version", "1.0")
	message.SetHeader("User-Agent", "Epistula")
	message.SetHeader("Content-Type", "text/plain; charset=utf-8")
	message.SetHeader("Content-Transfer-Encoding", "quoted-printable")
	message.SetHeader("In-Reply-To", origin_message_id)
	//if origin_references == "" && origin_in_reply_to == "" -- rfc2822?!?
	message.SetHeader("References", origin_references + origin_message_id)
	message.SetHeader("Message-ID", MessageId(config.user_primary_email))
	// message.SetHeader("Content-ID", )
	// message.SetHeader("Thread-Topic", )
	// message.SetHeader("Comments", )
	// message.SetHeader("Keywords", )
	// - retreives the desired keys
	// - encrypts the file via gpgme
	// - sends the email
	if b, err := message.Export(); err != nil {
		panic(err)
	} else {
		buffer = b
	}
	cmd := exec.Command("sendmail", "-t", )
	if stdin, err := cmd.StdinPipe(); err != nil {
		panic(err)
	} else {
		go func() {
			defer stdin.Close()
			stdin.Write(buffer)
		}()
		if out, err := cmd.CombinedOutput(); err != nil {
			panic(err)
		} else {
			log.Printf("sendmail output: %s\n", out)
		}
	}
	// - saves the email in maildir and kicks off notmuch new, tag 'sent'
	cmd = exec.Command("notmuch", "insert", "+sent", "+inbox")
	if stdin, err := cmd.StdinPipe(); err != nil {
		panic(err)
	} else {
		go func() {
			defer stdin.Close()
			stdin.Write(buffer)
		}()
		if out, err := cmd.CombinedOutput(); err != nil {
			panic(err)
		} else {
			log.Printf("notmuch output: %s\n", out)
		}
	}
	// TODO send USR1 to browser to notify of db change
	// ioutil.WriteFile("/tmp/temp", buffer, 0600)
}

func ParseFile(filename string) *gmime.Envelope {
	if fh, err := os.Open(filename); err != nil {
		return nil
	} else {
		defer fh.Close()
		if data, err := ioutil.ReadAll(bufio.NewReader(fh)); err != nil {
			return nil
		} else {
			if envelope, err := gmime.Parse(string(data)); err != nil {
				return nil
			} else {
				return envelope
			}
		}
	}
	return nil
}

func MessageId(email string) string {
	return fmt.Sprintf("<epistula-%x-%x@%s>", rand.Uint64(), rand.Uint64(), strings.Split(email, "@")[1])
}
