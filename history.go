// history.go
// history management module
//
// Copyright (c) 2024-2026 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var historyWarningOnce sync.Once

func warnHistoryPersistence(err error) {
	historyWarningOnce.Do(func() {
		fmt.Fprintf(os.Stderr, "browse: history persistence unavailable: %v\n", err)
	})
}

// loadHistory reads the history file and returns recent entries.
func loadHistory(historyFile string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		warnHistoryPersistence(err)
		return []string{}
	}

	historyPath := filepath.Join(home, RCDIRNAME, historyFile)
	file, err := os.OpenFile(historyPath, os.O_RDONLY, 0600)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// no history yet; saveHistory creates the file
			return []string{}
		}
		warnHistoryPersistence(err)
		return []string{}
	}
	defer file.Close()

	var history []string

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for scanner.Scan() {
		history = append(history, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		warnHistoryPersistence(err)
		return []string{}
	}
	if historyFile == fileHistory || historyFile == dirHistory {
		for i, entry := range history {
			history[i] = normalizePathHistoryEntry(entry)
		}
	}

	// If history is too large, keep only the most recent entries
	if len(history) > maxHistorySize {
		history = history[len(history)-maxHistorySize:]
	}

	return history
}

// saveHistory writes history entries to disk with trimming and cleanup.
func saveHistory(history []string, historyFile string) {
	if len(history) == 0 {
		return
	}

	// Clean history in-place to avoid new slice allocation
	n := 0
	for _, entry := range history {
		if historyFile != searchHistory {
			entry = strings.TrimSpace(entry)
		}
		if entry != "" {
			history[n] = entry
			n++
		}
	}
	history = history[:n]

	if len(history) == 0 {
		return
	}

	// Ensure history doesn't exceed max size
	if len(history) > maxHistorySize {
		history = history[len(history)-maxHistorySize:]
	}

	home, err := os.UserHomeDir()
	if err != nil {
		warnHistoryPersistence(err)
		return
	}

	historyPath := filepath.Join(home, RCDIRNAME, historyFile)
	file, err := os.OpenFile(historyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		warnHistoryPersistence(err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, cmd := range history {
		fmt.Fprintln(writer, cmd)
	}
	if err := writer.Flush(); err != nil {
		warnHistoryPersistence(err)
	}
}

// updateDirHistory records directory changes in the directory history.
func updateDirHistory(savDir, curDir string) {
	// save the previous directory
	updateHistory(savDir, dirHistory)

	// save the current directory
	updateHistory(curDir, dirHistory)
}

// updateHistory appends a single entry to the specified history file.
func updateHistory(newEntry, historyFile string) {
	if historyFile != searchHistory {
		newEntry = strings.TrimSpace(newEntry)
	}

	if newEntry == "" {
		return
	}

	switch historyFile {

	case commHistory:
		newEntry = unQuote(newEntry)

	case searchHistory:
		// no quoting/unquoting: a search pattern's leading/trailing quotes
		// (if any) are part of the pattern, not shell quoting

	default:
		newEntry = quotePathHistoryEntry(newEntry)
	}

	history := loadHistory(historyFile)

	// Remove duplicate consecutive entries
	if len(history) > 0 && history[len(history)-1] == newEntry {
		return
	}

	history = append(history, newEntry)
	saveHistory(history, historyFile)
}

// quotePathHistoryEntry stores a path so fieldsQuoted can replay it exactly.
func quotePathHistoryEntry(path string) string {
	if strings.ContainsAny(path, "'\"") {
		// Double quotes preserve apostrophes. Escape double quotes and
		// backslashes so fieldsQuoted can replay the path exactly.
		return `"` + strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(path) + `"`
	}
	if strings.ContainsAny(path, " \t") {
		return "'" + path + "'"
	}
	return path
}

// normalizePathHistoryEntry preserves old unescaped paths and accepts entries
// already stored in the quoted format.
func normalizePathHistoryEntry(entry string) string {
	fields := fieldsQuoted(entry)
	if len(fields) == 1 {
		quoted := quotePathHistoryEntry(fields[0])
		wrapped := len(entry) >= 2 &&
			((entry[0] == '\'' && entry[len(entry)-1] == '\'') ||
				(entry[0] == '"' && entry[len(entry)-1] == '"'))
		if quoted == entry || wrapped {
			return quoted
		}
	}
	return quotePathHistoryEntry(entry)
}

// vim: set ts=4 sw=4 noet:
