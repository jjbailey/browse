// completer.go
// central completion module (bash + file)
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/chzyer/readline"
	"golang.org/x/term"
)

type completeType int

const (
	bashComplete completeType = iota
	fileComplete
)

type completer struct {
	completionType completeType
}

func (c *completer) Do(line []rune, pos int) ([][]rune, int) {
	var candidates [][]rune
	var dir string
	var filePrefix string
	const maxCandidates = 50

	input := string(line[:pos])

	if c.completionType == bashComplete {
		// handle bash completion (executables in PATH)

		for _, pathdir := range filepath.SplitList(os.Getenv("PATH")) {
			dir, filePrefix = filepath.Split(input)
			if dir == "" {
				dir = "."
			}

			entries, err := filepath.Glob(filepath.Join(pathdir, filePrefix+"*"))
			if err != nil {
				continue
			}

			candidates = c.processEntries(entries, filePrefix, candidates, maxCandidates, false)
			if len(candidates) >= maxCandidates {
				break
			}
		}
	} else {
		// file name completion

		dir, filePrefix = filepath.Split(input)
		if dir == "" {
			dir = "."
		}

		entries, err := filepath.Glob(filepath.Join(dir, filePrefix+"*"))
		if err != nil {
			return nil, 0
		}

		candidates = c.processEntries(entries, filePrefix, candidates, maxCandidates, true)
	}

	return candidates, len(filePrefix)
}

func (c *completer) processEntries(entries []string, filePrefix string, candidates [][]rune, maxCandidates int, isFileComplete bool) [][]rune {
	for _, entry := range entries {
		if len(candidates) >= maxCandidates {
			break
		}

		name := filepath.Base(entry)

		if stat, err := os.Stat(entry); err == nil {
			if stat.IsDir() {
				if isFileComplete {
					name += "/"
				} else {
					continue
				}
			} else if isFileComplete && isBinaryFile(entry) {
				continue
			}
		}

		if strings.HasPrefix(name, filePrefix) {
			suffix := name[len(filePrefix):]
			candidates = append(candidates, []rune(suffix))
		}
	}

	return candidates
}

func isBinaryFile(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	const sampleSize = 8 * 1024
	buffer := make([]byte, sampleSize)

	bytesRead, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	for i := 0; i < bytesRead; i++ {
		if buffer[i] == 0 {
			return true
		}
	}

	return false
}

func (br *browseObj) userBashComp(prompt string) (string, bool, bool) {
	return br.promptWithCompletion(prompt, bashComplete)
}

func (br *browseObj) userFileComp(prompt string) (string, bool, bool) {
	return br.promptWithCompletion(prompt, fileComplete)
}

func (br *browseObj) promptWithCompletion(prompt string, cType completeType) (string, bool, bool) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		// readline does not support input from pipes
		return br.userInput(prompt)
	}

	cfg := &readline.Config{
		Prompt:          prompt,
		InterruptPrompt: "^C",
		EOFPrompt:       "^D",
		HistoryLimit:    0,
		Stdin:           br.tty,
		Stdout:          os.Stdout,
		AutoComplete:    &completer{completionType: cType},
	}

	signal.Ignore(syscall.SIGINT, syscall.SIGQUIT, syscall.SIGWINCH)
	moveCursor(br.dispHeight, 1, true)

	rl, err := readline.NewEx(cfg)
	if err != nil {
		errorExit(err)
		return "", false, false
	}

	defer func() {
		rl.Close()
		ttyBrowser()
		br.catchSignals()
		br.resizeWindow()
	}()

	line, err := rl.Readline()

	switch err {

	case readline.ErrInterrupt:
		// ctrl-c
		return "", true, false

	case io.EOF:
		// ctrl-d
		return "", true, true

	case nil:
		return line, false, false

	default:
		errorExit(err)
		return "", false, false
	}
}

// vim: set ts=4 sw=4 noet:
