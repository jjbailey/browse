// init.go
// functions to initialize objects
//
// Copyright (c) 2024-2026 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"os"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// fileInit initializes browseObj state for a new file.
func (br *browseObj) fileInit(fp *os.File, fileName, title string, fromStdin bool) {
	var rescueFd int

	br.mutex.Lock()
	oldFp := br.fp
	if br.rescueFd > 0 {
		rescueFd = br.rescueFd
		br.rescueFd = 0
		br.fdLink = ""
	}
	br.fileName = fileName
	br.fp = fp
	br.fromStdin = fromStdin
	if !fromStdin {
		br.stdinDone = nil
	}
	br.fileSeq++
	br.rereadPending = false
	br.readerAlive = false
	// drop display work posted for the previous file
	br.pendingMsg = ""
	br.pendingMsgTransient = false
	br.refreshPending = false
	br.scrollCancelPending = false
	br.partialRedrawFrom = 0
	br.mutex.Unlock()

	if oldFp != nil && oldFp != fp {
		oldFp.Close()
	}
	if rescueFd > 0 {
		_ = unix.Close(rescueFd)
	}

	if br.initTitle != "" {
		// one-time use of -t option
		br.title = br.initTitle
		br.initTitle = ""
	} else {
		br.title = title
	}
}

// closeBrowseFile releases the active display handle, including any replacement
// installed by R or rotation. A suspended parent's cleanup must not close a
// newer nested session. Advancing fileSeq also stops its reader from reopening.
func (br *browseObj) closeBrowseFile(seq uint64) {
	br.mutex.Lock()
	if br.fileSeq != seq {
		br.mutex.Unlock()
		return
	}
	fp, rescueFd := br.fp, br.rescueFd
	br.fp = nil
	br.rescueFd = 0
	br.fdLink = ""
	br.fileSeq++
	br.rereadPending = false
	br.readerAlive = false
	br.mutex.Unlock()
	if fp != nil {
		fp.Close()
	}
	if rescueFd > 0 {
		_ = unix.Close(rescueFd)
	}
}

// screenInit initializes terminal sizing information.
func (br *browseObj) screenInit(tty *os.File) {
	br.tty = tty

	var width, height int
	var err error

	if tty != nil {
		width, height, err = term.GetSize(int(tty.Fd()))
	}

	if tty == nil || err != nil {
		br.dispWidth = 80
		br.dispHeight = 25
	} else {
		br.dispWidth = width
		br.dispHeight = height
	}

	br.dispRows = br.dispHeight - 1
}

// vim: set ts=4 sw=4 noet:
