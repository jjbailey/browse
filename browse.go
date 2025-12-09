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

var CurrentList []string

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

	// Save arg list
	CurrentList = []string{fpStdin.Name()}
	browseFile(br, fp, fpStdin.Name(), "          ", true)
}

func processFileList(br *browseObj, args []string, toplevel bool) {
	if len(args) == 0 {
		// Handles file from browserc
		abs, err := filepath.Abs(br.fileName)
		if err != nil {
			abs = br.fileName
		}

		fp, err := validateAndOpenFile(br, abs)
		if err != nil {
			return
		}

		// Save arg list
		CurrentList = []string{abs}

		// Save for browserc
		br.absFileName = abs

		browseFile(br, fp, br.absFileName, setTitle(br.title, abs), false)
		fp.Close()
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

		// Save arg list
		CurrentList = args[i:]

		// Use a closure to ensure fp.Close() is called after each file.
		func(fp *os.File, absPath, fileName string) {
			defer fp.Close()

			// Save for browserc
			br.absFileName = absPath

			browseFile(br, fp, absPath, fileName, false)

		}(fp, absArgs[i], fileName)

		if i != lastIdx {
			resetState(br)
		}

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

	if !br.fromStdin {
		updateHistory(targetFile, fileHistory)
	}

	processFileBrowsing(br)
}

func validateAndOpenFile(br *browseObj, targetFile string) (*os.File, error) {
	stat, err := os.Stat(targetFile)
	if err != nil {
		br.userAnyKey(fmt.Sprintf("%s %s: cannot open ... [press any key] %s",
			MSG_RED, lastNChars(targetFile, br.dispWidth), VIDOFF))
		return nil, err
	}

	if stat.IsDir() {
		br.userAnyKey(fmt.Sprintf("%s %s: is a directory ... [press any key] %s",
			MSG_RED, lastNChars(targetFile, br.dispWidth), VIDOFF))
		return nil, fmt.Errorf("%s: is a directory", targetFile)
	}

	fp, err := os.Open(targetFile)
	if err != nil {
		br.userAnyKey(fmt.Sprintf("%s %s: cannot open ... [press any key] %s",
			MSG_RED, lastNChars(targetFile, br.dispWidth), VIDOFF))
		return nil, err
	}

	return fp, nil
}

func checkBinaryFile(br *browseObj, targetFile string) {
	if isBinaryFile(targetFile) {
		br.timedMessage(fmt.Sprintf("%s: is a binary file", filepath.Base(targetFile)), MSG_ORANGE)
	}
}

func processFileBrowsing(br *browseObj) {
	// Start file reading in background
	syncOK := make(chan bool, 1)
	go readFile(br, syncOK)

	// Wait for reader to be ready and process commands
	readerOK := <-syncOK
	if !readerOK {
		return
	}

	commands(br)

	// Save session state if requested
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

func preInitialization(br *browseObj) {
	setupBrDir()
	ttySaveTerm()
	syscall.Umask(077)
	br.browseInit()
}

func setTitle(primary, fallback string) string {
	if primary != "" {
		return primary
	}

	return fallback
}

// vim: set ts=4 sw=4 noet:
