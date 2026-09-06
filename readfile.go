// readfile.go
// The go routine for reading files
//
// Copyright (c) 2024-2026 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// readInit resets reader state for a new or truncated file.
func readInit(br *browseObj, bytesRead *int64) {
	br.mapSiz = 1
	br.seekMap = []int64{0}
	br.sizeMap = []int64{0}
	br.newFileSiz = 0
	br.savFileSiz = 0
	br.newInode = 0
	br.savInode = 0
	*bytesRead = 0
}

// readFile continuously reads a file and updates line maps.
func readFile(br *browseObj, ch chan bool) {
	var bytesRead int64
	var err error
	initialRead := true

	// entered is set once the main loop begins. The reader honors a pending
	// R (rereadPending) as the first thing in every loop iteration, so once
	// entered is true a fresh reader is guaranteed to consume the request.
	var entered bool

	br.mutex.Lock()
	br.readerAlive = true
	readInit(br, &bytesRead)
	sourceFp := br.fp
	savFileName := br.fileName
	savFileSeq := br.fileSeq
	br.mutex.Unlock()

	// Guarantee R is never dropped: if this reader exits (I/O error, lost
	// rescue fd, ...) while an R is still pending, hand off to a successor
	// so the re-read always happens. Gated on entered to avoid a tight
	// relaunch loop on pre-loop init failures, which surface their own error.
	defer func() {
		br.mutex.Lock()
		br.readerAlive = false
		relaunch := entered && br.rereadPending && !br.fromStdin
		if relaunch {
			br.readerAlive = true
		}
		br.mutex.Unlock()
		if relaunch {
			go readFile(br, make(chan bool, 1))
		}
	}()

	dupFd, err := unix.Dup(int(sourceFp.Fd()))
	if err != nil {
		br.postMessage("Failed to duplicate file descriptor: "+err.Error(), MSG_RED)
		select {
		case ch <- false:
		default:
		}
		return
	}

	readerFp := os.NewFile(uintptr(dupFd), savFileName)
	if readerFp == nil {
		unix.Close(dupFd)
		br.postMessage("Failed to create file from descriptor", MSG_RED)
		select {
		case ch <- false:
		default:
		}
		return
	}
	defer func() {
		readerFp.Close()
	}()
	fd := int(readerFp.Fd())

	// Keep the reader buffer bounded: readLineLength consumes long lines in
	// fixed-size fragments instead of allocating a string as large as the line.
	bufReader := bufio.NewReaderSize(readerFp, READBUFSIZ)
	type lineMeta struct{ offset, length int64 }
	pendingLines := make([]lineMeta, 0, 1024)
	var postRereadRefresh bool
	var rereadDetected bool
	var partialLinePublished bool
	var partialLineOffset int64

	entered = true

	reopenReader := func(target string, manual bool) error {
		newFp, err := os.Open(target)
		if err != nil {
			return err
		}

		newDupFd, err := unix.Dup(int(newFp.Fd()))
		if err != nil {
			newFp.Close()
			return err
		}

		newReaderFp := os.NewFile(uintptr(newDupFd), target)
		if newReaderFp == nil {
			unix.Close(newDupFd)
			newFp.Close()
			return fmt.Errorf("failed to create file from descriptor")
		}

		br.mutex.Lock()
		if br.fileSeq != savFileSeq {
			br.mutex.Unlock()
			newReaderFp.Close()
			newFp.Close()
			return fmt.Errorf("file changed")
		}

		oldFp := br.fp
		oldReaderFp := readerFp
		br.fp = newFp
		br.fileName = target
		readInit(br, &bytesRead)
		// Only an explicit R (manual) owns rereadPending; an automatic
		// reopen (e.g. inode change) must not clear a user's pending
		// request, or the R would be treated as already satisfied.
		if manual {
			br.rereadPending = false
		}
		if br.rescueFd > 0 {
			unix.Close(br.rescueFd)
			br.rescueFd = 0
			br.fdLink = ""
		}
		br.mutex.Unlock()

		oldFp.Close()
		oldReaderFp.Close()
		readerFp = newReaderFp
		fd = int(readerFp.Fd())
		bufReader.Reset(readerFp)
		initialRead = true
		savFileName = target
		rereadDetected = false
		partialLinePublished = false
		partialLineOffset = 0
		return nil
	}

	for {
		// Get current filename snapshot under lock
		br.mutex.Lock()
		currentFileName := br.fileName
		currentFileSeq := br.fileSeq
		br.mutex.Unlock()

		if currentFileSeq != savFileSeq || currentFileName != savFileName {
			// new file -- exit thread
			return
		}

		// Check for manual re-read request
		br.mutex.Lock()
		pendingReread := br.rereadPending
		targetReread := br.absFileName
		if targetReread == "" {
			targetReread = br.fileName
		}
		br.mutex.Unlock()

		if pendingReread && targetReread != "" {
			if err := reopenReader(targetReread, true); err != nil {
				br.mutex.Lock()
				br.rereadPending = false
				br.pendingMsg = "Cannot re-open: " + err.Error()
				br.pendingMsgColor = MSG_RED
				br.mutex.Unlock()
				continue
			}

			postRereadRefresh = true
			continue
		}

		// Get file info using our filename snapshot.
		// If the file was deleted, fall back to the /proc fd link.
		newFileSiz, newInode, err := getFileInodeSize(currentFileName)
		if err != nil {
			br.mutex.Lock()
			br.scrollCancelPending = true
			rescueWasSet := br.rescueFd > 0
			if !rescueWasSet {
				rescueFd, dupErr := unix.Dup(fd)
				if dupErr == nil {
					br.rescueFd = rescueFd
					br.fdLink = fdLinkPath(rescueFd)
				}
			}

			fdLink := br.fdLink
			br.mutex.Unlock()

			if fdLink != "" {
				newFileSiz, newInode, err = getFileInodeSize(fdLink)
			}

			if err != nil {
				br.postMessage("Rescue fd link no longer accessible", MSG_RED)
				select {
				case ch <- false:
				default:
				}
				return
			}

			if !rescueWasSet {
				br.postMessage(fmt.Sprintf("File removed: recover from %s", fdLink), MSG_ORANGE)
				br.mutex.Lock()
				br.fileName = fdLink
				br.mutex.Unlock()
				savFileName = fdLink
			}
		}

		br.mutex.Lock()
		reopenForNewInode := br.savInode > 0 && newInode != br.savInode && currentFileName == savFileName
		br.mutex.Unlock()

		if reopenForNewInode {
			if err := reopenReader(currentFileName, false); err == nil {
				postRereadRefresh = true
				continue
			}
		}

		// Check if original file reappeared while in rescue mode
		if !rereadDetected {
			br.mutex.Lock()
			rescueActive := br.rescueFd > 0
			absFile := br.absFileName
			br.mutex.Unlock()

			if rescueActive && absFile != "" {
				if _, _, absErr := getFileInodeSize(absFile); absErr == nil {
					rereadDetected = true
					br.postMessage("File removed: press R to re-read", MSG_ORANGE)
				}
			}
		}

		var shouldRead bool

		br.mutex.Lock()
		if br.fileSeq != savFileSeq || br.fileName != savFileName {
			br.mutex.Unlock()
			return
		}

		// Store the file info we just retrieved
		br.newFileSiz = newFileSiz
		br.newInode = newInode
		stdinFinalPending := br.fromStdin && br.stdinEOF && bytesRead < br.newFileSiz

		// handleFileReset runs with br.mutex held, so it posts
		// display work by writing the pending fields directly
		handleFileReset := func(msg string) {
			if msg != "" {
				br.pendingMsg = msg
				br.pendingMsgColor = MSG_RED
			}
			readInit(br, &bytesRead)
			br.scrollCancelPending = true
			shouldRead = true
			initialRead = true
			postRereadRefresh = true
			partialLinePublished = false
			partialLineOffset = 0
		}

		if br.savInode > 0 && br.newInode != br.savInode {
			handleFileReset("")
		} else if br.newFileSiz < br.savFileSiz {
			handleFileReset("File truncated")
		} else {
			shouldRead = initialRead || br.savFileSiz < br.newFileSiz || stdinFinalPending
		}

		br.mutex.Unlock()

		if shouldRead {
			// Seek to the last known end of file (or beginning if truncated).
			readStart := bytesRead
			replacePartialLine := partialLinePublished
			if replacePartialLine {
				readStart = partialLineOffset
			}

			if _, err := readerFp.Seek(readStart, io.SeekStart); err != nil {
				select {
				case ch <- false:
				default:
				}
				return
			}

			readOffset := readStart
			bufReader.Reset(readerFp)
			pendingLines = pendingLines[:0]
			nextPartialPublished := false
			nextPartialOffset := int64(0)

			for {
				lineLen, hasNewline, err := readLineLength(bufReader)
				if err != nil {
					if err != io.EOF {
						// Report and exit for unexpected error
						select {
						case ch <- false:
						default:
						}
						return
					}
				}

				if lineLen == 0 {
					break
				}

				// Publish unterminated lines, but remember their offset so later
				// growth can replace the provisional entry instead of splitting it.

				if err == io.EOF && !hasNewline {
					br.mutex.Lock()
					fromStdin := br.fromStdin
					stdinDone := fromStdin && br.stdinEOF
					br.mutex.Unlock()
					if fromStdin && !stdinDone {
						break
					}
					nextPartialPublished = true
					nextPartialOffset = readOffset
				}

				readLen := lineLen
				if hasNewline {
					readLen--
				}

				cappedLen := min(readLen, READBUFSIZ)
				pendingLines = append(pendingLines, lineMeta{offset: readOffset, length: cappedLen})
				readOffset += lineLen

				if err == io.EOF {
					break
				}
			}

			br.mutex.Lock()
			if br.fileSeq != savFileSeq || br.fileName != savFileName {
				br.mutex.Unlock()
				return
			}
			if replacePartialLine && br.mapSiz > 1 && br.seekMap[br.mapSiz-1] == partialLineOffset {
				br.seekMap = br.seekMap[:br.mapSiz-1]
				br.sizeMap = br.sizeMap[:br.mapSiz-1]
				br.mapSiz--
			}
			for _, info := range pendingLines {
				br.seekMap = append(br.seekMap, info.offset)
				br.sizeMap = append(br.sizeMap, info.length)
				br.mapSiz++
			}
			if len(pendingLines) > 0 {
				br.hitEOF = false
			}
			br.savFileSiz = br.newFileSiz
			br.savInode = br.newInode
			br.mutex.Unlock()
			bytesRead = readOffset
			initialRead = false
			partialLinePublished = nextPartialPublished
			partialLineOffset = nextPartialOffset

			select {
			case ch <- true:
			default:
			}
		}

		if postRereadRefresh {
			postRereadRefresh = false
			br.mutex.Lock()
			// don't clobber a queued message, e.g. "File truncated"
			if br.pendingMsg == "" {
				br.pendingMsg = "Re-reading file"
				br.pendingMsgColor = MSG_GREEN
				br.pendingMsgTransient = true
			}
			br.refreshPending = true
			br.mutex.Unlock()
		}

		time.Sleep(time.Second)
	}
}

