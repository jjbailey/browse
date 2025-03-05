// main.go
// start here
//
// Copyright (c) 2024-2025 jjb
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
	var fromStdin bool

	// define command line flags

	followFlag := getopt.BoolLong("follow", 'f', "follow file")
	caseFlag := getopt.BoolLong("ignore-case", 'i', "search ignores case")
	numberFlag := getopt.BoolLong("numbers", 'n', "line numbers")
	patternStr := getopt.StringLong("pattern", 'p', "", "search pattern")
	titleStr := getopt.StringLong("title", 't', "", "page title")
	versionFlag := getopt.BoolLong("version", 'v', "print version number")
	helpFlag := getopt.BoolLong("help", '?', "this message")

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

	preInitialization(&br)

	if fromStdin = !term.IsTerminal(int(os.Stdin.Fd())); !fromStdin {
		if argc == 0 {
			if !br.readRcFile() {
				usageMessage()
				os.Exit(1)
			}
		}
	}

	// set options from command line

	if *followFlag {
		br.modeScroll = MODE_SCROLL_FOLLOW
	}

	br.ignoreCase = *caseFlag
	br.modeNumbers = *numberFlag

	if len(*patternStr) > 0 {
		br.pattern = *patternStr
	}

	if len(*titleStr) > 0 {
		br.title = *titleStr
	}

	// init tty and signals

	tty, _ = os.Open("/dev/tty")
	br.screenInit(tty)
	br.catchSignals()

	if fromStdin {
		processPipeInput(&br)
	} else {
		processFileList(&br, argc, args)
	}

	// done
	br.saneExit()
}

func brVersion() {
	fmt.Printf("browse: version %s\n", BR_VERSION)
}

func usageMessage() {
	fmt.Print("Usage: browse [-finv] [-p pattern] [-t title] [filename...]\n")
	fmt.Print(" -f, --follow       follow file\n")
	fmt.Print(" -i, --ignore-case  search ignores case\n")
	fmt.Print(" -n, --numbers      line numbers\n")
	fmt.Print(" -p, --pattern      search pattern\n")
	fmt.Print(" -t, --title        page title\n")
	fmt.Print(" -v, --version      print version number\n")
	fmt.Print(" -?, --help         this message\n")
}

func browseFile(br *browseObj, fileName, title string, fromStdin bool) {
	// init

	fp, err := os.Open(fileName)

	if err != nil {
		br.timedMessage(fmt.Sprintf("%v", err), MSG_RED)
		return
	}

	br.fileInit(fp, fileName, title, fromStdin)

	// start a reader
	syncOK := make(chan bool)
	go readFile(br, syncOK)
	readerOK := <-syncOK
	close(syncOK)

	// process commands

	if readerOK {
		commands(br)
	}

	if !br.fromStdin && br.saveRC {
		br.writeRcFile()
	}

	fp.Close()
}

func resetState(br *browseObj) {
	br.firstRow = 0
	br.lastRow = 0
	br.shiftWidth = 0
	br.modeScroll = MODE_SCROLL_NONE
}

func setTitle(primary, fallback string) string {
	if len(primary) > 0 {
		return primary
	}

	return fallback
}

func preInitialization(br *browseObj) {
	ttySaveTerm()
	syscall.Umask(077)
	br.browseInit()
}

func processPipeInput(br *browseObj) {
	fpStdin, err := os.CreateTemp("", "browse")
	errorExit(err)

	go br.readStdin(os.Stdin, fpStdin)
	browseFile(br, fpStdin.Name(), setTitle(br.title, "          "), true)
	fpStdin.Close()
	os.Remove(fpStdin.Name())
}

func processFileList(br *browseObj, argc int, args []string) {
	if argc == 0 {
		browseFile(br, br.fileName, setTitle(br.title, br.fileName), false)
	} else {
		for index, fileName := range args {
			browseFile(br, fileName, setTitle(fileName, fileName), false)

			if br.exit {
				break
			}

			if index < len(args)-1 {
				resetState(br)
			}
		}
	}
}

// vim: set ts=4 sw=4 noet:
