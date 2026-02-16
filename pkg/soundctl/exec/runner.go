package exec

import (
	"bytes"
	"context"
	"fmt"
	osexec "os/exec"
	"strings"
	"sync"
)

// Runner abstracts shell command execution for testability.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

// OSRunner executes commands on the host system.
type OSRunner struct{}

func NewOSRunner() *OSRunner {
	return &OSRunner{}
}

func (r *OSRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := osexec.CommandContext(ctx, name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	out := strings.TrimSpace(stdout.String())
	if err != nil {
		errText := strings.TrimSpace(stderr.String())
		if errText != "" {
			return out, fmt.Errorf("%w: %s", err, errText)
		}
		// Some CLIs (including bluetoothctl) print failures to stdout.
		if out != "" {
			return out, fmt.Errorf("%w: %s", err, out)
		}
		return out, err
	}
	return out, nil
}

// CommandResult defines the fake result for a command key.
type CommandResult struct {
	Output string
	Err    error
}

// FakeRunner is a deterministic test runner.
type FakeRunner struct {
	mu        sync.Mutex
	responses map[string]CommandResult
	calls     []string
}

func NewFakeRunner() *FakeRunner {
	return &FakeRunner{responses: map[string]CommandResult{}}
}

func CommandKey(name string, args ...string) string {
	parts := append([]string{name}, args...)
	return strings.Join(parts, " ")
}

func (f *FakeRunner) Set(name string, args []string, result CommandResult) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.responses[CommandKey(name, args...)] = result
}

func (f *FakeRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := CommandKey(name, args...)
	f.calls = append(f.calls, key)
	res, ok := f.responses[key]
	if !ok {
		return "", fmt.Errorf("no fake response for command: %s", key)
	}
	return res.Output, res.Err
}

func (f *FakeRunner) Calls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.calls))
	copy(out, f.calls)
	return out
}
