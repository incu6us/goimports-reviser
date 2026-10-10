package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/incu6us/goimports-reviser/v3/reviser"
)

func TestCLIFormatsStdinToStdout(t *testing.T) {
	binaryName := "goimports-reviser"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(t.TempDir(), binaryName)
	build := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build goimports-reviser: %v\n%s", err, output)
	}

	const input = `package main
import (
"github.com/pkg/errors"
"fmt"
)
func main() {
fmt.Println(errors.New("boom"))
}
`
	const want = `package main

import (
	"fmt"

	"github.com/pkg/errors"
)

func main() {
	fmt.Println(errors.New("boom"))
}
`

	cmd := exec.Command(binaryPath, "-format")
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("goimports-reviser failed: %v\nstderr:\n%s", err, stderr.String())
	}
	if got := stdout.String(); got != want {
		t.Fatalf("unexpected stdout:\n%s\nwant:\n%s", got, want)
	}
}

func TestNormalizeOriginPaths(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		want  []string
	}{
		{
			name: "no paths reads standard input",
			want: []string{reviser.StandardInput},
		},
		{
			name:  "dash reads standard input",
			paths: []string{"-"},
			want:  []string{reviser.StandardInput},
		},
		{
			name:  "file path is preserved",
			paths: []string{"main.go"},
			want:  []string{"main.go"},
		},
		{
			name:  "multiple paths are preserved",
			paths: []string{"main.go", "reviser/file.go"},
			want:  []string{"main.go", "reviser/file.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeOriginPaths(tt.paths)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("normalizeOriginPaths(%v) = %v, want %v", tt.paths, got, tt.want)
			}
		})
	}
}

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
