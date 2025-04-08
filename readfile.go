// readfile.go
// the go routine for reading files
//
// Copyright (c) 2024-2025 jjb
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
		// Get file size with minimal mutex lock
		mutex.Lock()
		br.newFileSiz, err = getFileSize(br.fp)
		if err != nil {
			if !notified {
				ch <- false
			}
			mutex.Unlock()
			return
		}
		mutex.Unlock()

		// Handle file truncation
		if br.newFileSiz < br.savFileSiz {
			mutex.Lock()
			br.fileInit(br.fp, br.fileName, br.title, br.fromStdin)
			br.printMessage("File truncated", MSG_RED)
			br.modeScroll = MODE_SCROLL_NONE
			br.shownMsg = false
			br.savFileSiz, bytesRead = 0, 0
			mutex.Unlock()
		}

		// Read new content if file grew
		if br.savFileSiz == 0 || br.savFileSiz < br.newFileSiz {
			mutex.Lock()
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
			br.savFileSiz = br.newFileSiz

			if !notified {
				ch <- true
				notified = true
			}
			mutex.Unlock()
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

func (br *browseObj) readStdin(fin, fout *os.File) {
	// read from stdin, write to temp file

	r := bufio.NewReader(fin)
	w := bufio.NewWriter(fout)

	for {
		line, err := r.ReadString('\n')

		if errors.Is(err, syscall.EPIPE) {
			br.saneExit()
		}

		if err == io.EOF {
			break
		}

		w.WriteString(line)
		w.Flush()
	}
}

func (br *browseObj) readFromMap(lineno int) []byte {
	// use the maps to read a line from the file

	mutex.Lock()
	defer mutex.Unlock()

	if lineno >= br.mapSiz {
		// should not happen
		return nil
	}

	data := make([]byte, br.sizeMap[lineno])
	_, err := br.fp.ReadAt(data, br.seekMap[lineno])

	if err != nil || len(data) == 0 {
		return nil
	}

	return expandTabs(data)
}

// vim: set ts=4 sw=4 noet:
