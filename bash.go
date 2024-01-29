// bash.go
// run a command with bash
//
// Copyright (c) 2024 jjb
// All rights reserved.

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

	fmt.Printf("%s", LINEWRAPON)
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
			fmt.Printf("%s", LINEWRAPON) // again
			resetScrRegion()

			movecursor(x.dispHeight, 1, true)
			x.ptySignals()
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

func (x *browseObj) runInPty(cmdbuf string) error {
	cmd := exec.Command("bash", "-c", cmdbuf)

	ptmx, err := pty.Start(cmd)

	if err != nil {
		return err
	}

	pty.InheritSize(os.Stdin, ptmx)
	ptySave, err := term.MakeRaw(int(os.Stdin.Fd()))

	if err != nil {
		return err
	}

	execOK := make(chan bool)
	go func(ch chan bool) {
		io.Copy(ptmx, os.Stdin)
		ch <- true
	}(execOK)
	io.Copy(os.Stdout, ptmx)

	ptmx.Close()
	term.Restore(int(os.Stdin.Fd()), ptySave)
	movecursor(x.dispHeight, 1, true)
	fmt.Printf(VIDMESSAGE + " Press any key to continue... " + VIDOFF)
	<-execOK
	return nil
}

func (x *browseObj) ptySignals() {
	// signals

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan)
	signal.Ignore(syscall.SIGCHLD)
	signal.Ignore(syscall.SIGURG)

	go func() {
		for {
			switch <-sigChan {

			case syscall.SIGWINCH:
				x.resizeWindow()

			default:
				x.saneExit()
			}
		}
	}()
}

func subCommandChars(input, char, repl string) string {
	var rbuf1 string

	// negative lookbehind doesn't compile, e.g.
	// pattern := `(?<!\\)%`

	pattern := `(^|[^\\])` + char
	replstr := "${1}" + repl

	re, err := regexp.Compile(pattern)

	if err != nil {
		return ""
	}

	rbuf1 = input

	for {
		rbuf2 := re.ReplaceAllString(rbuf1, replstr)

		if rbuf2 == rbuf1 {
			break
		}

		rbuf1 = rbuf2
	}

	return rbuf1
}

// vim: set ts=4 sw=4 noet:
