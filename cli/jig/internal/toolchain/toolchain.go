// Package toolchain checks PATH and runs the external commands jig depends on.
//
// Commands are always argv slices — never shell strings — so there is nothing
// to quote and nothing to inject.
package toolchain

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// errEmptyArgv guards against a manifest or caller handing over an empty command.
var errEmptyArgv = errors.New("toolchain: empty command")

// Runner runs external commands. Swapped for a fake in tests.
type Runner interface {
	// Run executes argv in dir, streaming its output to the user.
	Run(dir string, argv []string) error
	// Output executes argv in dir and returns its stdout.
	Output(dir string, argv []string) (string, error)
}

// ExecRunner is the os/exec implementation of Runner.
// Run streams the command output to Stdout and Stderr; a nil writer discards.
type ExecRunner struct {
	Stdout, Stderr io.Writer
}

// Run executes argv in dir, streaming stdout and stderr as the command writes them.
func (e ExecRunner) Run(dir string, argv []string) error {
	if len(argv) == 0 {
		return errEmptyArgv
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Stdout = e.Stdout
	cmd.Stderr = e.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", strings.Join(argv, " "), err)
	}
	return nil
}

// Output executes argv in dir and returns its stdout, trimmed of trailing newlines.
// On failure the error carries the command and whatever it wrote to stderr.
func (e ExecRunner) Output(dir string, argv []string) (string, error) {
	if len(argv) == 0 {
		return "", errEmptyArgv
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := strings.TrimRight(stdout.String(), "\r\n")
	if err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return out, fmt.Errorf("%s: %w: %s", strings.Join(argv, " "), err, msg)
		}
		return out, fmt.Errorf("%s: %w", strings.Join(argv, " "), err)
	}
	return out, nil
}

// Looker resolves a command name on PATH. Production value: exec.LookPath.
type Looker func(name string) (string, error)

// CheckPath returns the names not found by look, in input order.
// An empty result means every name resolved.
func CheckPath(look Looker, names []string) []string {
	var missing []string
	for _, name := range names {
		if _, err := look(name); err != nil {
			missing = append(missing, name)
		}
	}
	return missing
}

// GitInit runs `git init` in dir.
func GitInit(r Runner, dir string) error {
	if err := r.Run(dir, []string{"git", "init"}); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}
	return nil
}

// GitIdentity errors if user.name or user.email is unset or empty:
// without both, `git commit` fails after the initializers have already run.
func GitIdentity(r Runner, dir string) error {
	for _, key := range []string{"user.name", "user.email"} {
		// git config exits non-zero when a key is unset — an error and an
		// empty value mean the same thing here.
		val, err := r.Output(dir, []string{"git", "config", key})
		if err != nil || strings.TrimSpace(val) == "" {
			return fmt.Errorf("git %s is not set: run `git config --global %s ...`", key, key)
		}
	}
	return nil
}

// GitAuthor returns `git config user.name` for template substitution.
// An empty value is not an error: it only matters when a commit is requested,
// and that case is covered by GitIdentity.
func GitAuthor(r Runner, dir string) (string, error) {
	out, err := r.Output(dir, []string{"git", "config", "user.name"})
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(out), nil
}

// GitCommit stages everything in dir and commits it with msg.
func GitCommit(r Runner, dir, msg string) error {
	if err := r.Run(dir, []string{"git", "add", "-A"}); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}
	if err := r.Run(dir, []string{"git", "commit", "-m", msg}); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}
	return nil
}
