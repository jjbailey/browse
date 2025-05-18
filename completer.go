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
	const maxCandidates = 50
	var candidates [][]rune
	input := string(line[:pos])
	dir, filePrefix := filepath.Split(input)

	if dir == "" {
		dir = "."
	}

	if c.completionType == bashComplete {
		// handle bash completion (executables in PATH)
		candidates = c.completeBash(filePrefix, maxCandidates)
	} else {
		// file name completion
		candidates = c.completeFiles(dir, filePrefix, maxCandidates)
	}

	return candidates, len(filePrefix)
}

func (c *completer) completeBash(filePrefix string, maxCandidates int) [][]rune {
	var candidates [][]rune

	for _, pathdir := range filepath.SplitList(os.Getenv("PATH")) {
		entries, err := filepath.Glob(filepath.Join(pathdir, filePrefix+"*"))
		if err != nil {
			continue
		}

		candidates = c.processEntries(entries, filePrefix, candidates, maxCandidates, false)
	}

	return candidates
}

func (c *completer) completeFiles(dir, filePrefix string, maxCandidates int) [][]rune {
	entries, err := filepath.Glob(filepath.Join(dir, filePrefix+"*"))
	if err != nil {
		return nil
	}

	return c.processEntries(entries, filePrefix, nil, maxCandidates, true)
}

func (c *completer) processEntries(entries []string, filePrefix string, candidates [][]rune, maxCandidates int, isFileComplete bool) [][]rune {
	if candidates == nil {
		candidates = make([][]rune, 0, maxCandidates)
	}

	for _, entry := range entries {
		if len(candidates) >= maxCandidates {
			break
		}

		name := filepath.Base(entry)
		if !strings.HasPrefix(name, filePrefix) {
			continue
		}

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

		suffix := name[len(filePrefix):]
		candidates = append(candidates, []rune(suffix))
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
	// userBashComp prompts the user with bash command completion
	return br.promptWithCompletion(prompt, bashComplete)
}

func (br *browseObj) userFileComp(prompt string) (string, bool, bool) {
	// userFileComp prompts the user with file name completion
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

	// ignore signals that could interfere with readline
	signal.Ignore(syscall.SIGINT, syscall.SIGQUIT, syscall.SIGWINCH)
	moveCursor(br.dispHeight, 1, true)

	rl, err := readline.NewEx(cfg)
	if err != nil {
		errorExit(err)
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
		// function needs to return something
		return "", false, false
	}
}

// vim: set ts=4 sw=4 noet:
