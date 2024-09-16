// bash.go
// run a command with bash
//
// Copyright (c) 2024 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"regexp"
)

var prevCommand string

func (x *browseObj) bashCommand() bool {
	// run a command with bash

	fmt.Print(LINEWRAPON)
	input, cancel := x.userInput("!")

	if len(input) == 0 {
		cancel = true
	}

	if len(input) > 0 {
		// substitute ! with the previous command
		bangbuf := subCommandChars(input, "!", prevCommand)

		// save
		prevCommand = bangbuf

		// substitute % with the current file name
		cmdbuf := subCommandChars(bangbuf, "%", `'`+x.fileName+`'`)

		// substitute & with the current search pattern
		if len(x.pattern) > 0 {
			cmdbuf = subCommandChars(cmdbuf, "&", `'`+x.pattern+`'`)
		}

		if len(cmdbuf) > 0 {
			// feedback
			moveCursor(x.dispHeight, 1, true)
			fmt.Print("---\n")
			fmt.Printf("$ %s\n", cmdbuf)

			// set up env, run
			fmt.Print(LINEWRAPON) // again
			resetScrRegion()

			moveCursor(x.dispHeight, 1, true)
			x.runInPty(cmdbuf)
		}

		cancel = false
	}

	if cancel {
		x.restoreLast()
		moveCursor(2, 1, false)
	} else {
		x.resizeWindow()
	}

	return cancel
}

func subCommandChars(input, char, repl string) string {
	// negative lookbehind not supported in golang RE2 engine
	// pattern := `(?<!\\)%`

	pattern := `(^|[^\\])` + regexp.QuoteMeta(char)
	replstr := "${1}" + repl

	re, err := regexp.Compile(pattern)

	if err != nil {
		return ""
	}

	rbuf1 := input

	for rbuf2 := re.ReplaceAllString(rbuf1, replstr); rbuf2 != rbuf1; {
		rbuf1 = rbuf2
	}

	return rbuf1
}

// vim: set ts=4 sw=4 noet:
