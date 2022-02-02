package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	// see ~/go/pkg/mod/github.com/proglottis/gpgme@v0.1.1
	"github.com/proglottis/gpgme"
	// see ~/go/pkg/mod/github.com/arran4/golang-ical@v0.0.0-20220115055431-e3ae8290e7b8/
	"github.com/arran4/golang-ical"
	// see ~/go/pkg/mod/github.com/sendgrid/go-gmime@v0.0.0-20211124164648-4c44cbd981d8/
	"github.com/sendgrid/go-gmime/gmime"
	// see ~/go/pkg/mod/github.com/zenhack/go.notmuch@v0.0.0-20211022191430-4d57e8ad2a8b/
	"github.com/zenhack/go.notmuch"
)

func parseMessage(filename string) *gmime.Envelope {
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

func ThreadHasTag(thread *notmuch.Thread, search string) bool {
	tags := thread.Tags()
	var tag *notmuch.Tag
	for tags.Next(&tag) {
		if tag.Value == search {
			return true
		}
	}
	return false
}

func ThreadAddTag(id, tag string) error {
	if db, err := notmuch.Open(NotMuchDatabasePath, notmuch.DBReadWrite); err != nil {
		return err
	} else {
		defer db.Close()
		query := db.NewQuery("thread:" + id)
		defer query.Close()
		if 1 != query.CountThreads() { return errors.New("not uniq") }
		if threads, err := query.Threads(); err != nil {
			return err
		} else {
			var thread *notmuch.Thread
			if threads.Next(&thread) {
				defer thread.Close()
			}
			messages := thread.Messages()
			var message *notmuch.Message
			for messages.Next(&message) {
				if err := message.AddTag(tag); err != nil { return err }
			}
			if threads.Next(&thread) { return errors.New("additional thread") }
		}
	}
	return nil
}


func ThreadRemoveTag(id, tag string) error {
	if db, err := notmuch.Open(NotMuchDatabasePath, notmuch.DBReadWrite); err != nil {
		return err
	} else {
		defer db.Close()
		query := db.NewQuery("thread:" + id)
		defer query.Close()
		if 1 != query.CountThreads() { return errors.New("not uniq") }
		if threads, err := query.Threads(); err != nil {
			return err
		} else {
			var thread *notmuch.Thread
			if threads.Next(&thread) {
				defer thread.Close()
			}
			messages := thread.Messages()
			var message *notmuch.Message
			for messages.Next(&message) {
				if err := message.RemoveTag(tag); err != nil { return err }
			}
			if threads.Next(&thread) { return errors.New("additional thread") }
		}
	}
	return nil
}


