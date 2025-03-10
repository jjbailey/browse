// signals.go
// signal handling functions
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

func (br *browseObj) catchSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)
	signal.Ignore(syscall.SIGALRM, syscall.SIGCHLD, syscall.SIGURG)

	go func() {
		for {
			sig := <-sigChan

			switch sig {

			case syscall.SIGWINCH:
				br.resizeWindow()

			default:
				br.printMessage(fmt.Sprintf("%v \n", sig), MSG_RED)
				br.saneExit()
			}
		}
	}()
}

func (br *browseObj) resizeWindow() {
	// process window size changes

	br.dispWidth, br.dispHeight, _ = term.GetSize(int(br.tty.Fd()))
	br.dispRows = br.dispHeight - 1
	br.lastMatch = SEARCH_RESET

	br.pageHeader()
	br.pageCurrent()

	if br.inMotion() {
		fmt.Print(CURRESTORE)
	}
}

func (br *browseObj) saneExit() {
	// clean up

	ttyRestore()
	resetScrRegion()
	fmt.Print(LINEWRAPON)
	moveCursor(br.dispHeight, 1, true)

	if br.fromStdin {
		os.Remove(br.fileName)
	}

	if !br.fromStdin && br.saveRC {
		br.writeRcFile()
	}

	os.Exit(0)
}

// vim: set ts=4 sw=4 noet:
