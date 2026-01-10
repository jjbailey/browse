// user.go
// user input functions
//
// Copyright (c) 2024-2026 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// wait for any key press
func (br *browseObj) userAnyKey(promptStr string) {
	const timeout = 500 * time.Millisecond

	signal.Ignore(syscall.SIGINT, syscall.SIGQUIT, syscall.SIGWINCH)
	defer signal.Reset(syscall.SIGINT, syscall.SIGQUIT, syscall.SIGWINCH)

	// reset tty
	ttyBrowser()

	// promptStr is optional

	if promptStr == "" {
		moveCursor(2, 1, false)
	} else {
		moveCursor(br.dispHeight, 1, true)
		fmt.Print(promptStr)
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
}

// get user input line with basic editing
func (br *browseObj) userInput(promptStr string) (string, bool) {
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
		linebuf     string
		cancelled   bool
		done        bool
		winchCaught bool
	)

	signal.Ignore(syscall.SIGINT, syscall.SIGQUIT, syscall.SIGWINCH)
	defer signal.Reset(syscall.SIGINT, syscall.SIGQUIT, syscall.SIGWINCH)

	// promptStr

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH, syscall.SIGINT)
	defer func() {
		signal.Stop(sigChan)
		close(sigChan)
	}()

	ttyPrompter()
	fmt.Print("\r" + CURSAVE)
	moveCursor(br.dispHeight, 1, true)
	fmt.Print(promptStr)
	br.shownMsg = true

	b := make([]byte, 1)

	for {
	DrainSignals:
		for {
			select {
			case sig := <-sigChan:
				switch sig {

				case syscall.SIGWINCH:
					winchCaught = true

				case syscall.SIGINT:
					cancelled = true
				}

			default:
				break DrainSignals
			}
		}

		if cancelled {
			break
		}

		n, err := br.tty.Read(b)

		if winchCaught {
			// restore and reset window size
			br.resizeWindow()
			moveCursor(br.dispHeight, 1, true)
			fmt.Print(promptStr, linebuf)
			winchCaught = false
		}

		if err != nil && err != io.EOF {
			errorExit(err)
			return "", false
		}

		if n == 0 {
			continue
		}

		inputChar := b[0]

		switch inputChar {

		case NEWLINE, CARRETURN:
			if len(linebuf) > 0 {
				done = true
			} else {
				cancelled = true
			}

		case ESCAPE:
			cancelled = true

		case BACKSPACE, DELETE:
			if len(linebuf) > 0 {
				linebuf = linebuf[:len(linebuf)-1]
				moveCursor(br.dispHeight, 1, true)
				fmt.Print(promptStr, linebuf)
			} else {
				cancelled = true
			}

		case ERASEWORD:
			if n := strings.LastIndex(linebuf, " "); n > 0 {
				linebuf = linebuf[:n]
			} else {
				linebuf = ""
			}

			moveCursor(br.dispHeight, 1, true)
			fmt.Print(promptStr, linebuf)

		case ERASELINE:
			linebuf = ""
			moveCursor(br.dispHeight, 1, true)
			fmt.Print(promptStr, linebuf)

		default:
			linebuf += string(inputChar)
			fmt.Print(string(inputChar))
		}

		if cancelled {
			br.restoreLast()
			moveCursor(2, 1, false)
			break
		}

		if done {
			break
		}
	}

	// reset signals
	br.catchSignals()

	// reset tty
	ttyBrowser()

	return linebuf, cancelled
}

// vim: set ts=4 sw=4 noet:
