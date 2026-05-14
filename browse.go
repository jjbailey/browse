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

// processPipeInput handles input piped on stdin into a temporary file for browsing.
func processPipeInput(br *browseObj) {
	fpStdin, err := os.CreateTemp("", "browse")
	if err != nil {
		errorExit(fmt.Errorf("error creating temporary file: %v", err))
		return
	}
	defer os.Remove(fpStdin.Name())

	// Goroutine owns fpStdin and closes it when stdin is exhausted
	go func() {
		br.readStdin(os.Stdin, fpStdin)
		fpStdin.Close()
	}()

	// Fast path for stdin: open temp file ourselves and pass to browseFile
	fp, err := os.Open(fpStdin.Name())
	if err != nil {
		errorExit(fmt.Errorf("cannot open temporary browse file: %v", err))
		return
	}
	defer fp.Close()

	// Save arg list
	br.currentList = []string{fpStdin.Name()}
	br.listAtStart = true
	browseFile(br, fp, fpStdin.Name(), "          ", true)
}

// processFileList iterates through a list of files and opens them for browsing.
func processFileList(br *browseObj, args []string, toplevel bool) bool {
	if len(args) == 0 {
		abs, err := filepath.Abs(br.fileName)
		if err != nil {
			abs = br.fileName
		}

		for {
			fp, err := validateAndOpenFile(br, abs)
			if err != nil {
				return false
			}

			br.currentList = []string{abs}
			br.listAtStart = true
			br.absFileName = abs
			browseFile(br, fp, br.absFileName, setTitle(br.title, abs), false)
			fp.Close()

			switch br.listAction {
			case LIST_ACTION_REWIND:
				br.listAction = LIST_ACTION_NONE
				resetState(br)

			case LIST_ACTION_RESUME:
				br.listAction = LIST_ACTION_NONE
				restoreResumeState(br)

			case LIST_ACTION_EXIT_ALL:
				return true

			default:
				return true
			}
		}
	}

	savedList := br.currentList
	savedListAtStart := br.listAtStart
	defer func() {
		br.currentList = savedList
		br.listAtStart = savedListAtStart
	}()

	// Build absolute and symlink-resolved paths against the starting cwd
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

	br.currentList = args
	br.listAtStart = true
	lastIdx := len(args) - 1
	openedAny := false

	for i := 0; i < len(args); i++ {
		fileName := args[i]
		fp, err := validateAndOpenFile(br, absArgs[i])
		if err != nil {
			continue
		}

		if !toplevel && !openedAny {
			resetState(br)
		}

		// Save for browserc
		br.absFileName = absArgs[i]
		br.currentList = args[i:]
		br.listAtStart = i == 0
		openedAny = true
		browseFile(br, fp, absArgs[i], fileName, false)
		fp.Close()

		if br.listAction == LIST_ACTION_REWIND {
			br.listAction = LIST_ACTION_NONE
			resetState(br)
			i = -1
			continue
		}

		if br.listAction == LIST_ACTION_RESUME {
			br.listAction = LIST_ACTION_NONE
			restoreResumeState(br)
			i--
			continue
		}

		if br.listAction == LIST_ACTION_EXIT_ALL {
			break
		}

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

	return openedAny
}

// browseFile initializes browsing state for a file and begins processing.
func browseFile(br *browseObj, fp *os.File, fileName, title string, fromStdin bool) {
	targetFile := strings.TrimSuffix(fileName, "/")

	checkBinaryFile(br, fp, targetFile)
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

	if !stat.Mode().IsRegular() {
		fp.Close()
		br.userAnyKey(fmt.Sprintf("%s %s: not a regular file ... [press any key] %s",
			MSG_RED, lastNChars(targetFile, br.dispWidth), VIDOFF))
		return nil, fmt.Errorf("%s: not a regular file", targetFile)
	}

	return fp, nil
}

// checkBinaryFile warns if the target file appears to be binary.
func checkBinaryFile(br *browseObj, fp *os.File, targetFile string) {
	if isBinaryFileFp(fp) {
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

// restoreResumeState restores the parent file's viewport after a nested list.
func restoreResumeState(br *browseObj) {
	br.fileName = br.resume.fileName
	br.absFileName = br.resume.absFileName
	br.title = br.resume.title
	br.fromStdin = br.resume.fromStdin
	br.firstRow = br.resume.firstRow
	br.lastRow = br.resume.lastRow
	br.shiftWidth = br.resume.shiftWidth
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
