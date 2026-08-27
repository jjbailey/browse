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
	mapSize := br.currentMapSize()
	wasAtEOF := br.lastRow > mapSize

	br.screenInit(br.tty)
	topLine := resizedTopLine(br.firstRow, br.dispRows, mapSize, wasAtEOF)

	// pageHeader clears the terminal. Update firstRow before printPage so its
	// nearby-page optimization cannot choose an incremental scroll against the
	// now-empty screen.
	br.firstRow = topLine
	br.pageHeader()
	br.printPage(topLine)

	if br.inMotion() {
		fmt.Print(CURRESTORE)
	}
}

// resizedTopLine preserves an EOF-bottom anchor across terminal size changes.
func resizedTopLine(firstRow, dispRows, mapSize int, wasAtEOF bool) int {
	if wasAtEOF {
		return maximum(0, mapSize-dispRows+1)
	}

	return adjustLineNumber(firstRow, dispRows, mapSize)
}

// saneExit restores terminal state and exits cleanly.
func (br *browseObj) saneExit() {
	ttyRestore()
	resetScrRegion()
	fmt.Print(LINEWRAPON + SGR0)
	moveCursor(br.dispHeight, 1, true)

	if br.fromStdin {
		os.Remove(br.currentFileName())
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
	signal.Reset(syscall.SIGCHLD)
	signal.Ignore(syscall.SIGALRM, syscall.SIGURG)

	go func() {
		for sig := range sigChan {
			switch sig {

			case syscall.SIGWINCH:
				// Rendering belongs to the main goroutine. The command loop
				// polls the tty, then drains this event on its next tick.
				br.mutex.Lock()
				br.resizePending = true
				br.mutex.Unlock()

			default:
				br.printMessage(fmt.Sprintf("%v \n", sig), MSG_RED)
				br.saneExit()
			}
		}
	}()
}

// vim: set ts=4 sw=4 noet:
