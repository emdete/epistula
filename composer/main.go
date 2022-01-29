package main

import (
	"log"
	"os"
	"io"
	"strings"
	"os/exec"
	"io/ioutil"
	"bufio"
	//
	"github.com/sendgrid/go-gmime/gmime"
	"github.com/proglottis/gpgme"
)

const (
	CRLF = "\r\n"
)

func main() {
	log.SetPrefix("email ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC | log.Lshortfile)
	log.SetOutput(os.Stderr)
	log.Printf("main %#v", os.Args)
	// The Idea is as follows: the composeser
	// - is called with all information in its arguments like --from, --reply, --to, --subject, --cc, --bcc, ...
	from := "M. Dietrich <mdt@emdete.de>"
	var envelope *gmime.Envelope
	switch len(os.Args) {
	case 1:
		// we compose a new email
		envelope = nil
	case 2:
		// we reply to the given email
		envelope = parseFile(os.Args[1]) // TODO for now it excepts a name to reply to only
		if strings.HasPrefix(envelope.Header("Content-Type"), "multipart/encrypted; ") {
		}
	default:
		panic("wrongs arguments")
	}
	// - decrypt reply email
	// - composes an email via gmime
	var buffer []byte
	if message, err := gmime.Parse(
		"Date: Fri, 28 Jan 2022 23:59:04 +0100" + CRLF +
		"From: " + from + CRLF +
		"MIME-Version: 1.0" + CRLF +
		"Content-Type: text/plain; charset=utf-8" + CRLF +
		"Content-Transfer-Encoding: quoted-printable" + CRLF +
		CRLF +
		CRLF); err != nil {
		panic(err)
	} else {
		// fields: from sender reply-to to cc bcc
		//message.AddAddress("From", "M. Dietrich", "mdt@emdete.de")
		//message.SetHeader("Content-Type", "text/plain")
		if envelope != nil {
			message.ParseAndAppendAddresses("To", envelope.Header("From"))
		} else {
			message.ParseAndAppendAddresses("To", "") // TODO how to add an empty "To:" ?
		}
		for _,name := range []string{"To", "Cc", "Bcc", } {
			if envelope != nil {
				message.ParseAndAppendAddresses(name, envelope.Header(name)) // TODO remove myself
			} else {
				message.ParseAndAppendAddresses(name, " ") // TODO how to add an empty "...:" ?
			}
		}
		if envelope != nil {
			message.SetSubject(envelope.Subject())
		} else {
			message.SetSubject(" ")
		}
		// Content-ID
		// Content-Transfer-Encoding: quoted-printable
		// Date: Thu, 13 Dec 2018 14:19:38 +0000
		// In-Reply-To
		// MIME-Version
		// Message-ID
		// References
		// Return-Path
		// Thread-Topic
		message.SetHeader("User-Agent", "Epistula")
		message.SetHeader("X-Epistula-State", "I am not done")
		message.SetHeader("X-Epistula-Comment", "This is your MUA talking to you. Add attachments as headerfield like below. Dont destroy the mail structure, if the outcome cant be parsed you will thrown into your editor again to fix it. Change the State to not contain 'not'.")
		message.SetHeader("X-Epistula-Attachment", "#sample entry#")
		if envelope != nil {
			text := ""
			if err := envelope.Walk(func (part *gmime.Part) error {
				if part.IsText() && part.ContentType() == "text/plain" {
					text = part.Text()
				}
				return nil
			}); err != nil {
				panic(err)
			}
			text = strings.ReplaceAll(text, "\n", "\n> ")
			if err := message.Walk(func (part *gmime.Part) error {
				if part.IsText() && part.ContentType() == "text/plain" {
					part.SetText(text)
				}
				return nil
			}); err != nil {
				panic(err)
			}
		}
		if b, err := message.Export(); err != nil {
			panic(err)
		} else {
			buffer = b
		}
	}
	// - exports it to a temp file
	var tempfilename string
	if f, err := os.CreateTemp("", "example"); err != nil {
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
	programname := "nvim"
	if programname, err := exec.LookPath(programname); err == nil {
		var procAttr os.ProcAttr
		procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
		if proc, err := os.StartProcess(programname, []string{programname, tempfilename}, &procAttr); err == nil {
			proc.Wait()
		}
	}
	// - parses the file via gmime
	message := parseFile(tempfilename)
	strings.Contains(message.Header("X-Epistula-Attachment"), "not")
	// - retreives the desired keys
	// - encrypts the file via gpgme
	// - sends the email
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

func decryptMessage(stream io.Reader) *gmime.Envelope {
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
	return nil
}

