// detectComp.go
// Detects the compression format of a file from its magic bytes and looks up
// the external command needed to decompress it to stdout.
//
// Copyright (c) 2024-2026 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// detectComp identifies a file's compression format by magic bytes.
// Returns ("", nil) when the file is not a recognised compressed format.
func detectComp(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return "", err
	}
	if !fi.Mode().IsRegular() {
		return "", nil
	}

	buf := make([]byte, 8)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	buf = buf[:n]

	switch {

	case len(buf) >= 2 &&
		buf[0] == 0x1f &&
		buf[1] == 0x8b:
		return "gzip", nil

	case len(buf) >= 3 &&
		bytes.Equal(buf[:3], []byte("BZh")):
		return "bzip2", nil

	case len(buf) >= 6 &&
		bytes.Equal(buf[:6], []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00}):
		return "xz", nil

	case len(buf) >= 4 &&
		bytes.Equal(buf[:4], []byte{0x28, 0xb5, 0x2f, 0xfd}):
		return "zstd", nil

	case len(buf) >= 4 &&
		bytes.Equal(buf[:4], []byte("PK\x03\x04")):
		return "zip", nil

	case len(buf) >= 4 &&
		(bytes.Equal(buf[:4], []byte{0x04, 0x22, 0x4d, 0x18}) ||
			bytes.Equal(buf[:4], []byte{0x02, 0x21, 0x4c, 0x18})):
		return "lz4", nil

	case len(buf) >= 6 &&
		bytes.Equal(buf[:6], []byte{0x37, 0x7a, 0xbc, 0xaf, 0x27, 0x1c}):
		return "7z", nil

	case len(buf) >= 2 &&
		buf[0] == 0x1f &&
		buf[1] == 0x9d:
		return "compress", nil
	}

	return "", nil
}

// decompressCommands maps a format name (as returned by detectComp)
// to the program and flags that decompress that format to stdout.
var decompressCommands = map[string][]string{
	"gzip":     {"gzip", "-dc"},
	"bzip2":    {"bzip2", "-dc"},
	"xz":       {"xz", "-dc"},
	"zstd":     {"zstd", "-dc"},
	"zip":      {"funzip"},
	"lz4":      {"lz4", "-dc"},
	"7z":       {"7z", "x", "-so"},
	"compress": {"uncompress", "-c"},
}

// decompressCommand returns the program name and full argument list (with
// filename appended) needed to decompress filename to stdout, suitable for
// passing to exec.Command(prog, args...).
func decompressCommand(format, filename string) (prog string, args []string, err error) {
	cmd, ok := decompressCommands[format]
	if !ok {
		return "", nil, fmt.Errorf("no decompressor known for format %q", format)
	}

	args = append(cmd[1:], filename)
	return cmd[0], args, nil
}

// vim: set ts=4 sw=4 noet:
