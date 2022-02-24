package main

import (
	// "bufio"
	// "fmt"
	// "io"
	// "io/ioutil"
	"log"
	"os"
	// "os/exec"
	// "os/signal"
	// "regexp"
	// "sort"
	// "strconv"
	// "strings"
	// "syscall"
	// "time"
	// "path"
	// "github.com/emdete/go-gmime/gmime"
	// "github.com/zenhack/go.notmuch"
	// "github.com/proglottis/gpgme"
	)

func main() {
	log.SetPrefix("epistula ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC | log.Lshortfile)
	log.SetOutput(os.Stderr)
	if err := Test(); err != nil {
		panic(err)
	}
}

