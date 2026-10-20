package cli

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/Gulliby-development/selwhy/internal/base"
)

type dispatchTestIO struct {
	stdin    []byte
	reads    int
	stdout   string
	outCalls int
	stderr   string
	errCalls int
}

func (io *dispatchTestIO) ReadStdin() []byte {
	io.reads++
	return io.stdin
}

func (io *dispatchTestIO) Out(text string) {
	io.outCalls++
	io.stdout += text
}

func (io *dispatchTestIO) Err(text string) {
	io.errCalls++
	io.stderr += text
}

var _ base.IO = (*dispatchTestIO)(nil)

func TestRunDispatch(t *testing.T) {
	tests := []struct {
		name     string
		argv     []string
		code     int
		stdout   string
		stderr   string
		outCalls int
		errCalls int
	}{
		{
			name:     "missing command",
			argv:     []string{},
			code:     2,
			stderr:   "selwhy: missing command (tree, select, explain, specificity, fmt, nth or cascade)\n",
			errCalls: 1,
		},
		{
			name:     "unknown command",
			argv:     []string{"tre"},
			code:     2,
			stderr:   "selwhy: unknown command \"tre\"\n",
			errCalls: 1,
		},
		{
			name:     "help is only recognized alone",
			argv:     []string{"--help", "tree"},
			code:     2,
			stderr:   "selwhy: unknown command \"--help\"\n",
			errCalls: 1,
		},
		{
			name:     "help",
			argv:     []string{"--help"},
			code:     0,
			stdout:   "selwhy tree DOC\nselwhy select SELECTOR DOC\nselwhy explain SELECTOR DOC N\nselwhy specificity SELECTOR\nselwhy fmt SELECTOR\nselwhy nth [--count N] FORMULA\nselwhy cascade SHEET DOC\n",
			outCalls: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			io := &dispatchTestIO{}
			if got := Run(test.argv, io); got != test.code {
				t.Fatalf("Run(%q) = %d, want %d", test.argv, got, test.code)
			}
			if io.stdout != test.stdout {
				t.Fatalf("Run(%q) stdout = %q, want %q", test.argv, io.stdout, test.stdout)
			}
			if io.stderr != test.stderr {
				t.Fatalf("Run(%q) stderr = %q, want %q", test.argv, io.stderr, test.stderr)
			}
			if io.reads != 0 {
				t.Fatalf("Run(%q) made %d stdin reads, want none", test.argv, io.reads)
			}
			if io.outCalls != test.outCalls {
				t.Fatalf("Run(%q) made %d Out calls, want %d", test.argv, io.outCalls, test.outCalls)
			}
			if io.errCalls != test.errCalls {
				t.Fatalf("Run(%q) made %d Err calls, want %d", test.argv, io.errCalls, test.errCalls)
			}
		})
	}
}

func TestMalformedUTF8OnStandardInput(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "selwhy")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	buildCtx, cancelBuild := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancelBuild()
	build := exec.CommandContext(buildCtx, "go", "build", "-o", bin, "../..")
	buildOutput, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, buildOutput)
	}

	runCtx, cancelRun := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancelRun()
	program := exec.CommandContext(runCtx, bin, "tree", "-")
	program.Stdin = bytes.NewReader([]byte{0x65, 0x63, 0xff, 0x0a})
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	program.Stdout = &stdout
	program.Stderr = &stderr
	err = program.Run()
	if err == nil {
		t.Fatal("tree process succeeded, want exit code 2")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("tree process error = %v, want exit code 2", err)
	}
	if got := exitErr.ExitCode(); got != 2 {
		t.Fatalf("tree process exit code = %d, want 2", got)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("tree process stdout = %q, want empty", got)
	}
	if got := stderr.String(); got != "selwhy: <stdin>: input is not valid UTF-8\n" {
		t.Fatalf("tree process stderr = %q, want %q", got, "selwhy: <stdin>: input is not valid UTF-8\n")
	}
}
