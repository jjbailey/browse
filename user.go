// user.go
// user input functions
//
// Copyright (c) 2024 jjb
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

func (x *browseObj) userAnyKey(prompt string) {
	// wait for a key press

	b := make([]byte, 1)

	signal.Ignore(syscall.SIGINT, syscall.SIGQUIT)

	// prompt is optional

	if prompt == "" {
		movecursor(2, 1, false)
	} else {
		movecursor(x.dispHeight, 1, true)
		ttyBrowser()
		fmt.Printf("%s", prompt)
	}

	for {
		count, err := x.tty.Read(b)

		if err != nil && err != io.EOF {
			errorExit(err)
		}

		if count > 0 {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	signal.Reset(syscall.SIGINT, syscall.SIGQUIT)
	x.restoreLast()
}

func (x *browseObj) userInput(prompt string) (string, bool) {
	const (
		NEWLINE   = '\n'
		CARRETURN = '\r'
		BACKSPACE = '\b'
		ERASEWORD = '\025'
		ERASELINE = '\027'
		ESCAPE    = '\033'
		DELETE    = '\177'
	)

	var linebuf string
	var cancel bool = false

	signal.Ignore(syscall.SIGINT, syscall.SIGQUIT)
	ttyPrompter()
	movecursor(x.dispHeight, 1, true)
	fmt.Printf("%s", prompt)
	x.shownMsg = true

	for {
		var nbuf string

		b := make([]byte, 1)
		_, err := x.tty.Read(b)
		errorExit(err)

		inputbuf := string(b)

		switch inputbuf[0] {

		case NEWLINE, CARRETURN, ESCAPE:
			// ignore for now

		case BACKSPACE, DELETE:
			if len(linebuf) == 0 {
				// cancel
				linebuf = ""
				cancel = true
				break
			} else {
				nbuf = strings.TrimSuffix(linebuf, string(linebuf[len(linebuf)-1]))
			}

			linebuf = nbuf
			movecursor(x.dispHeight, 1, true)
			fmt.Printf("%s%s", prompt, linebuf)

		case ERASEWORD:
			n := strings.LastIndex(linebuf, " ")

			if n > 0 {
				nbuf = linebuf[:n]
				linebuf = nbuf
			} else {
				linebuf = ""
			}

			movecursor(x.dispHeight, 1, true)
			fmt.Printf("%s%s", prompt, linebuf)

		case ERASELINE:
			linebuf = ""
			movecursor(x.dispHeight, 1, true)
			fmt.Printf("%s%s", prompt, linebuf)

		default:
			fmt.Printf("%s", inputbuf)
			linebuf += inputbuf
		}

		if cancel || inputbuf[0] == NEWLINE || inputbuf[0] == CARRETURN {
			break
		}
	}

	signal.Reset(syscall.SIGINT, syscall.SIGQUIT)
	ttyBrowser()
	x.restoreLast()

	return linebuf, cancel
}

// vim: set ts=4 sw=4 noet:
