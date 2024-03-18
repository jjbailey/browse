// main.go
// start here
//
// Copyright (c) 2024 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

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
		followFlag  = getopt.BoolLong("follow", 'f', "follow file")
		caseFlag    = getopt.BoolLong("ignore-case", 'i', "search ignores case")
		numberFlag  = getopt.BoolLong("numbers", 'n', "line numbers")
		patternStr  = getopt.StringLong("pattern", 'p', "", "search pattern")
		titleStr    = getopt.StringLong("title", 't', "", "page title")
		versionFlag = getopt.BoolLong("version", 'v', "print version number")
		helpFlag    = getopt.BoolLong("help", '?', "this message")
	)

	getopt.SetUsage(usageMessage)
	getopt.Parse()
	args := getopt.Args()
	argc := len(args)

	if *helpFlag {
		usageMessage()
		os.Exit(0)
	}

	if *versionFlag {
		brVersion()
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

			if len(br.title) > 0 {
				title = br.title
			} else {
				title = br.fileName
			}
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
	br.modeScrollDown = *followFlag
	br.ignoreCase = *caseFlag
	br.modeNumbers = *numberFlag

	if *patternStr != "" {
		br.pattern = *patternStr
	}
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
	moveCursor(br.dispHeight, 1, true)

	if err := recover(); err != nil {
		fmt.Printf("panic occurred: %v", err)
	}

	br.saneExit()
}

func brVersion() {
	fmt.Printf("browse: version %s\n", BR_VERSION)
}

func usageMessage() {
	fmt.Print("Usage: browse [-finv] [-p pattern] [-t title] [filename]\n")
	fmt.Print(" -f, --follow       follow file\n")
	fmt.Print(" -i, --ignore-case  search ignores case\n")
	fmt.Print(" -n, --numbers      line numbers\n")
	fmt.Print(" -p, --pattern      search pattern\n")
	fmt.Print(" -t, --title        page title\n")
	fmt.Print(" -v, --version      print version number\n")
	fmt.Print(" -?, --help         this message\n")
}

// vim: set ts=4 sw=4 noet:
