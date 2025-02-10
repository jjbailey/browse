// bash.go
// run a command with bash
//
// Copyright (c) 2024-2025 jjb
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

func (br *browseObj) bashCommand() bool {
	// run a command with bash

	fmt.Print(LINEWRAPON)
	input, cancel := br.userInput("!")

	if len(input) == 0 {
		cancel = true
	}

	if len(input) > 0 {
		// substitute ! with the previous command
		bangbuf := subCommandChars(input, "!", prevCommand)

		// save
		prevCommand = bangbuf

		// substitute % with the current file name
		cmdbuf := subCommandChars(bangbuf, "%", `'`+br.fileName+`'`)

		// substitute & with the current search pattern
		if len(br.pattern) > 0 {
			cmdbuf = subCommandChars(cmdbuf, "&", `'`+br.pattern+`'`)
		}

		if len(cmdbuf) > 0 {
			// feedback
			moveCursor(br.dispHeight, 1, true)
			fmt.Print("---\n")
			fmt.Printf("$ %s\n", cmdbuf)

			// set up env, run
			fmt.Print(LINEWRAPON) // again
			resetScrRegion()

			moveCursor(br.dispHeight, 1, true)
			br.runInPty(cmdbuf)
		}

		cancel = false
	}

	if cancel {
		br.restoreLast()
		moveCursor(2, 1, false)
	} else {
		br.resizeWindow()
	}

	return cancel
}

func subCommandChars(input, char, repl string) string {
	// negative lookbehind not supported in golang RE2 engine
	// pattern := `(?<!\\)%`

	pattern := `(^|[^\\])` + regexp.QuoteMeta(char)
	replstr := `${1}` + repl

	re, err := regexp.Compile(pattern)

	if err != nil {
		return ""
	}

	return re.ReplaceAllString(input, replstr)
}

// vim: set ts=4 sw=4 noet:
