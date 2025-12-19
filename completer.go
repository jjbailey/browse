// completer.go
// central completion module
//
// Copyright (c) 2024-2025 jjb
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

const (
	dispSuggestions = 5
	maxSuggestions  = 1000
	searchFiles     = 1
	searchPath      = 2
	searchSearch    = 3
	searchDirs      = 4
)

const (
	onlyDirs  = 1
	onlyExec  = 2
	onlyFiles = 3
)

var SearchType int

func userDirComp() (string, bool) {
	SearchType = searchDirs
	return runCompleter("Dir: ", dirHistory)
}

func userFileComp() (string, bool) {
	SearchType = searchFiles
	return runCompleter("File: ", fileHistory)
}

func userBashComp() (string, bool) {
	SearchType = searchPath
	return runCompleter("$ ", commHistory)
}

func userSearchComp(searchDir bool) (string, bool) {
	promptStr := "/"
	if !searchDir {
		promptStr = "?"
	}

	SearchType = searchSearch
	return runCompleter(promptStr, searchHistory)
}

func runCompleter(promptStr, historyFile string) (string, bool) {
	// Load history from file
	history := loadHistory(historyFile)

	// Get hostname title
	title := getHostnameTitle()

	// reset go-prompt BackedOut flag
	prompt.BackedOut = false

	p := prompt.New(
		func(in string) { /* no-op executor */ },
		completer,
		prompt.OptionDescriptionBGColor(prompt.DarkGray),
		prompt.OptionDescriptionTextColor(prompt.Yellow),
		prompt.OptionHistory(history),
		prompt.OptionMaxSuggestion(dispSuggestions),
		prompt.OptionPrefix(promptStr),
		prompt.OptionPrefixTextColor(prompt.White),
		prompt.OptionScrollbarBGColor(prompt.DefaultColor),
		prompt.OptionScrollbarThumbColor(prompt.DefaultColor),
		prompt.OptionSelectedSuggestionBGColor(prompt.DarkGray),
		prompt.OptionSelectedSuggestionTextColor(prompt.Yellow),
		prompt.OptionSuggestionTextColor(prompt.White),
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

func getHostnameTitle() string {
	// Can't prevent prompt.New from meddling with the title, so...

	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		return "unknown"
	}

	if dotIndex := strings.IndexByte(hostname, '.'); dotIndex != -1 {
		hostname = hostname[:dotIndex]
	}

	return hostname
}

func completer(d prompt.Document) []prompt.Suggest {
	word := strings.ReplaceAll(d.GetWordBeforeCursor(), "//", "/")
	originalWord := word

	// Handle ~ expansion early to prevent duplicated code
	if strings.HasPrefix(word, "~") {
		word = expandHome(word)
	}

	switch SearchType {

	case searchSearch:
		return searchCompleter()

	case searchDirs:
		return dirCompleter(word)

	case searchFiles:
		if isAbsOrRelPath(word) {
			return fileCompleter(word)
		}
		return anyCompleter(".", originalWord, false, onlyFiles)

	case searchPath:
		if isAbsOrRelPath(word) {
			return fileCompleter(word)
		}

		// Path completions at beginning of input or after |
		if !strings.Contains(d.TextBeforeCursor(), " ") ||
			len(word) == 2 || strings.TrimSuffix(word, " ") == "|" ||
			strings.HasPrefix(word, "|") {
			// Remove leading | if present (for pipes)
			if strings.HasPrefix(word, "|") {
				return pathCompleter(word[1:])
			}
			return pathCompleter(word)
		}
		return anyCompleter(".", originalWord, false, onlyFiles)
	}

	return anyCompleter(".", originalWord, false, onlyFiles)
}

func isAbsOrRelPath(word string) bool {
	// Only care about the very first rune for . and ..

	return strings.HasPrefix(word, "/") ||
		strings.HasPrefix(word, "./") ||
		strings.HasPrefix(word, "../") ||
		strings.Contains(word, "/")
}

func searchCompleter() []prompt.Suggest {
	// history contents only
	return nil
}

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

func pathCompleter(word string) []prompt.Suggest {
	// Handle files in PATH

	var suggestions []prompt.Suggest

	// cannot test PATH with go run .
	// go build .

	paths := strings.Split(os.Getenv("PATH"), ":")
	if len(paths) == 0 || (len(paths) == 1 && paths[0] == "") {
		paths = []string{"/usr/local/bin", "/usr/bin"}
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

func dirCompleter(word string) []prompt.Suggest {
	// Handle absolute path completions

	var suggestions []prompt.Suggest

	if strings.HasPrefix(word, "/") || strings.HasPrefix(word, "./") ||
		strings.HasPrefix(word, "../") || strings.Contains(word, "/") {

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

func anyCompleter(dir, prefix string, useFullPath bool, onlyType int) []prompt.Suggest {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	// Match all directory entries
	return matchFiles(files, dir, prefix, useFullPath, onlyType)
}

func matchFiles(files []os.DirEntry, dir, prefix string,
	useFullPath bool, onlyType int) []prompt.Suggest {

	suggestions := make([]prompt.Suggest, 0, dispSuggestions)
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
				// Stat the resolved link
				resolved, err := filepath.EvalSymlinks(fullPath)
				if err != nil {
					continue
				}
				fi, err := os.Stat(resolved)
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

		if strings.ContainsAny(displayName, " ") {
			displayName = "'" + displayName + "'"
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
