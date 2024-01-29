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

	"github.com/pborman/getopt/v2"
	"golang.org/x/term"
)

func main() {
	var br browseObj
	var tty *os.File
	var fileName, title string

	var (
		followFlag = getopt.BoolLong("follow", 'f', "follow file")
		numberFlag = getopt.BoolLong("numbers", 'n', "line numbers")
		helpFlag   = getopt.BoolLong("help", '?', "this message")
		titleStr   = getopt.StringLong("title", 't', "", "page title")
	)

	getopt.SetUsage(usageMessage)
	getopt.Parse()
	args := getopt.Args()
	argc := len(args)

	if *helpFlag {
		usageMessage()
		os.Exit(0)
	}

	// do this now
	ttySaveTerm()
	syscall.Umask(077)

	if term.IsTerminal(int(os.Stdin.Fd())) {
		if argc == 0 {
			if !readRcFile(&br) {
				usageMessage()
				os.Exit(1)
			}

			// we have some defaults
			fileName = br.fileName
			title = br.fileName
		} else {
			// use file given
			fileName = args[0]
			title = args[0]
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
		title = "          "
		fileName = tmpfp.Name()
		fp, err := os.Open(fileName)
		errorExit(err)
		br.fileInit(fp, fileName, true)

		// copy stdin to temp file
		go br.readStdin(os.Stdin, tmpfp)
	}

	tty, _ = os.Open("/dev/tty")
	br.screenInit(tty, title)

	// error recovery, graceful exit
	defer handlePanic(&br)

	// signals
	br.catchSignals()

	// set options from commandline
	br.modeNumbers = *numberFlag
	br.modeScrollDown = *followFlag
	if *titleStr != "" {
		br.title = *titleStr
	}

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

func handlePanic(br *browseObj) {
	movecursor(br.dispHeight, 1, true)

	if err := recover(); err != nil {
		fmt.Printf("panic occurred: %v", err)
	}

	br.saneExit()
}

func usageMessage() {
	fmt.Print("Usage: browse [-fn] [-t title] [filename]\n")
	fmt.Print(" -f, --follow   follow file\n")
	fmt.Print(" -n, --numbers  line numbers\n")
	fmt.Print(" -t, --title    page title\n")
	fmt.Print(" -?, --help     this message\n")
}

// vim: set ts=4 sw=4 noet:
