// gethomedir.go
// get home directory without using os/user
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"bufio"
	"os"
	"strings"
)

func GetHomeDir(username string) (string, error) {
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

	return "", err
}

// vim: set ts=4 sw=4 noet:
