// gethomedir.go
// get home directory without using os/user
//
// Copyright (c) 2024-2026 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// getHomeDir looks up a user's home directory from /etc/passwd.
func getHomeDir(username string) (string, error) {
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// Format: username:password:UID:GID:GECOS:home:shell
		parts := strings.SplitN(line, ":", 7)
		if len(parts) < 7 {
			continue
		}

		if parts[0] == username {
			return parts[5], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	// User not found
	return "", nil
}

// expandHome expands ~ and ~user prefixes in paths when possible.
func expandHome(path string) string {
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}

		return path
	}

	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}

		return path
	}

	if strings.HasPrefix(path, "~") {
		remaining := path[1:]

		var userPart, relative string

		if idx := strings.IndexByte(remaining, '/'); idx >= 0 {
			userPart = remaining[:idx]
			relative = remaining[idx+1:]
		} else {
			userPart = remaining
			relative = ""
		}

		homeDir, err := getHomeDir(userPart)
		if err == nil && homeDir != "" {
			if relative != "" {
				return filepath.Join(homeDir, relative)
			}

			return homeDir
		}
	}

	return path
}

// vim: set ts=4 sw=4 noet:
