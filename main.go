// main.go
// start here
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/term"
)

func main() {
	var br browseObj
	var tty *os.File
	var fileName string
	var screenName string

	argc := len(os.Args[0:])

	// do this now
	ttySaveTerm()
	syscall.Umask(077)

	if term.IsTerminal(int(os.Stdin.Fd())) {
		if argc < 2 {
			if !readRcFile(&br) {
				fmt.Printf("Usage: browse [filename]\n")
				os.Exit(1)
			}

			// we have some defaults
			fileName = br.fileName
			screenName = br.fileName
		} else {
			// use file given
			fileName = os.Args[1]
			screenName = os.Args[1]
		}

		// open file for reading
		fp, err := os.Open(fileName)
		errorExit(err)
		br.fileInit(fp, fileName, false)
	} else {
		// create temp file for writing
		tmpfp, err := os.CreateTemp("", "browse")
		errorExit(err)

		// open temp file for reading
		screenName = "          "
		fileName = tmpfp.Name()
		fp, err := os.Open(fileName)
		errorExit(err)
		br.fileInit(fp, fileName, true)

		// copy stdin to temp file
		go readStdin(os.Stdin, tmpfp)
	}

	tty, _ = os.Open("/dev/tty")
	br.screenInit(tty, screenName)

	// error recovery, graceful exit

	defer func() {
		if err := recover(); err != nil {
			movecursor(br.dispHeight, 1, true)
			fmt.Printf("panic occurred: %v", err)
		}

		br.saneExit()
	}()

	// signals
	br.catchSignals()

	// start a file reader
	syncOK := make(chan bool)
	go readFile(&br, syncOK)
	readerOK := <-syncOK
	close(syncOK)

	if readerOK {
		// go
		commands(&br)
	}

	// done
	br.saneExit()
}

// vim: set ts=4 sw=4 noet:
