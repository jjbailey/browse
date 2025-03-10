// browse.go
// file and rcfile handling functions
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
)

func browseFile(br *browseObj, fileName, title string, fromStdin bool) {
	// init

	fp, err := os.Open(fileName)

	if err != nil {
		br.timedMessage(fmt.Sprintf("error opening file: %v", err), MSG_RED)
		return
	}

	defer fp.Close()

	br.fileInit(fp, fileName, title, fromStdin)

	// start a reader

	syncOK := make(chan bool, 1)
	go readFile(br, syncOK)

	// process commands

	if readerOK := <-syncOK; readerOK {
		commands(br)
	}

	if !br.fromStdin && br.saveRC {
		br.writeRcFile()
	}
}

func resetState(br *browseObj) {
	br.firstRow = 0
	br.lastRow = 0
	br.shiftWidth = 0
	br.modeScroll = MODE_SCROLL_NONE
}

func setTitle(primary, fallback string) string {
	if primary != "" {
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

	if err != nil {
		errorExit(fmt.Errorf("error creating temporary file: %v", err))
		return
	}

	defer os.Remove(fpStdin.Name())
	defer fpStdin.Close()

	go br.readStdin(os.Stdin, fpStdin)
	browseFile(br, fpStdin.Name(), setTitle(br.title, "          "), true)
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
