// bash.go
// Run a command with bash
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"strings"
)

var PrevCommand string

func (br *browseObj) bashCommand() {
	// Run a command with bash

	for {
		moveCursor(br.dispRows, 1, true)

		input, cancelled := userBashComp()
		if cancelled {
			br.pageCurrent()
			return
		}

		input = strings.TrimSpace(input)
		if input == "" {
			br.pageCurrent()
			return
		}

		// Prevent overflows
		if len(input) > READBUFSIZ {
			br.printMessage("Command too long", MSG_RED)
			return
		}

		// Unquote bash commands
		if strings.Contains(input, " ") || strings.Contains(input, "|") {
			input = strings.ReplaceAll(input, "'", "")
		}

		// Fast-path: batch substitutions
		cmdbuf := input
		if strings.Contains(cmdbuf, "!") {
			cmdbuf = subCommandChars(cmdbuf, "!", PrevCommand)
		}
		PrevCommand = cmdbuf

		if strings.Contains(cmdbuf, "%") {
			cmdbuf = subCommandChars(cmdbuf, "%", shellEscapeSingle(br.fileName))
		}

		if br.pattern != "" && strings.Contains(cmdbuf, "&") {
			cmdbuf = subCommandChars(cmdbuf, "&", shellEscapeSingle(br.pattern))
		}

		if cmdbuf == "" {
			br.pageCurrent()
			return
		}

		// Save command to history
		updateHistory(cmdbuf, commHistory)

		// Setup
		fmt.Print(LINEWRAPON)

		// Run command in a PTY
		resetScrRegion()
		br.runInPty(cmdbuf)

		// Repeat only if lastKey is bang
		if br.lastKey != '!' {
			break
		}
	}

	br.resizeWindow()

	// gratuitous save cursor
	moveCursor(2, 1, false)
	fmt.Print(CURSAVE)
}

// vim: set ts=4 sw=4 noet:
