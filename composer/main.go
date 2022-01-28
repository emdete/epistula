package main

import (
	"log"
	"os"
)

func main() {
	log.SetPrefix("email ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC | log.Lshortfile)
	log.SetOutput(os.Stderr)
	log.Printf("main")
	// The Idea is as follows: the composeser
	// - is called with all information in its arguments like --to, --subject, --cc, ...
	// - composes an email via gmime
	// - exports it to a temp file
	// - execs the editor and waits for its termination
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

