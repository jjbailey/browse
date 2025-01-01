// readfile.go
// the go routine for reading files
//
// Copyright (c) 2024 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"bufio"
	"errors"
	"io"
	"os"
	"sync"
	"syscall"
	"time"
)

var (
	mutex sync.Mutex
)

func readFile(br *browseObj, ch chan bool) {
	// initial read file, and continuously read for updates

	var bytesRead int64
	var notified bool
	var err error

	reader := bufio.NewReader(br.fp)

	if br.fromStdin {
		// wait for some input from stdin
		time.Sleep(time.Second)
	}

	br.newFileSiz, br.savFileSiz = 0, 0

	for {
		mutex.Lock()
		br.newFileSiz, err = getFileSize(br.fp)

		if err != nil {
			// fatal

			if !notified {
				ch <- false
			}

			mutex.Unlock()
			return
		}

		if br.newFileSiz < br.savFileSiz {
			// file shrunk -- reinitialize
			br.fileInit(br.fp, br.fileName, br.title, br.fromStdin)
			br.printMessage("File truncated", MSG_RED)

			// need to show the user
			br.modeScroll = MODE_SCROLL_NONE
			br.shownMsg = false

			// reset and fall through
			br.savFileSiz, bytesRead = 0, 0
		}

		if br.savFileSiz == 0 || br.savFileSiz < br.newFileSiz {
			// file unread or grew
			// read and map the new lines

			br.fp.Seek(br.seekMap[br.mapSiz], io.SeekStart)

			for {
				br.seekMap[br.mapSiz] = bytesRead
				line, err := reader.ReadString('\n')

				if err != nil {
					break
				}

				lineLen := len(line)
				bytesRead += int64(lineLen)
				br.sizeMap[br.mapSiz] = int64(minimum(lineLen-1, READBUFSIZ))
				br.mapSiz++
			}

			br.hitEOF = false

			if !notified {
				ch <- true
				notified = true
			}
		}

		br.savFileSiz, err = getFileSize(br.fp)
		mutex.Unlock()

		if err != nil {
			// fatal

			if !notified {
				ch <- false
			}

			return
		}

		time.Sleep(time.Second)
	}
}

func getFileSize(fp *os.File) (int64, error) {
	fInfo, err := fp.Stat()

	if err != nil {
		return 0, err
	}

	return fInfo.Size(), nil
}

func (x *browseObj) readStdin(fin, fout *os.File) {
	// read from stdin, write to temp file

	r := bufio.NewReader(fin)
	w := bufio.NewWriter(fout)

	for {
		line, err := r.ReadString('\n')

		if errors.Is(err, syscall.EPIPE) {
			x.saneExit()
		}

		if err == io.EOF {
			break
		}

		w.WriteString(line)
		w.Flush()
	}
}

func (x *browseObj) readFromMap(lineno int) []byte {
	// use the maps to read a line from the file

	mutex.Lock()
	defer mutex.Unlock()

	if lineno >= x.mapSiz {
		// should not happen
		return nil
	}

	data := make([]byte, x.sizeMap[lineno])
	_, err := x.fp.ReadAt(data, x.seekMap[lineno])

	if err != nil || len(data) == 0 {
		return nil
	}

	return expandTabs(data)
}

// vim: set ts=4 sw=4 noet:
