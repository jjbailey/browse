// main.go
//
// browse is a terminal-based file and stream viewer with support for
// pattern searching, line numbering, and follow mode.
//
// It can read input from files or standard input, making it suitable
// for use as a pager or as part of a pipeline.
//
// Copyright (c) 2024-2026 jjb
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

// main parses flags and launches the browse session.
func main() {
	var br browseObj
	var tty *os.File
	var fromStdin bool

	// define command line flags

	followFlag := getopt.BoolLong("follow", 'f', "follow file")
	tailFlag := getopt.BoolLong("tail", 'F', "fast follow")
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

	// subtle precedence
	if *tailFlag {
		br.modeScroll = MODE_SCROLL_TAIL
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
	defer tty.Close()
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

// brVersion prints the current version string.
func brVersion() {
	fmt.Printf("browse: version %s\n", BR_VERSION)
}

// usageMessage prints CLI usage information.
func usageMessage(arg0 string) {
	fmt.Printf("Usage: %s [-fFinv] [-p pattern] [-t title] [filename...]\n",
		filepath.Base(arg0))
	fmt.Print("  -f, --follow       follow file\n")
	fmt.Print("  -F, --tail         fast follow\n")
	fmt.Print("  -i, --ignore-case  search ignores case\n")
	fmt.Print("  -n, --numbers      line numbers\n")
	fmt.Print("  -p, --pattern      search pattern\n")
	fmt.Print("  -t, --title        page title\n")
	fmt.Print("  -v, --version      print version number\n")
	fmt.Print("  -?, --help         this message\n")
}

// vim: set ts=4 sw=4 noet:
