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
	pathDir, filePrefix := filepath.Split(input)

	if pathDir == "" {
		pathDir = "."
	}

	if c.completionType == bashComplete {
		// handle bash completion (executables in PATH)
		candidates = c.completeBash(pathDir, filePrefix, maxCandidates)
	} else {
		// file name completion
		candidates = c.completeFiles(pathDir, filePrefix, maxCandidates)
	}

	return candidates, len(filePrefix)
}

func (c *completer) completeBash(pathDir, filePrefix string, maxCandidates int) [][]rune {
	candidates := make([][]rune, 0, maxCandidates)

	// absolute path

	if strings.HasPrefix(pathDir, "/") {
		// glob directories
		entries, err := filepath.Glob(pathDir + "*")
		if err != nil {
			return nil
		}

		candidates = c.processEntries(entries, filePrefix, candidates, maxCandidates, false)
		return candidates
	}

	// search PATH

	for _, pathDir := range filepath.SplitList(os.Getenv("PATH")) {
		entries, err := filepath.Glob(filepath.Join(pathDir, filePrefix+"*"))
		if err != nil {
			continue
		}

		candidates = c.processEntries(entries, filePrefix, candidates, maxCandidates, false)
	}

	return candidates
}

func (c *completer) completeFiles(pathDir, filePrefix string, maxCandidates int) [][]rune {
	// search dir

	entries, err := filepath.Glob(filepath.Join(pathDir, filePrefix+"*"))
	if err != nil {
		return nil
	}

	candidates := make([][]rune, 0, maxCandidates)
	candidates = c.processEntries(entries, filePrefix, candidates, maxCandidates, false)
	return candidates
}

func (c *completer) processEntries(entries []string, filePrefix string, candidates [][]rune, maxCandidates int, isFileComplete bool) [][]rune {
	// processEntries

	for _, entry := range entries {
		if len(candidates) >= maxCandidates {
			break
		}

		name := filepath.Base(entry)
		if !strings.HasPrefix(name, filePrefix) {
			continue
		}

		stat, err := os.Stat(entry)
		if err != nil {
			continue
		}

		if stat.IsDir() {
			name += "/"
		} else if isFileComplete && isBinaryFile(entry) {
			continue
		}

		candidates = append(candidates, []rune(name[len(filePrefix):]))
	}

	return candidates
}

func isBinaryFile(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	const sampleSize = 4 * 1024
	buffer := make([]byte, sampleSize)

	bytesRead, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	for _, b := range buffer[:bytesRead] {
		if b == 0 {
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

	// don't allow readline to redraw the screen on sigwinch
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
