// readfile.go
// the go routine for reading files
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"bufio"
	"io"
	"os"
	"time"
)

func readFile(br *browseObj, ch chan bool) {
	var notified bool
	var bytesRead int64
	var savFileSiz, newFileSiz int64

	reader := bufio.NewReader(br.fp)

	if br.fromStdin {
		// wait a sec for input from stdin
		time.Sleep(1 * time.Second)
	}

	for {
		fInfo, _ := br.fp.Stat()
		newFileSiz = fInfo.Size()

		if newFileSiz < savFileSiz {
			// file shrunk

			br.fileInit(br.fp, br.fileName, br.fromStdin)
			br.printMessage("File truncated")

			// reset and fall through
			savFileSiz = 0
			bytesRead = 0
		}

		if savFileSiz == 0 || savFileSiz < newFileSiz {
			// file unread or grew

			br.fp.Seek(br.seekMap[br.mapSiz], io.SeekStart)

			for {
				br.seekMap[br.mapSiz] = bytesRead
				line, err := reader.ReadString('\n')

				if err != nil {
					br.hitEOF = true
					break
				}

				lineLen := len(line)
				bytesRead += int64(lineLen)
				br.sizeMap[br.mapSiz] = int64(minimum(lineLen-1, READBUFSIZ))
				br.mapSiz++

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

func readStdin(fin, fout *os.File) {
	// read from stdin, write to temp file

	r := bufio.NewReader(fin)
	w := bufio.NewWriter(fout)

	for {
		line, err := r.ReadString('\n')

		if err == io.EOF {
			break
		}

		w.WriteString(line)
	}

	w.Flush()
}

func (x *browseObj) readFromMap(lineno int) ([]byte, int) {
	data := make([]byte, x.sizeMap[lineno])
	x.fp.Seek(x.seekMap[lineno], io.SeekStart)
	x.fp.Read(data)
	return expandTabs(data)
}

// vim: set ts=4 sw=4 noet:
