// readfile.go
// the go routine for reading files
//
// Copyright (c) 2024 jjb
// All rights reserved.

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
	var savFileSiz int64 = 0
	var newFileSiz int64 = 0

	reader := bufio.NewReader(br.fp)

	if br.fromStdin {
		// wait a sec for input from stdin
		time.Sleep(1 * time.Second)
	}

	for {
		fInfo, err := br.fp.Stat()

		if err != nil {
			if !notified {
				ch <- false
			}

			return
		}

		newFileSiz = fInfo.Size()

		if newFileSiz < savFileSiz {
			// file shrunk -- reinitialize
			br.fileInit(br.fp, br.fileName, br.fromStdin)
			br.printMessage("File truncated")

			// reset and fall through
			savFileSiz = 0
			bytesRead = 0

			// need to show the user
			br.modeScrollUp = false
			br.modeScrollDown = false
			br.modeTail = false
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

		fInfo, _ = br.fp.Stat()
		savFileSiz = fInfo.Size()
		time.Sleep(2 * time.Second)
	}
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
	}

	w.Flush()
}

func (x *browseObj) readFromMap(lineno int) ([]byte, int) {
	// use the maps to read a line from the file
	// don't read more than we can display

	if lineSize, ok := x.sizeMap[lineno]; ok && lineSize > 0 {
		lineSize = int64(minimum(int(lineSize), x.dispWidth))
		data := make([]byte, lineSize)
		x.fp.ReadAt(data, x.seekMap[lineno])
		newdata, n := expandTabs(data)
		return newdata[x.shiftWidth:], n
	}

	return nil, 0
}

// vim: set ts=4 sw=4 noet:
