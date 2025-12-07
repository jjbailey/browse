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
	br.newInode = 0
	br.savInode = 0
	*bytesRead = 0
}

func readFile(br *browseObj, ch chan bool) {
	// initial read file, and continuously read for updates

	var bytesRead int64
	var err error

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
	savFileName := br.fileName
	br.mutex.Unlock()

	for {
		// Check if file changed with mutex protection
		br.mutex.Lock()
		currentFileName := br.fileName
		br.mutex.Unlock()

		if currentFileName != savFileName {
			// new file -- exit thread
			return
		}

		br.newFileSiz, br.newInode, err = getFileInodeSize(currentFileName)
		if err != nil {
			select {
			case ch <- false:
			default:
			}
			return
		}

		var shouldRead bool

		br.mutex.Lock()

		handleFileReset := func(msg string) {
			br.printMessage(msg, MSG_RED)
			readInit(br, &bytesRead)
			br.modeScroll = MODE_SCROLL_NONE
			br.shownMsg = true
			shouldRead = true
		}

		if br.savInode > 0 && br.newInode != br.savInode {
			handleFileReset("File replaced")
		} else if br.newFileSiz < br.savFileSiz {
			handleFileReset("File truncated")
		} else {
			shouldRead = br.savFileSiz == 0 || br.savFileSiz < br.newFileSiz
		}

		br.mutex.Unlock()

		if shouldRead {
			// Seek to the last known end of file (or beginning if truncated)
			if _, err := readerFp.Seek(bytesRead, io.SeekStart); err != nil {
				select {
				case ch <- false:
				default:
				}
				return
			}

			readOffset := bytesRead
			bufReader := bufio.NewReader(readerFp)

			// We accumulate new lines' offsets/lengths before locking and merging
			type lineMeta struct{ offset, length int64 }
			var pendingLines []lineMeta

			for {
				line, err := bufReader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break
					}
					// Report and exit for unexpected error
					select {
					case ch <- false:
					default:
					}
					return
				}
				lineLen := len(line)
				cappedLen := int64(lineLen - 1)
				if cappedLen > READBUFSIZ {
					cappedLen = READBUFSIZ
				}
				pendingLines = append(pendingLines, lineMeta{offset: readOffset, length: cappedLen})
				readOffset += int64(lineLen)
			}

			if len(pendingLines) > 0 {
				br.mutex.Lock()
				for _, info := range pendingLines {
					br.seekMap[br.mapSiz] = info.offset
					br.sizeMap[br.mapSiz] = info.length
					br.mapSiz++
				}
				br.hitEOF = false
				br.savFileSiz = br.newFileSiz
				br.savInode = br.newInode
				br.mutex.Unlock()
			}
			bytesRead = readOffset

			select {
			case ch <- true:
			default:
			}
		}

		time.Sleep(time.Second)
	}
}

func getFileInodeSize(filename string) (int64, uint64, error) {
	// Returns size, inode, error for given filename

	var stat unix.Stat_t

	err := unix.Stat(filename, &stat)
	if err != nil {
		return 0, 0, err
	}

	return stat.Size, stat.Ino, nil
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
	if err != nil {
		return nil
	}

	return expandTabs(data)
}

// vim: set ts=4 sw=4 noet:
