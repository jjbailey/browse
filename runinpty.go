// runinpty.go
// Run a bash command in a pseudo tty
//
// Copyright (c) 2024-2026 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// Signal modes for PTY handling.
const (
	RUNSIGS  = 1
	WAITSIGS = 2
)

// runInPty executes a command in a pseudo-terminal and restores state.
func (br *browseObj) runInPty(cmdbuf string) {
	var err error

	bashPath, err := exec.LookPath("bash")
	if err != nil {
		if errors.Is(err, exec.ErrDot) {
			bashPath, _ = filepath.Abs(bashPath)
		} else {
			br.printMessage("Cannot find bash", MSG_RED)
			return
		}
	}

	cmd := exec.Command(bashPath, "-c", cmdbuf)

	// for manPage()
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("MANWIDTH=%d", br.dispWidth-1))

	// child signals
	br.ptySignals(RUNSIGS, nil)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		br.printMessage(fmt.Sprintf("Failed to start pty: %v", err), MSG_RED)
		// reset signals
		br.catchSignals()
		return
	}

	// need CURSAVE and CURRESTORE before this point
	defer ptmx.Close()

	pty.InheritSize(os.Stdout, ptmx)
	ptySave, err := term.MakeRaw(int(os.Stdout.Fd()))
	if err != nil {
		// Failed to set raw mode - restore signals and return
		ptmx.Close()
		br.catchSignals()
		return
	}

	// parent signals - pass ptmx to signal handler
	br.ptySignals(WAITSIGS, ptmx)

	execOK := make(chan bool, 1)
	var wg sync.WaitGroup

	// Goroutine to copy from tty to ptmx
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Custom copy that captures the last key press
		buf := make([]byte, 1)

		for {
			n, err := br.tty.Read(buf)
			if err != nil || n == 0 {
				break
			}

			// Store the last pressed key
			br.lastKey = buf[0]

			// Write to ptmx
			if _, err := ptmx.Write(buf[:n]); err != nil {
				// ptmx may be closed, break out of loop
				break
			}
		}

		// Send completion signal - channel is buffered so won't block
		execOK <- true
	}()

	// Copy from ptmx to stdout
	io.Copy(os.Stdout, ptmx)

	// Wait for command to finish
	cmd.Wait()

	// Restore terminal and reset window size BEFORE waiting for input
	term.Restore(int(os.Stdout.Fd()), ptySave)
	pty.InheritSize(os.Stdout, ptmx)
	br.dispHeight, br.dispWidth, _ = pty.Getsize(ptmx)
	br.dispRows = br.dispHeight - 1

	// Wait for the input goroutine to finish
	moveCursor(br.dispHeight, 1, true)
	fmt.Printf(MSG_GREEN + " Press any key to continue... " + VIDOFF)

	// Wait for user input or goroutine completion
	<-execOK

	// Ensure goroutine completes
	wg.Wait()

	br.catchSignals()
}

// ptySignals configures signal handling for PTY runs.
func (br *browseObj) ptySignals(sigSet int, ptmx *os.File) {
	if sigChan != nil {
		signal.Stop(sigChan)
		close(sigChan)
		sigChan = nil
	}

	switch sigSet {

	case RUNSIGS:
		signal.Ignore(syscall.SIGALRM, syscall.SIGURG)
		signal.Reset(syscall.SIGWINCH, syscall.SIGCHLD)
		return

	case WAITSIGS:
		sigChan = make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGWINCH, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)
		signal.Ignore(syscall.SIGALRM, syscall.SIGURG)
	}

	sc := sigChan

	go func() {
		defer signal.Stop(sc)

		for sig := range sc {
			switch sig {

			case syscall.SIGWINCH:
				// Only handle SIGWINCH if we have a valid ptmx
				if ptmx != nil {
					pty.InheritSize(os.Stdout, ptmx)
				}

			case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM:
				br.printMessage(fmt.Sprintf("%v \n", sig), MSG_RED)
				br.saneExit()
			}
		}
	}()
}

// vim: set ts=4 sw=4 noet:
