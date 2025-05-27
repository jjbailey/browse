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
	"path/filepath"
	"strings"
	"syscall"
)

func browseFile(br *browseObj, fileName, title string, fromStdin bool, reset bool) bool {
	// init

	targetFile := strings.TrimSuffix(fileName, "/")
	basename := filepath.Base(targetFile)

	stat, err := os.Stat(targetFile)
	if err != nil {
		br.timedMessage(fmt.Sprintf("stat error: %v", err), MSG_RED)
		return false
	}

	if stat.IsDir() {
		br.timedMessage(fmt.Sprintf("%s: is a directory", basename), MSG_RED)
		return false
	}

	fp, err := os.Open(targetFile)
	if err != nil {
		br.timedMessage(fmt.Sprintf("open error: %v", err), MSG_RED)
		return false
	}

	defer fp.Close()

	if reset {
		resetState(br)
	}

	br.fileInit(fp, targetFile, title, fromStdin)

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

	return true
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
	browseFile(br, fpStdin.Name(), setTitle(br.title, "          "), true, false)
}

func processFileList(br *browseObj, args []string) {
	if len(args) == 0 {
		browseFile(br, br.fileName, setTitle(br.title, br.fileName), false, false)
		return
	}

	for index, fileName := range args {
		title := br.title

		if index != 0 || title == "" {
			title = fileName
		}

		browseFile(br, fileName, setTitle(title, fileName), false, false)

		if br.exit {
			break
		}

		if index < len(args)-1 {
			resetState(br)
		}
	}
}

// vim: set ts=4 sw=4 noet:
