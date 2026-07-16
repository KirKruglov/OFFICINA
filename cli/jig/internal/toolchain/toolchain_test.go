package toolchain

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// call records one invocation of the fake Runner.
type call struct {
	dir  string
	argv []string
}

// result is a programmed answer for one command.
type result struct {
	out string
	err error
}

// fakeRunner records calls and returns programmed results keyed by the joined argv.
type fakeRunner struct {
	calls   []call
	results map[string]result
}

func (f *fakeRunner) record(dir string, argv []string) result {
	f.calls = append(f.calls, call{dir: dir, argv: argv})
	return f.results[strings.Join(argv, " ")]
}

func (f *fakeRunner) Run(dir string, argv []string) error {
	return f.record(dir, argv).err
}

func (f *fakeRunner) Output(dir string, argv []string) (string, error) {
	r := f.record(dir, argv)
	return r.out, r.err
}

func (f *fakeRunner) argvs() []string {
	var got []string
	for _, c := range f.calls {
		got = append(got, strings.Join(c.argv, " "))
	}
	return got
}

func newFake(results map[string]result) *fakeRunner {
	if results == nil {
		results = map[string]result{}
	}
	return &fakeRunner{results: results}
}

// fakeLooker resolves only the names in found.
func fakeLooker(found ...string) Looker {
	set := map[string]bool{}
	for _, n := range found {
		set[n] = true
	}
	return func(name string) (string, error) {
		if set[name] {
			return "/usr/bin/" + name, nil
		}
		return "", errors.New("not found: " + name)
	}
}

