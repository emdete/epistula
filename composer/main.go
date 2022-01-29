package main

import (
	"log"
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

func main() {
	log.SetPrefix("email ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC | log.Lshortfile)
	log.SetOutput(os.Stderr)
	log.Printf("main %#v", os.Args)
	// The Idea is as follows: the composeser
	// - is called with all information in its arguments like --from, --reply, --to, --subject, --cc, --bcc, ...
	var meta_from_name, meta_from, meta_to, meta_cc, meta_bcc, meta_subject, content_text string
	for i:=1;i<len(os.Args);i++ {
		if strings.HasPrefix(os.Args[i], "--") {
			x := strings.Split(os.Args[i][2:], "=")
			switch x[0] {
			case "from":
				meta_from = x[1]
			case "from-name":
				meta_from_name = x[1]
			case "to":
				meta_to = x[1]
			case "cc":
				meta_cc = x[1]
			case "bcc":
				meta_bcc = x[1]
			case "subject":
				meta_subject = x[1]
			case "reply":
				//
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
			default:
				panic(fmt.Sprintf("wrong arg: %s", os.Args[i]))
			}
		} else {
			panic(fmt.Sprintf("wrong arg: %s", os.Args[i]))
		}
	}
	// - decrypt reply email
	// - composes an email via gmime
	var buffer []byte
	date_string := time.Now().Format(time.RFC1123Z)
	if message, err := gmime.Parse(
		"Date: " + date_string + CRLF +
		"From: " + meta_from + CRLF +
		CRLF +
		CRLF); err != nil {
		panic(err)
	} else {
		message.ClearAddress("From")
		message.AddAddress("From", meta_from_name, meta_from)
		message.ParseAndAppendAddresses("To", meta_to) // TODO how to add an empty "To:", .. ?
		message.ParseAndAppendAddresses("Cc", meta_cc)
		message.ParseAndAppendAddresses("Bcc", meta_bcc)
		message.SetSubject(meta_subject)
		// Content-ID
		// Date: Thu, 13 Dec 2018 14:19:38 +0000
		// In-Reply-To
		// MIME-Version
		// Message-ID
		// References
		// Return-Path
		// Thread-Topic
		message.SetHeader("X-Epistula-State", "I am not done")
		message.SetHeader("X-Epistula-Comment", "This is your MUA talking to you. Add attachments as headerfield like below. Dont destroy the mail structure, if the outcome cant be parsed you will thrown into your editor again to fix it. Change the State to not contain 'not'.")
		message.SetHeader("X-Epistula-Attachment", "#sample entry#")
		content_text = "> " + strings.ReplaceAll(content_text, "\n", "\n> ")
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
		defer os.Remove(f.Name()) // clean up
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
	//if true { return }
	var message *gmime.Envelope
	done := false
	for !done {
		if EDITOR, err := exec.LookPath(EDITOR); err == nil {
			var procAttr os.ProcAttr
			procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
			if proc, err := os.StartProcess(EDITOR, []string{EDITOR, "+", tempfilename}, &procAttr); err == nil {
				proc.Wait()
			}
		}
	// - parses the file via gmime
		message = parseFile(tempfilename)
		done = !strings.Contains(message.Header("X-Epistula-Status"), "not")
	}
	message.SetHeader("MIME-Version", "1.0")
	message.SetHeader("User-Agent", "Epistula")
	message.SetHeader("Content-Type", "text/plain; charset=utf-8")
	message.SetHeader("Content-Transfer-Encoding", "quoted-printable")
	message.RemoveHeader("X-Epistula-Status")
	message.RemoveHeader("X-Epistula-Comment")
	message.RemoveHeader("X-Epistula-Attachment")
	// TODO add attachment
	// - retreives the desired keys
	// - encrypts the file via gpgme
	// - sends the email
	if b, err := message.Export(); err != nil {
		panic(err)
	} else {
		buffer = b
	}
	_ = []string{"sendmail", "-oem", "-oi", "-t", }[1]
	//
	// because the exported (for edit) email includes all meta data the program can add
	// x-epistula-* meta data that "talks to the user", for example telling her
	// about missing public keys. it can contain as well a "i am not done" flag
	// the user has to change to flag "done"
	//
	// the composer should be able to reply on multiple emails.
	//
	// open:
	// - the "replied" flag must be set somewhere
	// - attachments
}

func parseFile(filename string) *gmime.Envelope {
	if fh, err := os.Open(filename); err != nil {
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