// readLineLength consumes one line without retaining its contents. Long lines
// are read in bufio-sized fragments, keeping memory use bounded by the reader
// buffer rather than by the line length.
func readLineLength(reader *bufio.Reader) (length int64, hasNewline bool, err error) {
	for {
		fragment, readErr := reader.ReadSlice('\n')
		length += int64(len(fragment))

		switch readErr {

		case nil:
			return length, true, nil

		case bufio.ErrBufferFull:
			continue

		case io.EOF:
			return length, false, io.EOF

		default:
			return length, false, readErr
		}
	}
}

// getFileInodeSize returns the size and inode for a filename.
func getFileInodeSize(filename string) (int64, uint64, error) {
	var stat unix.Stat_t

	err := unix.Stat(filename, &stat)
	if err != nil {
		return 0, 0, err
	}

	return stat.Size, stat.Ino, nil
}

// readStdin copies stdin into a temp file and returns true if empty.
func (br *browseObj) readStdin(fin, fout *os.File) bool {
	const copyBufSize = 64 * 1024

	buf := make([]byte, copyBufSize)
	bytesWritten, err := io.CopyBuffer(fout, fin, buf)
	if err != nil {
		// Surface a truncated stream rather than presenting it as complete.
		br.postMessage("Error reading standard input: "+err.Error(), MSG_RED)
	}
	return bytesWritten == 0
}

// readFromMap reads a line by index using the seek and size maps. Like
// lineIsMatch it reuses per-session scratch buffers rather than allocating
// per line, so it is only safe to call from the main goroutine and the
// returned bytes are valid only until the next call.
func (br *browseObj) readFromMap(lineno int) []byte {
	br.mutex.Lock()
	if lineno < 0 || lineno >= br.mapSiz || br.fp == nil {
		br.mutex.Unlock()
		return nil
	}

	seek, size := br.seekMap[lineno], br.sizeMap[lineno]

	// Make sure size is reasonable to avoid panics (16MB)
	if size < 0 || size > (16<<20) {
		br.mutex.Unlock()
		return nil
	}

	if int64(cap(br.readScratch)) < size {
		br.readScratch = make([]byte, size)
	}
	data := br.readScratch[:size]

	n, err := br.fp.ReadAt(data, seek)
	br.mutex.Unlock()
	if err != nil && err != io.EOF {
		return nil
	}

	expanded, scratch := expandTabs(data[:n], br.readTabScratch)
	br.readTabScratch = scratch
	return expanded
}

// vim: set ts=4 sw=4 noet:
