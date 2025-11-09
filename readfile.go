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
	"math"
	"os"
	"sync/atomic"
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
	var notified int32

	readInit(br, &bytesRead)

	fd := int(br.fp.Fd())
	dupFd, err := unix.Dup(fd)
	if err != nil {
		br.printMessage("Failed to duplicate file descriptor: "+err.Error(), MSG_RED)
		select {
		case ch <- false:
		default:
		}
		return
	}

	readerFp := os.NewFile(uintptr(dupFd), br.fileName)
	if readerFp == nil {
		unix.Close(dupFd)
		br.printMessage("Failed to create file from descriptor", MSG_RED)
		select {
		case ch <- false:
		default:
		}
		return
	}
	defer func() {
		readerFp.Close()
		unix.Close(dupFd)
	}()

	// Get initial filename with mutex protection
	br.mutex.Lock()
	saveFileName := br.fileName
	br.mutex.Unlock()

	for {
		// Check if file changed with mutex protection
		br.mutex.Lock()
		currentFileName := br.fileName
		br.mutex.Unlock()

		if currentFileName != saveFileName {
			// new file -- exit thread
			return
		}

		var newSize int64
		newSize, err = getFileSize(readerFp)
		if err != nil {
			select {
			case ch <- false:
			default:
			}
			return
		}

		var shouldRead bool
		br.mutex.Lock()
		br.newFileSiz = newSize

		// Handle file truncation
		if br.newFileSiz < br.savFileSiz {
			br.printMessage("File truncated", MSG_RED)
			readInit(br, &bytesRead)
			br.modeScroll = MODE_SCROLL_NONE
			br.shownMsg = false
			shouldRead = true
		} else {
			// Check if file grew
			shouldRead = br.savFileSiz == 0 || br.savFileSiz < br.newFileSiz
		}
		br.mutex.Unlock()

		if shouldRead {
			// Seek to the last known end of file (or beginning if truncated)
			_, err = readerFp.Seek(bytesRead, io.SeekStart)
			if err != nil {
				// Cannot seek, exit
				select {
				case ch <- false:
				default:
				}
				return
			}

			// Read new lines into temporary structures without holding the lock
			// to avoid blocking the UI.
			type lineInfo struct {
				seek int64
				size int64
			}
			var newLines []lineInfo
			currentOffset := bytesRead
			reader := bufio.NewReader(readerFp)

			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					break
				}
				lineLen := len(line)
				newLines = append(newLines, lineInfo{
					seek: currentOffset,
					size: int64(minimum(lineLen-1, READBUFSIZ)),
				})
				currentOffset += int64(lineLen)
			}

			// Now, lock and merge the temporary maps into the main ones
			if len(newLines) > 0 {
				br.mutex.Lock()
				for _, info := range newLines {
					br.seekMap[br.mapSiz] = info.seek
					br.sizeMap[br.mapSiz] = info.size
					br.mapSiz++
				}
				br.hitEOF = false
				br.savFileSiz = br.newFileSiz
				br.mutex.Unlock()
			}
			bytesRead = currentOffset

			if atomic.CompareAndSwapInt32(&notified, 0, 1) {
				select {
				case ch <- true:
				default:
				}
			}
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
		if err == io.EOF {
			if !empty && len(line) > 0 {
				w.WriteString(line)
			}

			return empty
		}

		if err != nil {
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
	// Use the maps to read a line from the file

	br.mutex.Lock()
	if lineno >= br.mapSiz {
		br.mutex.Unlock()
		return nil
	}

	seek := br.seekMap[lineno]
	size := br.sizeMap[lineno]
	br.mutex.Unlock()

	// Make sure size is reasonable to avoid panics (16MB)
	if size < 0 || size > (16<<20) {
		return nil
	}

	// Check for safe cast to int
	if size > int64(math.MaxInt) {
		return nil
	}

	data := make([]byte, int(size))
	_, err := br.fp.ReadAt(data, seek)
	if err != nil || len(data) == 0 {
		return nil
	}

	return expandTabs(data)
}

// vim: set ts=4 sw=4 noet:
