package main

import (
	"os"
	"runtime"
	"testing"
)

func TestIsTerminal(t *testing.T) {
	t.Run("pipe is not a terminal", func(t *testing.T) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("failed to create pipe: %v", err)
		}
		defer r.Close()
		defer w.Close()

		if isTerminal(r) {
			t.Error("expected pipe to not be a terminal")
		}
	})

	t.Run("regular file is not a terminal", func(t *testing.T) {
		f, err := os.CreateTemp("", "terminal-test")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(f.Name())
		defer f.Close()

		if isTerminal(f) {
			t.Error("expected regular file to not be a terminal")
		}
	})

	t.Run("dev/tty is a terminal", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping /dev/tty test on Windows")
		}

		f, err := os.Open("/dev/tty")
		if err != nil {
			t.Skip("skipping: /dev/tty not available (CI environment)")
		}
		defer f.Close()

		if !isTerminal(f) {
			t.Error("expected /dev/tty to be a terminal")
		}
	})

	t.Run("closed file returns false", func(t *testing.T) {
		f, err := os.CreateTemp("", "terminal-test-closed")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		os.Remove(f.Name())
		f.Close()

		if isTerminal(f) {
			t.Error("expected closed file to not be a terminal")
		}
	})
}