func TestCheckPath(t *testing.T) {
	tests := []struct {
		name  string
		found []string
		names []string
		want  []string
	}{
		{
			name:  "all found",
			found: []string{"git", "go", "uv"},
			names: []string{"git", "go", "uv"},
			want:  nil,
		},
		{
			name:  "one missing",
			found: []string{"git", "go"},
			names: []string{"git", "uv", "go"},
			want:  []string{"uv"},
		},
		{
			name:  "missing in input order",
			found: []string{"go"},
			names: []string{"npm", "go", "uv", "git"},
			want:  []string{"npm", "uv", "git"},
		},
		{
			name:  "all missing",
			found: nil,
			names: []string{"git", "uv", "npm"},
			want:  []string{"git", "uv", "npm"},
		},
		{
			name:  "empty input",
			found: []string{"git"},
			names: nil,
			want:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckPath(fakeLooker(tt.found...), tt.names)
			if len(got) != len(tt.want) {
				t.Fatalf("CheckPath() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("CheckPath() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestGitInit(t *testing.T) {
	f := newFake(nil)
	if err := GitInit(f, "/tmp/demo"); err != nil {
		t.Fatalf("GitInit() error = %v", err)
	}
	if len(f.calls) != 1 {
		t.Fatalf("calls = %v, want exactly 1", f.argvs())
	}
	if got := strings.Join(f.calls[0].argv, " "); got != "git init" {
		t.Errorf("argv = %q, want %q", got, "git init")
	}
	if f.calls[0].dir != "/tmp/demo" {
		t.Errorf("dir = %q, want %q", f.calls[0].dir, "/tmp/demo")
	}
}

func TestGitInitError(t *testing.T) {
	f := newFake(map[string]result{
		"git init": {err: errors.New("boom")},
	})
	err := GitInit(f, "/tmp/demo")
	if err == nil {
		t.Fatal("GitInit() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "git init") {
		t.Errorf("error = %q, want it to name the failed command", err)
	}
}

func TestGitIdentityOK(t *testing.T) {
	f := newFake(map[string]result{
		"git config user.name":  {out: "Test User"},
		"git config user.email": {out: "dev@example.com"},
	})
	if err := GitIdentity(f, "/tmp/demo"); err != nil {
		t.Fatalf("GitIdentity() error = %v", err)
	}
	if len(f.calls) != 2 {
		t.Fatalf("calls = %v, want 2", f.argvs())
	}
	for _, c := range f.calls {
		if c.dir != "/tmp/demo" {
			t.Errorf("dir = %q, want %q", c.dir, "/tmp/demo")
		}
	}
}

func TestGitIdentityErrors(t *testing.T) {
	tests := []struct {
		name    string
		results map[string]result
		want    string
	}{
		{
			name: "name empty",
			results: map[string]result{
				"git config user.name":  {out: ""},
				"git config user.email": {out: "dev@example.com"},
			},
			want: "user.name",
		},
		{
			name: "email empty",
			results: map[string]result{
				"git config user.name":  {out: "Test User"},
				"git config user.email": {out: ""},
			},
			want: "user.email",
		},
		{
			name: "name call errors",
			results: map[string]result{
				"git config user.name":  {err: errors.New("exit status 1")},
				"git config user.email": {out: "dev@example.com"},
			},
			want: "user.name",
		},
		{
			name: "email call errors",
			results: map[string]result{
				"git config user.name":  {out: "Test User"},
				"git config user.email": {err: errors.New("exit status 1")},
			},
			want: "user.email",
		},
		{
			name:    "both unset",
			results: map[string]result{},
			want:    "user.name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFake(tt.results)
			err := GitIdentity(f, "/tmp/demo")
			if err == nil {
				t.Fatal("GitIdentity() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %q, want it to mention %q", err, tt.want)
			}
		})
	}
}

func TestGitAuthor(t *testing.T) {
	tests := []struct {
		name    string
		results map[string]result
		want    string
	}{
		{
			name:    "returns the name",
			results: map[string]result{"git config user.name": {out: "Test User"}},
			want:    "Test User",
		},
		{
			name:    "empty output is not an error",
			results: map[string]result{"git config user.name": {out: ""}},
			want:    "",
		},
		{
			name:    "erroring call is not an error",
			results: map[string]result{"git config user.name": {err: errors.New("exit status 1")}},
			want:    "",
		},
		{
			name:    "erroring call with output is not an error",
			results: map[string]result{"git config user.name": {out: "junk", err: errors.New("exit status 1")}},
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFake(tt.results)
			got, err := GitAuthor(f, "/tmp/demo")
			if err != nil {
				t.Fatalf("GitAuthor() error = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("GitAuthor() = %q, want %q", got, tt.want)
			}
			if len(f.calls) != 1 || strings.Join(f.calls[0].argv, " ") != "git config user.name" {
				t.Errorf("calls = %v, want exactly [git config user.name]", f.argvs())
			}
		})
	}
}

func TestGitCommit(t *testing.T) {
	f := newFake(nil)
	const msg = "chore: initial scaffold from jig"
	if err := GitCommit(f, "/tmp/demo", msg); err != nil {
		t.Fatalf("GitCommit() error = %v", err)
	}
	want := []string{"git add -A", "git commit -m " + msg}
	got := f.argvs()
	if len(got) != len(want) {
		t.Fatalf("calls = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("calls = %v, want %v", got, want)
		}
	}
	// argv must carry the message as a single element, not shell-quoted.
	commit := f.calls[1].argv
	if len(commit) != 4 || commit[3] != msg {
		t.Errorf("commit argv = %#v, want message as one element", commit)
	}
	for _, c := range f.calls {
		if c.dir != "/tmp/demo" {
			t.Errorf("dir = %q, want %q", c.dir, "/tmp/demo")
		}
	}
}

func TestGitCommitStopsWhenAddFails(t *testing.T) {
	f := newFake(map[string]result{
		"git add -A": {err: errors.New("boom")},
	})
	err := GitCommit(f, "/tmp/demo", "chore: initial scaffold from jig")
	if err == nil {
		t.Fatal("GitCommit() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "git add") {
		t.Errorf("error = %q, want it to name the failed command", err)
	}
	got := f.argvs()
	if len(got) != 1 || got[0] != "git add -A" {
		t.Fatalf("calls = %v, want only [git add -A]", got)
	}
}

func TestGitCommitPropagatesCommitError(t *testing.T) {
	f := newFake(map[string]result{
		"git commit -m msg": {err: errors.New("nothing to commit")},
	})
	err := GitCommit(f, "/tmp/demo", "msg")
	if err == nil {
		t.Fatal("GitCommit() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "git commit") {
		t.Errorf("error = %q, want it to name the failed command", err)
	}
}

func TestExecRunnerEmptyArgv(t *testing.T) {
	var e ExecRunner
	if err := e.Run(t.TempDir(), nil); err == nil {
		t.Error("Run() with empty argv: error = nil, want error")
	}
	if _, err := e.Output(t.TempDir(), []string{}); err == nil {
		t.Error("Output() with empty argv: error = nil, want error")
	}
}

func TestExecRunnerOutputTrimsNewline(t *testing.T) {
	if _, err := exec.LookPath("echo"); err != nil {
		t.Skip("echo not in PATH")
	}
	var e ExecRunner
	got, err := e.Output(t.TempDir(), []string{"echo", "hello"})
	if err != nil {
		t.Fatalf("Output() error = %v", err)
	}
	if got != "hello" {
		t.Errorf("Output() = %q, want %q", got, "hello")
	}
}

func TestExecRunnerRunStreams(t *testing.T) {
	if _, err := exec.LookPath("echo"); err != nil {
		t.Skip("echo not in PATH")
	}
	var out strings.Builder
	e := ExecRunner{Stdout: &out, Stderr: &out}
	if err := e.Run(t.TempDir(), []string{"echo", "streamed"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(out.String(), "streamed") {
		t.Errorf("stdout = %q, want it to contain %q", out.String(), "streamed")
	}
}

func TestExecRunnerRunFailureNamesCommand(t *testing.T) {
	var e ExecRunner
	err := e.Run(t.TempDir(), []string{"jig-no-such-command-xyz"})
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "jig-no-such-command-xyz") {
		t.Errorf("error = %q, want it to name the command", err)
	}
}
