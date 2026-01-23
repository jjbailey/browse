// completer.go
// central completion module
//
// Copyright (c) 2024-2026 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jjbailey/go-prompt"
)

// Completion search modes.
const (
	dispSuggestions = 5
	maxSuggestions  = 1000
	searchFiles     = 1
	searchPath      = 2
	searchSearch    = 3
	searchDirs      = 4
)

// Completion filters for file types.
const (
	onlyDirs  = 1
	onlyExec  = 2
	onlyFiles = 3
)

// SearchType controls which completion mode is active.
var SearchType int

// userDirComp prompts for a directory with completion.
func userDirComp() (string, bool) {
	SearchType = searchDirs
	return runCompleter("Dir: ", dirHistory)
}

// userFileComp prompts for a file with completion.
func userFileComp() (string, bool) {
	SearchType = searchFiles
	return runCompleter("File: ", fileHistory)
}

// userBashComp prompts for a command with PATH-aware completion.
func userBashComp() (string, bool) {
	SearchType = searchPath
	return runCompleter("$ ", commHistory)
}

// userSearchComp prompts for a search pattern with completion.
func userSearchComp(searchDir bool) (string, bool) {
	promptStr := "/"
	if !searchDir {
		promptStr = "?"
	}

	SearchType = searchSearch
	return runCompleter(promptStr, searchHistory)
}

// runCompleter starts the prompt UI and returns user input and cancellation state.
func runCompleter(promptStr, historyFile string) (string, bool) {
	history := loadHistory(historyFile)

	// Get hostname title
	title := getHostnameTitle()

	// reset go-prompt BackedOut flag
	prompt.BackedOut = false

	// set RawPrefix to allow escape chars in prefix, turns off color
	// escape char usage works only in simple cases, not here
	prompt.RawPrefix = true

	p := prompt.New(
		func(in string) { /* no-op executor */ },
		completer,
		prompt.OptionDescriptionBGColor(prompt.DarkGray),
		prompt.OptionDescriptionTextColor(prompt.Yellow),
		prompt.OptionHistory(history),
		prompt.OptionMaxSuggestion(dispSuggestions),
		prompt.OptionPrefix(promptStr),
		prompt.OptionScrollbarBGColor(prompt.DefaultColor),
		prompt.OptionScrollbarThumbColor(prompt.DefaultColor),
		prompt.OptionSelectedSuggestionBGColor(prompt.DarkGray),
		prompt.OptionSelectedSuggestionTextColor(prompt.Yellow),
		prompt.OptionSwitchKeyBindMode(prompt.EmacsKeyBind),
		prompt.OptionTitle(title),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn: func(buf *prompt.Buffer) {
				fmt.Print("\r" + CURUP)
			},
		}),
	)

	input := p.Input()
	fmt.Printf(XTERMTITLE, title)
	ttyBrowser()

	if len(input) == 0 {
		return "", prompt.BackedOut
	}

	return input, false
}

// getHostnameTitle returns a short hostname for display in the prompt title.
func getHostnameTitle() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		return "unknown"
	}

	if dotIndex := strings.IndexByte(hostname, '.'); dotIndex != -1 {
		hostname = hostname[:dotIndex]
	}

	return hostname
}

// completer provides suggestions based on the current SearchType and input.
func completer(d prompt.Document) []prompt.Suggest {
	word := strings.ReplaceAll(d.GetWordBeforeCursor(), "//", "/")
	originalWord := word

	// Handle tilde expansions early to prevent duplicated code

	if strings.HasPrefix(word, "~-") {
		word = prevDirectory()
	}

	if strings.HasPrefix(word, "~") {
		word = expandHome(word)
	}

	switch SearchType {

	case searchSearch:
		return searchCompleter()

	case searchDirs:
		return dirCompleter(word)

	case searchFiles:
		if hasPathSeparator(word) {
			return fileCompleter(word)
		}
		return anyCompleter(".", originalWord, onlyFiles)

	case searchPath:
		if hasPathSeparator(word) {
			return fileCompleter(word)
		}

		// Current user input before the cursor
		text := d.TextBeforeCursor()

		// If just at the very start
		if strings.TrimSpace(text) == "" {
			return pathCompleter("")
		}

		// If currently typing (partial) after a pipe: "|cmd"
		if strings.HasPrefix(word, "|") {
			if word == "|" {
				return nil
			}

			suggestions := pathCompleter(word[1:])
			for i := range suggestions {
				suggestions[i].Text = "|" + suggestions[i].Text
			}
			return suggestions
		}

		// Split on whitespace for tokens *before* current cursor position
		parts := strings.Fields(text)
		numParts := len(parts)

		// If previous *token* is a pipe
		if numParts > 0 && ((word == "" && parts[numParts-1] == "|") ||
			(word != "" && numParts > 1 && parts[numParts-2] == "|")) {
			return pathCompleter(word)
		}

		// If second or later word supplied, drop to anyCompleter
		if numParts > 1 || (numParts == 1 && word == "") {
			return anyCompleter(".", originalWord, onlyFiles)
		}

		// By default, normal path completion
		return pathCompleter(word)

	default:
		return anyCompleter(".", originalWord, onlyFiles)
	}
}

// hasPathSeparator reports whether the word contains a path separator.
func hasPathSeparator(word string) bool {
	return strings.Contains(word, "/")
}

// searchCompleter returns suggestions for search history (currently none).
func searchCompleter() []prompt.Suggest {
	return nil
}

// fileCompleter completes file paths for a given word.
func fileCompleter(word string) []prompt.Suggest {
	dir := filepath.Dir(word)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	prefix := word
	if !strings.HasSuffix(word, "/") {
		prefix = filepath.Base(word)
	} else {
		prefix = ""
	}

	return matchFiles(files, dir, prefix, true, onlyFiles)
}

// pathCompleter completes executable names from PATH entries.
func pathCompleter(word string) []prompt.Suggest {
	var suggestions []prompt.Suggest

	// cannot test PATH with go run .
	// go build .

	paths := strings.Split(os.Getenv("PATH"), ":")
	if len(paths) == 0 || (len(paths) == 1 && paths[0] == "") {
		paths = []string{"/usr/local/bin", "/usr/bin", "/usr/sbin"}
	}

	for _, dir := range paths {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		suggestions = append(suggestions,
			matchFiles(files, dir, word, false, onlyExec)...)

		if len(suggestions) > maxSuggestions {
			suggestions = suggestions[:maxSuggestions]
			break
		}
	}

	return suggestions
}

// dirCompleter completes directories using absolute paths and CDPATH.
func dirCompleter(word string) []prompt.Suggest {
	var suggestions []prompt.Suggest

	if hasPathSeparator(word) {
		dir := filepath.Dir(word)
		files, err := os.ReadDir(dir)
		if err != nil {
			return nil
		}

		prefix := word
		if !strings.HasSuffix(word, "/") {
			prefix = filepath.Base(word)
		} else {
			prefix = ""
		}

		suggestions = matchFiles(files, dir, prefix, true, onlyDirs)

		if len(suggestions) > maxSuggestions {
			suggestions = suggestions[:maxSuggestions]
			return suggestions
		}
	}

	// Handle other paths

	paths := strings.Split(os.Getenv("CDPATH"), ":")
	if len(paths) == 0 || (len(paths) == 1 && paths[0] == "") {
		paths = []string{".", ".."}
	}

	for _, dir := range paths {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		suggestions = append(suggestions,
			matchFiles(files, dir, word, true, onlyDirs)...)

		if len(suggestions) > maxSuggestions {
			suggestions = suggestions[:maxSuggestions]
			break
		}
	}

	return suggestions
}

// anyCompleter completes entries in a directory with optional filtering.
func anyCompleter(dir, prefix string, onlyType int) []prompt.Suggest {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	// Match all directory entries
	return matchFiles(files, dir, prefix, false, onlyType)
}

// matchFiles filters directory entries into prompt suggestions.
func matchFiles(files []os.DirEntry, dir, prefix string,
	useFullPath bool, onlyType int) []prompt.Suggest {

	suggestions := make([]prompt.Suggest, 0, minimum(len(files), maxSuggestions))

	for _, file := range files {
		if len(suggestions) >= maxSuggestions {
			break
		}

		name := file.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}

		fullPath := filepath.Join(dir, name)

		// Handle onlyDirs filter
		if onlyType == onlyDirs {
			isDir := file.IsDir()
			isSymlink := file.Type()&os.ModeSymlink != 0
			if !isDir && !isSymlink {
				continue
			}

			if isSymlink {
				// Stat the link target
				fi, err := os.Stat(fullPath)
				if err != nil || !fi.IsDir() {
					continue
				}
			}
		}

		// If filtering for executables, only stat if not directory
		if onlyType == onlyExec && !file.IsDir() {
			info, err := file.Info()
			if err != nil {
				continue
			}

			mode := info.Mode().Perm()
			if mode&0111 == 0 {
				continue
			}
		}

		// Determine display name (quote if contains spaces)
		var displayName string

		if useFullPath {
			displayName = fullPath
		} else {
			displayName = name
		}

		if strings.Contains(displayName, " ") {
			if strings.Contains(displayName, "'") {
				displayName = "\"" + displayName + "\""
			} else {
				displayName = "'" + displayName + "'"
			}
		}

		var desc string

		switch {

		case file.Type()&os.ModeSymlink != 0:
			linkTarget, _ := resolveSymlink(fullPath)
			desc = "-> " + linkTarget

		case file.IsDir():
			desc = "directory"

		case file.Type()&os.ModeNamedPipe != 0:
			desc = "named pipe"

		case file.Type()&os.ModeSocket != 0:
			desc = "socket"

		case file.Type()&os.ModeCharDevice != 0:
			desc = "character device"

		case file.Type()&os.ModeDevice != 0:
			desc = "block device"

		default:
			desc = ""
		}

		suggestions = append(suggestions, prompt.Suggest{
			Text:        displayName,
			Description: desc,
		})
	}

	return suggestions
}

// vim: set ts=4 sw=4 noet:
