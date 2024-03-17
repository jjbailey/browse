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
	"syscall"
	"time"
)

func readFile(br *browseObj, ch chan bool) {
	// initial read file, and continuously read for updates

	var bytesRead int64
	var notified bool
	var newFileSiz, savFileSiz int64
	var err error

	reader := bufio.NewReader(br.fp)

	if br.fromStdin {
		// wait for some input from stdin
		time.Sleep(500 * time.Millisecond)
	}

	for {
		newFileSiz, err = getFileSize(br.fp)

		if err != nil || (br.stdinEOF && newFileSiz == 0) {
			// error or nothing to read
			if !notified {
				ch <- false
			}

			return
		}

		if newFileSiz == savFileSiz {
			// no change
			time.Sleep(2 * time.Second)
			continue
		}

		if newFileSiz < savFileSiz {
			// file shrunk -- reinitialize
			br.fileInit(br.fp, br.fileName, br.fromStdin)
			br.printMessage("File truncated")

			// need to show the user
			br.modeScrollDown = false
			br.modeScrollUp = false
			br.modeTail = false
			br.shownMsg = false

			// reset and fall through
			savFileSiz, bytesRead = 0, 0
		}

		if savFileSiz == 0 || savFileSiz < newFileSiz {
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

				// init the next map entry
				br.seekMap[br.mapSiz] = 0
				br.sizeMap[br.mapSiz] = 0
			}

			br.hitEOF = false

			if !notified {
				ch <- true
				notified = true
			}
		}

		savFileSiz, err = getFileSize(br.fp)

		if err != nil {
			if !notified {
				ch <- false
			}

			return
		}
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
			x.stdinEOF = true
			return
		}

		w.WriteString(line)
		w.Flush()
	}
}

func (x *browseObj) readFromMap(lineno int) ([]byte, int) {
	// use the maps to read a line from the file

	if lineno >= x.mapSiz {
		// should not happen
		return nil, 0
	}

	data := make([]byte, x.sizeMap[lineno])
	_, err := x.fp.ReadAt(data, x.seekMap[lineno])

	if err != nil {
		return nil, 0
	}

	if len(data) == 0 {
		return nil, 0
	}

	newdata, n := expandTabs(data)
	return newdata[x.shiftWidth:], n
}

// vim: set ts=4 sw=4 noet:
