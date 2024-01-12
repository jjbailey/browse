// main.go
// start here
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"fmt"
	"os"
	"os/signal"
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
			saneExit(&br)
		}
	}()

	// signals

	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGABRT, syscall.SIGTERM, syscall.SIGWINCH)
	go func() {
		for {
			sig := <-sigChan

			switch sig {

			case syscall.SIGABRT:
			case syscall.SIGTERM:
				saneExit(&br)

			case syscall.SIGWINCH:
				resizeWindow(&br)
			}
		}
	}()

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
	saneExit(&br)
}

func resizeWindow(br *browseObj) {
	// catch SIGWINCH for window size changes

	br.dispWidth, br.dispHeight, _ = term.GetSize(int(br.tty.Fd()))
	br.dispRows = br.dispHeight - 1
	br.lastMatch = SEARCH_RESET

	br.pageHeader()
	br.pageCurrent()

	if br.modeTail {
		fmt.Printf("%s", CURRESTORE)
	}
}

func saneExit(br *browseObj) {
	// clean up

	ttyRestore()
	resetScrRegion()
	fmt.Printf("%s", LINEWRAPON)
	movecursor(br.dispHeight, 1, true)

	if br.fromStdin {
		os.Remove(br.fileName)
	}

	if !br.fromStdin && br.saveRC {
		writeRcFile(br)
	}

	os.Exit(0)
}

// vim: set ts=4 sw=4 noet:
