// runinpty.go
// run a bash command in a pseudo tty
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

const (
	RUNSIGS  = 1
	WAITSIGS = 2
)

// global to avoid race
var ptmx *os.File

func (br *browseObj) runInPty(cmdbuf string) {
	var err error

	cmd := exec.Command("bash", "-c", cmdbuf)
	// child signals
	br.ptySignals(RUNSIGS)
	ptmx, err = pty.Start(cmd)
	if err != nil {
		// reset signals
		br.catchSignals()
		return
	}

	// need CURSAVE and CURRESTORE before this point
	defer ptmx.Close()
	pty.InheritSize(os.Stdout, ptmx)
	ptySave, _ := term.MakeRaw(int(os.Stdout.Fd()))

	// parent signals
	br.ptySignals(WAITSIGS)

	execOK := make(chan bool)
	go func(ch chan bool) {
		io.Copy(ptmx, br.tty)
		ch <- true
	}(execOK)
	io.Copy(os.Stdout, ptmx)
	cmd.Wait()

	// restore and reset window size
	term.Restore(int(os.Stdout.Fd()), ptySave)
	pty.InheritSize(os.Stdout, ptmx)
	br.dispHeight, br.dispWidth, _ = pty.Getsize(ptmx)
	br.dispRows = br.dispHeight - 1

	moveCursor(br.dispHeight, 1, true)
	fmt.Printf(MSG_GREEN + " Press any key to continue... " + VIDOFF)
	<-execOK

	// reset signals
	br.catchSignals()
}

func (br *browseObj) ptySignals(sigSet int) {
	// signals for pty processing

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan)

	switch sigSet {

	case RUNSIGS:
		signal.Ignore(syscall.SIGALRM, syscall.SIGURG)
		signal.Reset(syscall.SIGWINCH)

	case WAITSIGS:
		signal.Ignore(syscall.SIGALRM, syscall.SIGCHLD, syscall.SIGURG)
	}

	go func() {
		for sig := range sigChan {
			switch sig {

			case syscall.SIGWINCH:
				pty.InheritSize(os.Stdout, ptmx)

			default:
				br.printMessage(fmt.Sprintf("%v \n", sig), MSG_RED)
				br.saneExit()
			}
		}
	}()
}

// vim: set ts=4 sw=4 noet:
