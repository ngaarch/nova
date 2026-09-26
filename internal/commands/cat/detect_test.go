package cat

import (
	"strings"
	"testing"
)

func TestIsBinary(t *testing.T) {
	// Plain text
	text := []byte("Hello, world! This is a standard ASCII text string.\nWith multiple lines.\n")
	if IsBinary(text) {
		t.Errorf("expected text to not be detected as binary")
	}

	// Null byte
	nullBytes := []byte("Some text\x00with null")
	if !IsBinary(nullBytes) {
		t.Errorf("expected null bytes to be detected as binary")
	}

	// High control code ratio
	controlBytes := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 'a', 'b'}
	if !IsBinary(controlBytes) {
		t.Errorf("expected control bytes to be detected as binary")
	}

	// Empty
	if IsBinary([]byte{}) {
		t.Errorf("expected empty bytes to not be binary")
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		filename string
		header   []byte
		expected string
	}{
		{"main.go", nil, "go"},
		{"script.py", nil, "python"},
		{"app.ts", nil, "typescript"},
		{"index.js", nil, "javascript"},
		{"data.json", nil, "json"},
		{"config.yaml", nil, "yaml"},
		{"config.yml", nil, "yaml"},
		{"Cargo.toml", nil, "toml"},
		{"README.md", nil, "markdown"},
		{"deploy.sh", nil, "bash"},
		{"query.sql", nil, "sql"},
		{"Dockerfile", nil, "dockerfile"},
		{"Makefile", nil, "makefile"},
		{"unknown", []byte("#!/usr/bin/env python3\nprint('hi')"), "python"},
		{"script", []byte("#!/bin/bash\necho ok"), "bash"},
		{"plain.txt", []byte("just text"), "text"},
	}

	for _, tt := range tests {
		got := DetectLanguage(tt.filename, tt.header)
		if got != tt.expected {
			t.Errorf("DetectLanguage(%q) = %q; want %q", tt.filename, got, tt.expected)
		}
	}
}

func TestFormatHexDump(t *testing.T) {
	data := []byte("Hello, World!\x00\x01\x02")
	rows := FormatHexDump(data, 0)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if !strings.HasPrefix(rows[0], "00000000  48 65 6c 6c 6f 2c 20 57  6f 72 6c 64 21 00 01 02") {
		t.Errorf("unexpected hex row: %s", rows[0])
	}
	if !strings.HasSuffix(rows[0], "|Hello, World!...|") {
		t.Errorf("unexpected ascii representation: %s", rows[0])
	}
}
