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

	// Fast path for stdin: open temp file ourselves and pass to browseFile
	fp, err := os.Open(fpStdin.Name())
	if err != nil {
		errorExit(fmt.Errorf("cannot open temporary browse file: %v", err))
		return
	}
	defer fp.Close()

	browseFile(br, fp, fpStdin.Name(), "          ", true)
}

func processFileList(br *browseObj, args []string, toplevel bool) {
	if len(args) == 0 {
		// Handles file from browserc
		fp, err := validateAndOpenFile(br, br.fileName)
		if err != nil {
			return
		}
		defer fp.Close()

		// Save for browserc
		br.absFileName = br.fileName

		browseFile(br, fp, br.absFileName, br.fileName, false)
		return
	}

	// Build absolute and symlink-resolved paths
	absArgs := make([]string, len(args))
	for i, fileName := range args {
		abs, err := filepath.Abs(fileName)
		if err != nil {
			abs = fileName
		}

		abs, err = resolveSymlink(abs)
		if err != nil {
			abs = fileName
		}

		absArgs[i] = abs
	}

	lastIdx := len(args) - 1
	for i, fileName := range args {
		fp, err := validateAndOpenFile(br, absArgs[i])
		if err != nil {
			continue
		}
		func() {
			// Ensure close happens per file
			defer fp.Close()

			// Save for browserc
			br.absFileName = absArgs[i]

			if br.title == "" {
				br.title = fileName
			}

			browseFile(br, fp, br.absFileName, br.title, false)

			if i != lastIdx {
				resetState(br)
			}
		}()

		if br.exit {
			if !toplevel {
				br.exit = false
			}

			break
		}
	}
}

func browseFile(br *browseObj, fp *os.File, fileName, title string, fromStdin bool) {
	targetFile := strings.TrimSuffix(fileName, "/")

	checkBinaryFile(br, targetFile)
	br.fileInit(fp, targetFile, title, fromStdin)

	updateFileHistory(br, targetFile)
	updateSearchHistory(br.pattern)

	processFileBrowsing(br)
}

func validateAndOpenFile(br *browseObj, targetFile string) (*os.File, error) {
	stat, err := os.Stat(targetFile)
	if err != nil {
		br.userAnyKey(fmt.Sprintf("%s %s: cannot open ... [press enter] %s",
			MSG_RED, filepath.Base(targetFile), VIDOFF))
		return nil, err
	}

	if stat.IsDir() {
		br.userAnyKey(fmt.Sprintf("%s %s: is a directory ... [press enter] %s",
			MSG_RED, filepath.Base(targetFile), VIDOFF))
		return nil, fmt.Errorf("file is a directory")
	}

	fp, err := os.Open(targetFile)
	if err != nil {
		br.userAnyKey(fmt.Sprintf("%s %s: cannot open ... [press enter] %s",
			MSG_RED, filepath.Base(targetFile), VIDOFF))
		return nil, err
	}

	return fp, nil
}

func checkBinaryFile(br *browseObj, targetFile string) {
	if isBinaryFile(targetFile) {
		br.timedMessage(fmt.Sprintf("%s: is a binary file", filepath.Base(targetFile)), MSG_ORANGE)
	}
}

func updateFileHistory(br *browseObj, targetFile string) {
	if !br.fromStdin && len(targetFile) > 0 {
		history := append(loadHistory(fileHistory), targetFile)
		saveHistory(history, fileHistory)
	}
}

func processFileBrowsing(br *browseObj) {
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

	// Reset the title
	br.title = ""
}

func resetState(br *browseObj) {
	br.firstRow = 0
	br.lastRow = 0
	br.shiftWidth = 0
	br.modeScroll = MODE_SCROLL_NONE
}

func preInitialization(br *browseObj) {
	ttySaveTerm()
	syscall.Umask(077)
	br.browseInit()
}

// vim: set ts=4 sw=4 noet:
