// browse.go
// file and rcfile handling functions
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
	"strings"
	"syscall"
)

// CurrentList tracks the current list of files being browsed.
var CurrentList []string

// processPipeInput handles input piped on stdin into a temporary file for browsing.
func processPipeInput(br *browseObj) {
	fpStdin, err := os.CreateTemp("", "browse")
	if err != nil {
		errorExit(fmt.Errorf("error creating temporary file: %v", err))
		return
	}
	defer os.Remove(fpStdin.Name())
	defer fpStdin.Close()

	go func() {
		br.readStdin(os.Stdin, fpStdin)
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

// processFileList iterates through a list of files and opens them for browsing.
func processFileList(br *browseObj, args []string, toplevel bool) {
	if len(args) == 0 {
		abs, err := filepath.Abs(br.fileName)
		if err != nil {
			abs = br.fileName
		}

		fp, err := validateAndOpenFile(br, abs)
		if err != nil {
			return
		}
		defer fp.Close()

		CurrentList = []string{abs}
		br.absFileName = abs
		browseFile(br, fp, br.absFileName, setTitle(br.title, abs), false)
		return
	}

	savedList := CurrentList
	defer func() { CurrentList = savedList }()

	// Build absolute and symlink-resolved paths
	absArgs := make([]string, len(args))
	for i, fileName := range args {
		abs, err := filepath.Abs(fileName)
		if err != nil {
			abs = fileName
		}
		if resolved, err := resolveSymlink(abs); err == nil {
			abs = resolved
		}
		absArgs[i] = abs
	}

	CurrentList = args
	lastIdx := len(args) - 1

	for i, fileName := range args {
		fp, err := validateAndOpenFile(br, absArgs[i])
		if err != nil {
			continue
		}

		// Save for browserc
		br.absFileName = absArgs[i]
		CurrentList = args[i:]
		browseFile(br, fp, absArgs[i], fileName, false)
		fp.Close()

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

// browseFile initializes browsing state for a file and begins processing.
func browseFile(br *browseObj, fp *os.File, fileName, title string, fromStdin bool) {
	targetFile := strings.TrimSuffix(fileName, "/")

	checkBinaryFile(br, targetFile)
	br.fileInit(fp, targetFile, title, fromStdin)

	if !br.fromStdin {
		updateHistory(targetFile, fileHistory)
	}

	processFileBrowsing(br)
}

// validateAndOpenFile opens a file and validates it is suitable for browsing.
func validateAndOpenFile(br *browseObj, targetFile string) (*os.File, error) {
	fp, err := os.Open(targetFile)
	if err != nil {
		br.userAnyKey(fmt.Sprintf("%s %s: cannot open ... [press any key] %s",
			MSG_RED, lastNChars(targetFile, br.dispWidth), VIDOFF))
		return nil, err
	}

	stat, err := fp.Stat()
	if err != nil {
		fp.Close()
		br.userAnyKey(fmt.Sprintf("%s %s: cannot open ... [press any key] %s",
			MSG_RED, lastNChars(targetFile, br.dispWidth), VIDOFF))
		return nil, err
	}

	if stat.IsDir() {
		fp.Close()
		br.userAnyKey(fmt.Sprintf("%s %s: is a directory ... [press any key] %s",
			MSG_RED, lastNChars(targetFile, br.dispWidth), VIDOFF))
		return nil, fmt.Errorf("%s: is a directory", targetFile)
	}

	return fp, nil
}

// checkBinaryFile warns if the target file appears to be binary.
func checkBinaryFile(br *browseObj, targetFile string) {
	if isBinaryFile(targetFile) {
		br.timedMessage(fmt.Sprintf("%s: is a binary file", filepath.Base(targetFile)), MSG_ORANGE)
	}
}

// processFileBrowsing runs the read loop and command processing for a file.
func processFileBrowsing(br *browseObj) {
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

// resetState clears browsing state between file switches.
func resetState(br *browseObj) {
	br.firstRow = 0
	br.lastRow = 0
	br.shiftWidth = 0
	br.modeScroll = MODE_SCROLL_NONE
}

// preInitialization performs startup setup before browsing begins.
func preInitialization() {
	setupBrDir()
	ttySaveTerm()
	syscall.Umask(077)
}

// setTitle returns the primary title when set, otherwise the fallback.
func setTitle(primary, fallback string) string {
	if primary != "" {
		return primary
	}

	return fallback
}

// vim: set ts=4 sw=4 noet:
