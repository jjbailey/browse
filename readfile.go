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
	"io"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

func readInit(br *browseObj, bytesRead *int64) {
	// reader initializations for new and truncated files

	br.mapSiz = 1
	br.seekMap = map[int]int64{0: 0}
	br.sizeMap = map[int]int64{0: 0}
	br.newFileSiz = 0
	br.savFileSiz = 0
	*bytesRead = 0
}

func readFile(br *browseObj, ch chan bool) {
	// initial read file, and continuously read for updates

	var bytesRead int64
	var err error
	var notified bool

	readInit(br, &bytesRead)

	fd := int(br.fp.Fd())
	dupFd, err := unix.Dup(fd)
	if err != nil {
		return
	}

	readerFp := os.NewFile(uintptr(dupFd), br.fileName)
	defer readerFp.Close()

	reader := bufio.NewReader(readerFp)

	// Get initial filename
	saveFileName := br.fileName

	for {
		// Check if file changed before doing anything else
		br.mutex.Lock()
		if br.fileName != saveFileName {
			// new file -- exit thread
			br.mutex.Unlock()
			return
		}
		br.mutex.Unlock()

		// Get file size with minimal mutex lock
		br.mutex.Lock()
		br.newFileSiz, err = getFileSize(readerFp)
		if err != nil {
			if !notified {
				ch <- false
			}
			br.mutex.Unlock()
			return
		}
		br.mutex.Unlock()

		// Handle file truncation
		if br.newFileSiz < br.savFileSiz {
			br.printMessage("File truncated", MSG_RED)
			readInit(br, &bytesRead)
			br.modeScroll = MODE_SCROLL_NONE
			br.shownMsg = false
		}

		// Read new content if file grew
		if br.savFileSiz == 0 || br.savFileSiz < br.newFileSiz {
			br.mutex.Lock()
			readerFp.Seek(br.seekMap[br.mapSiz], io.SeekStart)

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
			br.mutex.Unlock()
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

func (br *browseObj) readStdin(fin, fout *os.File) bool {
	// read from stdin, write to temp file

	r := bufio.NewReader(fin)
	w := bufio.NewWriter(fout)
	defer w.Flush()

	empty := true

	for {
		line, err := r.ReadString('\n')
		if err == io.EOF && empty {
			return empty
		}

		if _, err := w.WriteString(line); err != nil {
			return empty
		}

		if err := w.Flush(); err != nil {
			return empty
		}

		empty = false
	}
}

func (br *browseObj) readFromMap(lineno int) []byte {
	// use the maps to read a line from the file

	br.mutex.Lock()
	defer br.mutex.Unlock()

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
