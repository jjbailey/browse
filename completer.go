// completer.go
// central completion module (bash + file)
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

	"github.com/c-bata/go-prompt"
)

const (
	dispSuggestions = 5
	maxSuggestions  = 1000
	searchFiles     = 1
	searchPath      = 2
)

var searchType int

func userBashComp() (string, bool) {
	searchType = searchPath
	input, flag := runCompleter("$ ", commHistory)

	// Get hostname title
	title := getHostnameTitle()
	fmt.Printf(XTERMTITLE, title)

	ttyBrowser()
	return input, flag
}

func userFileComp() (string, bool) {
	searchType = searchFiles
	input, flag := runCompleter("File: ", fileHistory)

	// Get hostname title
	title := getHostnameTitle()
	fmt.Printf(XTERMTITLE, title)

	ttyBrowser()
	return input, flag
}

func runCompleter(promptStr, historyFile string) (string, bool) {
	// Load history from file
	history := loadHistory(historyFile)

	// Get hostname title
	title := getHostnameTitle()

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
	if len(input) == 0 {
		return "", true
	}
	return input, false
}

func getHostnameTitle() string {
	// Can't prevent prompt.New from meddling with the title, so...

	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}

	if dotIndex := strings.Index(hostname, "."); dotIndex != -1 {
		hostname = hostname[:dotIndex]
	}

	return hostname
}

func completer(d prompt.Document) []prompt.Suggest {
	var suggestions []prompt.Suggest

	// Get the word being completed
	word := strings.ReplaceAll(d.GetWordBeforeCursor(), "//", "/")

	// Handle home directory expansion
	if strings.HasPrefix(word, "~") {
		if len(word) == 1 || word[1] == '/' {
			// ~ or ~/something
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return suggestions
			}
			word = filepath.Join(homeDir, word[1:])
		} else {
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

			wordHome, err := GetHomeDir(userName)
			if err != nil {
				return suggestions
			}
			word = filepath.Join(wordHome, pathSuffix)
		}
	}

	// Handle absolute paths and relative paths
	if strings.HasPrefix(word, "/") ||
		strings.HasPrefix(word, ".") ||
		strings.HasPrefix(word, "..") ||
		strings.HasPrefix(word, "~") ||
		strings.Contains(word, "/") {

		// Get the directory part of the path
		dir := filepath.Dir(word)

		files, err := os.ReadDir(dir)
		if err != nil {
			return suggestions
		}

		if strings.HasSuffix(word, "/") {
			word = ""
		} else {
			word = filepath.Base(word)
		}

		appendSuggestions(&suggestions, files, dir, word, true)
		return suggestions
	}

	var pathDirs []string

	if searchType == searchPath {
		// Handle $PATH
		pathDirs = strings.Split(os.Getenv("PATH"), ":")
	} else {
		// Handle current directory
		pathDirs = []string{"."}
	}

	for _, dir := range pathDirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		appendSuggestions(&suggestions, files, dir, word, false)
	}

	return suggestions
}

func appendSuggestions(suggestions *[]prompt.Suggest, files []os.DirEntry, dir string, prefix string, useFullPath bool) {
	for _, file := range files {
		if len(*suggestions) >= maxSuggestions {
			break
		}

		if !strings.HasPrefix(file.Name(), prefix) {
			continue
		}

		var suggest prompt.Suggest
		text := file.Name()
		path := filepath.Join(dir, text)
		if useFullPath {
			text = path
		}

		switch {

		case file.Type()&os.ModeSymlink != 0:
			realPath := resolveSymlink(path)
			suggest = prompt.Suggest{
				Text:        text,
				Description: "-> " + realPath,
			}

		case file.Type()&os.ModeNamedPipe != 0:
			suggest = prompt.Suggest{
				Text:        text,
				Description: "named pipe",
			}

		case file.IsDir():
			suggest = prompt.Suggest{
				Text:        text,
				Description: "directory",
			}

		default:
			suggest = prompt.Suggest{
				Text: text,
			}
		}

		*suggestions = append(*suggestions, suggest)
	}
}

// vim: set ts=4 sw=4 noet:
