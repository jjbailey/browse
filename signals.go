// signals.go
// signal handling functions
//
// Copyright (c) 2024-2026 jjb
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
)

// sigChan receives OS signals for the application.
var sigChan chan os.Signal

// resizeWindow handles terminal resize events.
func (br *browseObj) resizeWindow() {
	br.screenInit(br.tty)
	br.pageHeader()
	br.pageCurrent()

	if br.inMotion() {
		fmt.Print(CURRESTORE)
	}
}

// saneExit restores terminal state and exits cleanly.
func (br *browseObj) saneExit() {
	const SGR0 = "\033[0m\017"

	ttyRestore()
	resetScrRegion()
	fmt.Print(LINEWRAPON + SGR0)
	moveCursor(br.dispHeight, 1, true)

	if br.fromStdin {
		os.Remove(br.fileName)
	}

	if !br.fromStdin && br.saveRC {
		br.writeRcFile()
	}

	os.Exit(0)
}

// catchSignals installs signal handlers for the browse session.
func (br *browseObj) catchSignals() {
	if sigChan != nil {
		signal.Stop(sigChan)
		close(sigChan)
	}

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGWINCH)
	signal.Ignore(syscall.SIGALRM, syscall.SIGCHLD, syscall.SIGURG)

	go func() {
		for sig := range sigChan {
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

// vim: set ts=4 sw=4 noet:
