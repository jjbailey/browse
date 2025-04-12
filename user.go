// user.go
// user input functions
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"io"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func (br *browseObj) userAnyKey(prompt string) {
	// wait for a key press

	const timeout = 500 * time.Millisecond

	signal.Ignore(syscall.SIGINT, syscall.SIGQUIT, syscall.SIGWINCH)
	defer signal.Reset(syscall.SIGINT, syscall.SIGQUIT)

	// prompt is optional

	if prompt == "" {
		moveCursor(2, 1, false)
	} else {
		moveCursor(br.dispHeight, 1, true)
		ttyBrowser()
		fmt.Print(prompt)
	}

	b := make([]byte, 1)

	for {
		if n, err := br.tty.Read(b); err != nil {
			if err != io.EOF {
				errorExit(err)
			}
		} else if n > 0 {
			break
		}

		time.Sleep(timeout)
	}

	br.restoreLast()
}

func (br *browseObj) userInput(prompt string) (string, bool) {
	const (
		NEWLINE   = '\n'
		CARRETURN = '\r'
		BACKSPACE = '\b'
		ERASEWORD = '\027'
		ERASELINE = '\025'
		ESCAPE    = '\033'
		DELETE    = '\177'
	)

	var (
		linebuf string
		cancel  bool
	)

	signal.Ignore(syscall.SIGINT, syscall.SIGQUIT, syscall.SIGWINCH)
	defer signal.Reset(syscall.SIGINT, syscall.SIGQUIT)
	ttyPrompter()
	fmt.Printf("\r%s", CURSAVE)
	moveCursor(br.dispHeight, 1, true)
	fmt.Printf("%s", prompt)
	br.shownMsg = true

	for {
		b := make([]byte, 1)
		_, err := br.tty.Read(b)

		if err != nil {
			errorExit(err)
			return "", false
		}

		inputChar := b[0]

		switch inputChar {

		case NEWLINE, CARRETURN, ESCAPE:
			// ignore for now

		case BACKSPACE, DELETE:
			if len(linebuf) > 0 {
				linebuf = strings.TrimSuffix(linebuf, string(linebuf[len(linebuf)-1]))
				moveCursor(br.dispHeight, 1, true)
				fmt.Printf("%s%s", prompt, linebuf)
			} else {
				cancel = true
			}

		case ERASEWORD:
			if n := strings.LastIndex(linebuf, " "); n > 0 {
				linebuf = linebuf[:n]
			} else {
				linebuf = ""
			}

			moveCursor(br.dispHeight, 1, true)
			fmt.Printf("%s%s", prompt, linebuf)

		case ERASELINE:
			linebuf = ""
			moveCursor(br.dispHeight, 1, true)
			fmt.Printf("%s%s", prompt, linebuf)

		default:
			linebuf += string(inputChar)
			fmt.Print(string(inputChar))
		}

		if cancel || inputChar == NEWLINE || inputChar == CARRETURN {
			break
		}
	}

	ttyBrowser()
	br.restoreLast()
	moveCursor(2, 1, false)

	return linebuf, cancel
}

// vim: set ts=4 sw=4 noet:
