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
	"path/filepath"

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

	getopt.SetUsage(func() { usageMessage(os.Args[0]) })
	getopt.Parse()
	args := getopt.Args()
	argc := len(args)

	if *helpFlag {
		usageMessage(os.Args[0])
		os.Exit(0)
	}

	if *versionFlag {
		brVersion()
		os.Exit(0)
	}

	preInitialization()

	fromStdin = !term.IsTerminal(int(os.Stdin.Fd()))
	if !fromStdin {
		if argc == 0 {
			if !br.readRcFile() {
				usageMessage(os.Args[0])
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
		br.initTitle = *titleStr
	}

	// init tty and signals

	var err error
	tty, err = os.Open("/dev/tty")
	if err != nil {
		fmt.Fprintf(os.Stderr, "browse: %v\n", err)
		os.Exit(1)
	}
	br.screenInit(tty)
	br.catchSignals()

	if fromStdin {
		processPipeInput(&br)
	} else {
		processFileList(&br, args, true)
	}

	// done
	br.saneExit()
}

func brVersion() {
	fmt.Printf("browse: version %s\n", BR_VERSION)
}

func usageMessage(arg0 string) {
	fmt.Printf("Usage: %s [-finv] [-p pattern] [-t title] [filename...]\n",
		filepath.Base(arg0))
	fmt.Print("  -f, --follow       follow file\n")
	fmt.Print("  -i, --ignore-case  search ignores case\n")
	fmt.Print("  -n, --numbers      line numbers\n")
	fmt.Print("  -p, --pattern      search pattern\n")
	fmt.Print("  -t, --title        page title\n")
	fmt.Print("  -v, --version      print version number\n")
	fmt.Print("  -?, --help         this message\n")
}

// vim: set ts=4 sw=4 noet:
