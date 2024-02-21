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
	"io"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

var prevCommand string

func (x *browseObj) bashCommand() bool {
	// run a command with bash

	fmt.Print(LINEWRAPON)
	input, cancel := x.userInput("!")

	if len(input) > 0 {
		// substitute ! with the previous command
		bangbuf := subCommandChars(input, "!", prevCommand)

		// save
		prevCommand = bangbuf

		// substitute % with the current file name
		cmdbuf := subCommandChars(bangbuf, "%", x.fileName)

		if len(cmdbuf) > 0 {
			// feedback
			movecursor(x.dispHeight, 1, true)
			fmt.Print("---\n")
			fmt.Printf("$ %s\n", cmdbuf)

			// set up env, run
			fmt.Print(LINEWRAPON) // again
			resetScrRegion()

			movecursor(x.dispHeight, 1, true)
			x.runInPty(cmdbuf)
		}
	}

	if cancel {
		x.restoreLast()
		movecursor(2, 1, false)
	} else {
		x.resizeWindow()
	}

	return cancel
}

// global to avoid race
var ptmx *os.File

func (x *browseObj) runInPty(cmdbuf string) error {
	var err error

	cmd := exec.Command("bash", "-c", cmdbuf)
	x.ptySignals()
	ptmx, err = pty.Start(cmd)

	if err != nil {
		x.catchSignals()
		return err
	}

	movecursor(x.dispHeight, 1, true)
	defer ptmx.Close()
	pty.InheritSize(os.Stdout, ptmx)
	ptySave, err := term.MakeRaw(int(os.Stdout.Fd()))

	if err != nil {
		x.catchSignals()
		return err
	}

	execOK := make(chan bool)
	go func(ch chan bool) {
		io.Copy(ptmx, x.tty)
		ch <- true
	}(execOK)
	io.Copy(os.Stdout, ptmx)

	term.Restore(int(os.Stdout.Fd()), ptySave)

	// reset window size
	pty.InheritSize(os.Stdout, ptmx)
	x.dispHeight, x.dispWidth, _ = pty.Getsize(ptmx)
	x.dispRows = x.dispHeight - 1

	movecursor(x.dispHeight, 1, true)
	fmt.Printf(VIDMESSAGE + " Press any key to continue... " + VIDOFF)

	// reset signals
	x.catchSignals()

	<-execOK
	return nil
}

func (x *browseObj) ptySignals() {
	// signals

	sigChan := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGWINCH)
	signal.Notify(sigChan)
	signal.Ignore(syscall.SIGCHLD)
	signal.Ignore(syscall.SIGURG)

	go func() {
		for {
			switch <-sigChan {

			case syscall.SIGWINCH:
				pty.InheritSize(os.Stdout, ptmx)

			default:
				x.saneExit()
			}
		}
	}()
}

func subCommandChars(input, char, repl string) string {
	// negative lookbehind doesn't compile, e.g.
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
		rbuf2 = re.ReplaceAllString(rbuf1, replstr)
	}

	return rbuf1
}

// vim: set ts=4 sw=4 noet:
