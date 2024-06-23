// runinpty.go
//
// Copyright (c) 2024 jjb
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

// global to avoid race
var ptmx *os.File

func (x *browseObj) runInPty(cmdbuf string) {
	var err error

	cmd := exec.Command("bash", "-c", cmdbuf)
	x.ptySignals()
	ptmx, err = pty.Start(cmd)

	if err != nil {
		// reset signals
		x.catchSignals()
		return
	}

	moveCursor(x.dispHeight, 1, true)
	defer ptmx.Close()
	pty.InheritSize(os.Stdout, ptmx)
	ptySave, err := term.MakeRaw(int(os.Stdout.Fd()))

	// reset signals
	x.catchSignals()

	execOK := make(chan bool)
	go func(ch chan bool) {
		io.Copy(ptmx, x.tty)
		ch <- true
	}(execOK)
	io.Copy(os.Stdout, ptmx)
	cmd.Wait()

	// restore and reset window size
	term.Restore(int(os.Stdout.Fd()), ptySave)
	pty.InheritSize(os.Stdout, ptmx)
	x.dispHeight, x.dispWidth, _ = pty.Getsize(ptmx)
	x.dispRows = x.dispHeight - 1

	moveCursor(x.dispHeight, 1, true)
	fmt.Printf(MSG_GREEN + " Press any key to continue... " + VIDOFF)
	<-execOK
}

func (x *browseObj) ptySignals() {
	// signals

	sigChan := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGWINCH)
	signal.Notify(sigChan)
	signal.Ignore(syscall.SIGALRM)
	signal.Ignore(syscall.SIGURG)

	go func() {
		for {
			sig := <-sigChan

			switch sig {

			case syscall.SIGWINCH:
				pty.InheritSize(os.Stdout, ptmx)

			default:
				x.warnMessage(fmt.Sprintf("%v \n", sig))
				x.saneExit()
			}
		}
	}()
}

// vim: set ts=4 sw=4 noet:
