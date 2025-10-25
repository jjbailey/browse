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
	targetFile := strings.TrimSuffix(fileName, "/")

	// Validate and open the file
	fp, err := validateAndOpenFile(targetFile, br)
	if err != nil {
		return false
	}
	defer fp.Close()

	// Check if file is binary and warn user
	checkBinaryFile(br, targetFile)

	// Reset browser state if requested
	if reset {
		resetState(br)
	}

	br.fileInit(fp, targetFile, title, fromStdin)
	updateFileHistory(targetFile, br)

	return processFileBrowsing(br)
}

func validateAndOpenFile(targetFile string, br *browseObj) (*os.File, error) {
	// Check if file exists and get file info

	stat, err := os.Stat(targetFile)
	if err != nil {
		br.timedMessage(fmt.Sprintf("stat error: %v", err), MSG_RED)
		return nil, err
	}

	// Ensure it's not a directory
	if stat.IsDir() {
		br.timedMessage(fmt.Sprintf("%s: is a directory", filepath.Base(targetFile)), MSG_RED)
		return nil, fmt.Errorf("file is a directory")
	}

	// Open the file
	fp, err := os.Open(targetFile)
	if err != nil {
		br.timedMessage(fmt.Sprintf("open error: %v", err), MSG_RED)
		return nil, err
	}

	return fp, nil
}

func checkBinaryFile(br *browseObj, targetFile string) {
	if isBinaryFile(targetFile) {
		br.timedMessage(fmt.Sprintf("%s: is a binary file", filepath.Base(targetFile)), MSG_ORANGE)
	}
}

func updateFileHistory(targetFile string, br *browseObj) {
	if !br.fromStdin && len(targetFile) > 0 {
		history := loadHistory(fileHistory)
		// Use resolved symlink path for consistent history entries
		history = append(history, resolveSymlink(targetFile))
		saveHistory(history, fileHistory)
	}
}

func processFileBrowsing(br *browseObj) bool {
	// Start file reading in background
	syncOK := make(chan bool, 1)
	go readFile(br, syncOK)

	// Wait for reader to be ready and process commands
	if readerOK := <-syncOK; readerOK {
		commands(br)
	}

	// Save session state if requested
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

	go func() {
		empty := br.readStdin(os.Stdin, fpStdin)
		if empty {
			br.saneExit()
		}
	}()

	browseFile(br, fpStdin.Name(), setTitle(br.title, "          "), true, false)
}

func processFileList(br *browseObj, args []string, toplevel bool) {
	if len(args) == 0 {
		// handles file from browserc
		browseFile(br, br.fileName, setTitle(br.title, br.fileName), false, false)
		return
	}

	for index, fileName := range args {
		// handles list of files
		browseFile(br, fileName, setTitle(fileName, fileName), false, false)

		if br.exit {
			if toplevel {
				return
			} else {
				br.exit = false
				return
			}
		}

		if index < len(args)-1 {
			resetState(br)
		}
	}
}

// vim: set ts=4 sw=4 noet:
