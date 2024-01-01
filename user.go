// user.go
// user input functions
//
// Copyright (c) 2024 jjb
// All rights reserved.

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
	b := make([]byte, 1)

	// prompt is optional

	signal.Ignore(syscall.SIGINT, syscall.SIGQUIT)

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

func (x *browseObj) userInput(prompt string) string {
	const (
		NEWLINE   = "\n"
		CARRETURN = "\r"
		BACKSPACE = "\b"
		ERASEWORD = "\025"
		ERASELINE = "\027"
	)

	var linebuf string
	var nbuf string

	b := make([]byte, 1)

	signal.Ignore(syscall.SIGINT, syscall.SIGQUIT)
	ttyPrompter()
	movecursor(x.dispHeight, 1, true)
	fmt.Printf("%s", prompt)

	for {
		_, err := x.tty.Read(b)
		errorExit(err)

		inputbuf := string(b)

		if inputbuf == NEWLINE || inputbuf == CARRETURN {
			break
		}

		if inputbuf == BACKSPACE {
			if len(linebuf) > 0 {
				nbuf = strings.TrimSuffix(linebuf, string(linebuf[len(linebuf)-1]))
			}

			linebuf = nbuf
			movecursor(x.dispHeight, 1, true)
			fmt.Printf("%s%s", prompt, linebuf)
			continue
		}

		if inputbuf == ERASEWORD {
			n := strings.LastIndex(linebuf, " ")

			if n > 0 {
				nbuf = linebuf[:n]
				linebuf = nbuf
			} else {
				linebuf = ""
			}

			movecursor(x.dispHeight, 1, true)
			fmt.Printf("%s%s", prompt, linebuf)
			continue
		}

		if inputbuf == ERASELINE {
			linebuf = ""
			movecursor(x.dispHeight, 1, true)
			fmt.Printf("%s%s", prompt, linebuf)
			continue
		}

		fmt.Printf("%s", inputbuf)
		linebuf += inputbuf
	}

	signal.Reset(syscall.SIGINT, syscall.SIGQUIT)
	ttyBrowser()
	x.restoreLast()

	return linebuf
}

// vim: set ts=4 sw=4 noet:
