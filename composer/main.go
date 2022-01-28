package main

import (
	"log"
	"os"
	"os/exec"
	"io/ioutil"
	"bufio"
	//
	"github.com/sendgrid/go-gmime/gmime"
)

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

func main() {
	log.SetPrefix("email ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC | log.Lshortfile)
	log.SetOutput(os.Stderr)
	log.Printf("main %#v", os.Args)
	// The Idea is as follows: the composeser
	// - is called with all information in its arguments like --to, --subject, --cc, ...
	var envelope *gmime.Envelope
	switch len(os.Args) {
	case 1:
		// we compose a new email
		envelope = nil
	case 2:
		// we reply to the given email
		envelope = parseFile(os.Args[1])
	default:
		panic("wrongs arguments")
	}
	// - composes an email via gmime
	// - exports it to a temp file
	// - execs the editor and waits for its termination
	if envelope != nil {
		envelope.Subject()
	}
	programname := "nvim"
	if programname, err := exec.LookPath(programname); err == nil {
		var procAttr os.ProcAttr
		procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
		if proc, err := os.StartProcess(programname, []string{programname, "/tmp/epistula.new.mail.0000"}, &procAttr); err == nil {
			proc.Wait()
		}
	}
	// - parses the file via gmimg
	// - retreives the desired keys
	// - encrypts the file via gpgme
	// - sends the email
	// because the exported email includes all meta data the program can add
	// x-epistula-* meta data that "talks to the user", for example telling her
	// about missing public keys. it can contain as well a "i am not done" flag
	// the user has to change to flag "done"
	//
	// the composer should be able to reply on multiple emails.
	//
	// open:
	// - the "replied" flag must be set somewhere
}

