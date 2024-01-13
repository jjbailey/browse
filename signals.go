// signals.go
// paging and some support functions
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

func (x *browseObj) resetSignals() {
	// signals

	signal.Reset()
}

func (x *browseObj) catchSignals() {
	// signals

	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan)
	signal.Ignore(syscall.SIGCHLD)
	signal.Ignore(syscall.SIGURG)

	go func() {
		for {
			sig := <-sigChan

			switch sig {

			case syscall.SIGWINCH:
				x.resizeWindow()

			default:
				fmt.Printf("caught signal: %v\n", sig)
				x.saneExit()
			}
		}
	}()
}

func (x *browseObj) resizeWindow() {
	// process window size changes

	x.dispWidth, x.dispHeight, _ = term.GetSize(int(x.tty.Fd()))
	x.dispRows = x.dispHeight - 1
	x.lastMatch = SEARCH_RESET

	x.pageHeader()
	x.pageCurrent()

	if x.modeTail || x.modeScrollDown {
		fmt.Printf("%s", CURRESTORE)
	}
}

func (x *browseObj) saneExit() {
	// clean up

	ttyRestore()
	resetScrRegion()
	fmt.Printf("%s", LINEWRAPON)
	movecursor(x.dispHeight, 1, true)

	if x.fromStdin {
		os.Remove(x.fileName)
	}

	if !x.fromStdin && x.saveRC {
		writeRcFile(x)
	}

	os.Exit(0)
}

// vim: set ts=4 sw=4 noet:
