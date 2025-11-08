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
)

var searchType int

func userFileComp() (string, bool) {
	searchType = searchFiles
	return runCompleter("File: ", fileHistory)
}

func userBashComp() (string, bool) {
	searchType = searchPath
	return runCompleter("$ ", commHistory)
}

func userSearchComp(searchDir bool) (string, bool) {
	promptStr := "/"
	if !searchDir {
		promptStr = "?"
	}

	searchType = searchSearch
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
	// Use local variable for efficiency, also avoids race if global is modified mid-call
	switch searchType {

	case searchSearch:
		return searchCompleter()

	case searchFiles:
		// Fall through below for word handling

	case searchPath:
		// Fall through below for word handling
	}

	word := strings.ReplaceAll(d.GetWordBeforeCursor(), "//", "/")
	originalWord := word

	// Home directory expansion (including ~user)
	if strings.HasPrefix(word, "~") {
		word = expandHome(word)
	}

	// Fast path: raw path or file completion
	if isAbsOrRelPath(word) {
		return fileCompleter(word)
	}

	// Command completion only if searchType is path and word is at prompt start
	if searchType == searchPath && !strings.Contains(d.TextBeforeCursor(), " ") {
		return pathCompleter(word)
	}

	// Fallback: Just use whatever word we've got in current directory
	return anyCompleter(".", originalWord, false, false)
}

func expandHome(word string) string {
	if len(word) == 1 || word[1] == '/' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return word
		}

		return filepath.Join(homeDir, word[1:])
	}

	// ~username or ~username/something
	slashIndex := strings.IndexRune(word, '/')
	var userName, pathSuffix string
	if slashIndex == -1 {
		userName = word[1:]
		pathSuffix = ""
	} else {
		userName = word[1:slashIndex]
		pathSuffix = word[slashIndex:]
	}

	wordHome, err := getHomeDir(userName)
	if err != nil || wordHome == "" {
		return word
	}

	// Confirm the directory exists, stat once here
	if fi, err := os.Stat(wordHome); err != nil || !fi.IsDir() {
		return word
	}

	return filepath.Join(wordHome, pathSuffix)
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

	// Only stat once; if fails, bail
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

	return matchFiles(files, dir, prefix, true, false)
}

func pathCompleter(word string) []prompt.Suggest {
	paths := strings.Split(os.Getenv("PATH"), ":")
	var suggestions []prompt.Suggest

	for _, dir := range paths {
		// Defensive: avoid unnecessary os.ReadDir
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		if len(suggestions) >= maxSuggestions {
			break
		}

		suggestions = append(suggestions, matchFiles(files, dir, word, false, true)...)
		// Check again in case matchFiles added a lot
		if len(suggestions) >= maxSuggestions {
			break
		}
	}

	if len(suggestions) > maxSuggestions {
		suggestions = suggestions[:maxSuggestions]
	}

	return suggestions
}

func anyCompleter(dir, prefix string, useFullPath bool, onlyExec bool) []prompt.Suggest {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	return matchFiles(files, dir, prefix, useFullPath, onlyExec)
}

func matchFiles(files []os.DirEntry, dir, prefix string, useFullPath, onlyExec bool) []prompt.Suggest {
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
		// Only stat (expensive) if necessary
		var desc string

		// If filtering for executables, only stat if not directory
		if onlyExec && !file.IsDir() {
			info, err := file.Info()
			if err != nil {
				continue
			}

			mode := info.Mode().Perm()
			if mode&0111 == 0 {
				continue
			}
		}

		if useFullPath {
			name = fullPath
		}

		switch {

		case file.Type()&os.ModeSymlink != 0:
			desc = "-> " + resolveSymlink(fullPath)

		case file.Type()&os.ModeNamedPipe != 0:
			desc = "named pipe"

		case file.IsDir():
			desc = "directory"

		default:
			desc = ""
		}

		suggestions = append(suggestions, prompt.Suggest{
			Text:        name,
			Description: desc,
		})
	}

	return suggestions
}

// vim: set ts=4 sw=4 noet:
