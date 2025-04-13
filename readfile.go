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
	mutex sync.RWMutex
)

// FileReader handles the file reading operations
type FileReader struct {
	reader    *bufio.Reader
	file      *os.File
	notified  bool
	stopChan  chan struct{}
	errorChan chan error
}

// NewFileReader creates a new FileReader instance
func NewFileReader(fp *os.File) *FileReader {
	return &FileReader{
		reader:    bufio.NewReader(fp),
		file:      fp,
		stopChan:  make(chan struct{}),
		errorChan: make(chan error, 1),
	}
}

func (fr *FileReader) Stop() {
	// signals the reader to stop
	close(fr.stopChan)
}

func readFile(br *browseObj, ch chan bool) {
	// initial read file, and continuously read for updates

	var bytesRead int64
	var err error

	fr := NewFileReader(br.fp)
	defer func() {
		fr.Stop()
		close(ch)
	}()

	if br.fromStdin {
		// wait for some input from stdin
		time.Sleep(time.Second)
	}

	br.newFileSiz, br.savFileSiz = 0, 0

	for {
		select {

		case <-fr.stopChan:
			return

		default:
			// get file size with minimal mutex lock
			mutex.RLock()
			br.newFileSiz, err = getFileSize(br.fp)
			mutex.RUnlock()

			if err != nil {
				if !fr.notified {
					ch <- false
					fr.errorChan <- err
				}

				return
			}

			// handle file truncation
			if br.newFileSiz < br.savFileSiz {
				mutex.Lock()

				if err := br.handleFileTruncation(); err != nil {
					mutex.Unlock()
					fr.errorChan <- err
					return
				}

				bytesRead = 0
				mutex.Unlock()
			}

			// read new content if file grew
			if br.savFileSiz == 0 || br.savFileSiz < br.newFileSiz {
				if err := br.readNewContent(fr, &bytesRead); err != nil {
					fr.errorChan <- err
					return
				}

				if !fr.notified {
					ch <- true
					fr.notified = true
				}
			}

			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (br *browseObj) handleFileTruncation() error {
	br.fileInit(br.fp, br.fileName, br.title, br.fromStdin)
	br.printMessage("File truncated", MSG_RED)
	br.modeScroll = MODE_SCROLL_NONE
	br.shownMsg = false
	br.savFileSiz = 0
	return nil
}

func (br *browseObj) readNewContent(fr *FileReader, bytesRead *int64) error {
	mutex.Lock()
	defer mutex.Unlock()

	_, err := br.fp.Seek(br.seekMap[br.mapSiz], io.SeekStart)

	if err != nil {
		return err
	}

	for {
		br.seekMap[br.mapSiz] = *bytesRead
		line, err := fr.reader.ReadString('\n')

		if err != nil {
			if err != io.EOF {
				return err
			}

			break
		}

		lineLen := len(line)
		*bytesRead += int64(lineLen)
		br.sizeMap[br.mapSiz] = int64(minimum(lineLen-1, READBUFSIZ))
		br.mapSiz++
	}

	br.hitEOF = false
	br.savFileSiz = br.newFileSiz
	return nil
}

func getFileSize(fp *os.File) (int64, error) {
	fInfo, err := fp.Stat()

	if err != nil {
		return 0, err
	}

	return fInfo.Size(), nil
}

func (br *browseObj) readStdin(fin, fout *os.File) error {
	// read from stdin, write to temp file

	r := bufio.NewReader(fin)
	w := bufio.NewWriter(fout)
	defer w.Flush()

	for {
		line, err := r.ReadString('\n')

		if err != nil {
			if errors.Is(err, syscall.EPIPE) {
				br.saneExit()
			}

			if err == io.EOF {
				break
			}

			return err
		}

		if _, err := w.WriteString(line); err != nil {
			return err
		}

		if err := w.Flush(); err != nil {
			return err
		}
	}

	return nil
}

func (br *browseObj) readFromMap(lineno int) []byte {
	// use the maps to read a line from the file

	mutex.RLock()
	defer mutex.RUnlock()

	if lineno >= br.mapSiz {
		return nil
	}

	size := br.sizeMap[lineno]
	if size <= 0 || size > READBUFSIZ {
		return nil
	}

	data := make([]byte, size)
	n, err := br.fp.ReadAt(data, br.seekMap[lineno])

	if err != nil || n == 0 {
		return nil
	}

	return expandTabs(data)
}

// vim: set ts=4 sw=4 noet:
