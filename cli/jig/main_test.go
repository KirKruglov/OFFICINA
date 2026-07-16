package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"jig/internal/interview"
	"jig/internal/manifest"
	"jig/internal/render"
)

// --- fakes ---------------------------------------------------------------

// call is one recorded invocation of the fake runner.
type call struct {
	dir  string
	argv []string
}

func (c call) line() string { return strings.Join(c.argv, " ") }

// fakeRunner records every command instead of running it. No unit test ever
// invokes a real git, uv or npm.
type fakeRunner struct {
	calls  []call
	output map[string]string // argv joined → stdout
	err    map[string]error  // argv joined → error
	// onRun, when set, fires at the moment a command is invoked. It is how a
	// test observes the state of the disk mid-run: "was the file already
	// there when go mod init ran?"
	onRun func(dir string, argv []string)
}

func newRunner() *fakeRunner {
	return &fakeRunner{
		output: map[string]string{
			"git config user.name":  "Test User",
			"git config user.email": "dev@example.com",
		},
		err: map[string]error{},
	}
}

func (f *fakeRunner) Run(dir string, argv []string) error {
	f.calls = append(f.calls, call{dir, argv})
	if f.onRun != nil {
		f.onRun(dir, argv)
	}
	return f.err[strings.Join(argv, " ")]
}

func (f *fakeRunner) Output(dir string, argv []string) (string, error) {
	f.calls = append(f.calls, call{dir, argv})
	key := strings.Join(argv, " ")
	if err, ok := f.err[key]; ok {
		return "", err
	}
	return f.output[key], nil
}

// lines returns every recorded command as "argv joined".
func (f *fakeRunner) lines() []string {
	out := make([]string, 0, len(f.calls))
	for _, c := range f.calls {
		out = append(out, c.line())
	}
	return out
}

func (f *fakeRunner) ran(line string) bool { return slices.Contains(f.lines(), line) }

// testFS mirrors the shape of the real embedded tree without depending on its
// content. Only TestEmbeddedManifestsParse touches the real embed.
func testFS() fstest.MapFS {
	return fstest.MapFS{
		"templates/core/CLAUDE.md.tmpl": &fstest.MapFile{
			Data: []byte("# {{.Name}}\n{{.Description}}\n{{.Date}} {{.Author}}\n"),
		},
		"templates/core/gitignore": &fstest.MapFile{Data: []byte("bin/\n")},
		"templates/core/docs/CLAUDE.md": &fstest.MapFile{
			Data: []byte("docs\n"),
		},
		"templates/core/docs/marketing/CLAUDE.md": &fstest.MapFile{
			Data: []byte("marketing\n"),
		},
		"templates/full-only.json": &fstest.MapFile{
			Data: []byte(`{ "full_only": ["docs/marketing"] }` + "\n"),
		},
		"templates/langs/cli-go/main.go.tmpl": &fstest.MapFile{
			Data: []byte("package main\n\nfunc main() { println(\"hello from {{.Name}}\") }\n"),
		},
		"templates/langs/go/gitignore": &fstest.MapFile{Data: []byte("# Go\n/bin/\n")},
		"templates/typedocs/design/CLAUDE.md": &fstest.MapFile{
			Data: []byte("design docs\n"),
		},
		"templates/typedocs/design/design-guide/gitkeep": &fstest.MapFile{Data: []byte("")},
		"templates/langs/library-go/lib.go.tmpl": &fstest.MapFile{
			Data: []byte("package {{.PackageName}}\n\nfunc Hello() string { return \"hello from {{.Name}}\" }\n"),
		},
		"templates/types/cli.json": &fstest.MapFile{Data: []byte(`{
			"name": { "ru": "Инструмент командной строки", "en": "Command-line tool" },
			"order": 1,
			"slots": [{
				"dir": ".",
				"default_lang": "go",
				"langs": {
					"go": ["go", "mod", "init", "{{.Name}}"],
					"python": ["uv", "init"]
				},
				"files": { "go": ["cli-go", "go"] }
			}]
		}`)},
		"templates/types/library.json": &fstest.MapFile{Data: []byte(`{
			"name": { "ru": "Библиотека", "en": "Library" },
			"order": 3,
			"slots": [{
				"dir": ".",
				"default_lang": "go",
				"langs": {
					"go": ["go", "mod", "init", "{{.Name}}"],
					"python": ["uv", "init"]
				},
				"files": { "go": ["library-go"] }
			}]
		}`)},
		"templates/types/web.json": &fstest.MapFile{Data: []byte(`{
			"name": { "ru": "Веб-приложение", "en": "Web application" },
			"order": 2,
			"docs": "design",
			"slots": [
				{
					"dir": "backend",
					"default_lang": "python",
					"langs": {
						"python": ["uv", "init"],
						"go": ["go", "mod", "init", "{{.Name}}"]
					}
				},
				{
					"dir": "frontend",
					"default_lang": "ts",
					"langs": {
						"ts": ["npm", "create", "vite@latest", ".", "--", "--template", "vanilla-ts"],
						"js": ["npm", "create", "vite@latest", ".", "--", "--template", "vanilla"]
					}
				}
			]
		}`)},
	}
}

// harness is one prepared run: deps plus the buffers to assert against.
type harness struct {
	d      deps
	out    bytes.Buffer
	errOut bytes.Buffer
	runner *fakeRunner
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{runner: newRunner()}
	h.d = deps{
		stdin:       strings.NewReader(""),
		stdout:      &h.out,
		stderr:      &h.errOut,
		env:         func(string) string { return "" }, // → English
		runner:      h.runner,
		look:        func(name string) (string, error) { return "/usr/bin/" + name, nil },
		templates:   testFS(),
		now:         func() time.Time { return time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC) },
		workdir:     t.TempDir(),
		interactive: false,
	}
	return h
}

func (h *harness) run(argv ...string) int { return run(argv, h.d) }

// --- exit codes ----------------------------------------------------------

func TestExitCodes(t *testing.T) {
	base := []string{"--level", "standard", "--description", "demo", "--no-commit"}

	tests := []struct {
		name  string
		argv  []string
		setup func(t *testing.T, h *harness)
		want  int
	}{
		{
			name: "success",
			argv: append(slices.Clone(base), "my-app"),
			want: 0,
		},
		{
			name: "non-empty dir",
			argv: append(slices.Clone(base), "my-app"),
			setup: func(t *testing.T, h *harness) {
				dir := filepath.Join(h.d.workdir, "my-app")
				mkdir(t, dir)
				write(t, filepath.Join(dir, "keep"), "x")
			},
			want: 1,
		},
		{
			name: "existing empty dir is fine",
			argv: append(slices.Clone(base), "my-app"),
			setup: func(t *testing.T, h *harness) {
				mkdir(t, filepath.Join(h.d.workdir, "my-app"))
			},
			want: 0,
		},
		{
			name: "unknown type",
			argv: []string{"--level", "full", "--description", "d", "--no-commit", "--type", "desktop", "--lang", "go", "my-app"},
			want: 2,
		},
		{
			name: "unknown flag",
			argv: append(slices.Clone(base), "--force", "my-app"),
			want: 2,
		},
		{
			name: "bad name",
			argv: append(slices.Clone(base), "myApp"),
			want: 2,
		},
		{
			name: "path traversal name",
			argv: append(slices.Clone(base), "../evil"),
			want: 2,
		},
		{
			name: "trailing hyphen name",
			argv: append(slices.Clone(base), "my-"),
			want: 2,
		},
		{
			name: "missing tool",
			argv: append(slices.Clone(base), "my-app"),
			setup: func(t *testing.T, h *harness) {
				h.d.look = func(name string) (string, error) { return "", os.ErrNotExist }
			},
			want: 1,
		},
		{
			name: "commit and no-commit together",
			argv: []string{"--level", "standard", "--description", "d", "--commit", "--no-commit", "my-app"},
			want: 2,
		},
		{
			name: "non-interactive without enough data",
			argv: []string{"my-app"},
			want: 2,
		},
		{
			name: "slot without a language",
			argv: []string{"--level", "full", "--description", "d", "--no-commit", "--type", "web", "--lang", "backend=go", "my-app"},
			want: 2,
		},
		{
			name: "unknown language for a slot",
			argv: []string{"--level", "full", "--description", "d", "--no-commit", "--type", "cli", "--lang", "rust", "my-app"},
			want: 2,
		},
		{
			name: "type with standard is rejected",
			argv: []string{"--level", "standard", "--type", "cli", "--description", "d", "--no-commit", "my-app"},
			want: 2,
		},
		{
			name: "unknown lang slot is rejected",
			argv: []string{"--level", "full", "--type", "cli", "--lang", "backend=go", "--description", "d", "--no-commit", "my-app"},
			want: 2,
		},
		{
			name: "go keyword name with go is rejected",
			argv: []string{"--level", "full", "--type", "library", "--lang", "go", "--description", "d", "--no-commit", "type"},
			want: 2,
		},
		{
			name: "explicit dot slot in --lang is rejected",
			argv: []string{"--level", "full", "--type", "cli", "--lang", ".=go", "--description", "d", "--no-commit", "my-app"},
			want: 2,
		},
		{
			name: "lang on standard is rejected",
			argv: []string{"--level", "standard", "--lang", "go", "--description", "d", "--no-commit", "my-app"},
			want: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			if tc.setup != nil {
				tc.setup(t, h)
			}
			if got := h.run(tc.argv...); got != tc.want {
				t.Fatalf("exit = %d, want %d\nstdout: %s\nstderr: %s", got, tc.want, h.out.String(), h.errOut.String())
			}
		})
	}
}

func TestNonEmptyDirWritesNothing(t *testing.T) {
	h := newHarness(t)
	dir := filepath.Join(h.d.workdir, "my-app")
	mkdir(t, dir)
	write(t, filepath.Join(dir, "keep"), "x")

	if got := h.run("--level", "standard", "--description", "d", "--no-commit", "my-app"); got != 1 {
		t.Fatalf("exit = %d, want 1", got)
	}
	if entries := readDir(t, dir); len(entries) != 1 {
		t.Fatalf("directory touched: %v", entries)
	}
	if len(h.runner.calls) != 0 {
		t.Fatalf("commands ran despite refusal: %v", h.runner.lines())
	}
}

// --- language pack write policy ------------------------------------------

func TestPackGitignoreAppendedToRoot(t *testing.T) {
	h := newHarness(t)
	code := h.run("--level", "full", "--type", "cli", "--lang", "go",
		"--description", "d", "--no-commit", "my-app")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", code, h.errOut.String())
	}

	got := read(t, filepath.Join(h.d.workdir, "my-app", ".gitignore"))
	want := "bin/\n\n# Go\n/bin/\n"
	if got != want {
		t.Errorf(".gitignore = %q, want %q", got, want)
	}

	line := "Created: " + filepath.Join("my-app", ".gitignore")
	if n := strings.Count(h.out.String(), line+"\n"); n != 1 {
		t.Errorf(".gitignore reported %d times, want 1\n%s", n, h.out.String())
	}
}

func TestPackDoesNotOverwriteInitializerFile(t *testing.T) {
	h := newHarness(t)
	// The initializer drops its own main.go; the cli-go pack must leave it be.
	h.runner.onRun = func(dir string, argv []string) {
		if argv[0] == "go" {
			write(t, filepath.Join(dir, "main.go"), "from initializer\n")
		}
	}
	code := h.run("--level", "full", "--type", "cli", "--lang", "go",
		"--description", "d", "--no-commit", "my-app")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", code, h.errOut.String())
	}

	if got := read(t, filepath.Join(h.d.workdir, "my-app", "main.go")); got != "from initializer\n" {
		t.Errorf("initializer's main.go overwritten: %q", got)
	}
}

// --- type doc pack ---------------------------------------------------------

func TestTypeDocPackRenderedIntoDocs(t *testing.T) {
	h := newHarness(t)
	// The doc pack must be on disk before the first slot command runs.
	sawDocs := false
	h.runner.onRun = func(dir string, argv []string) {
		if argv[0] == "uv" {
			if _, err := os.Stat(filepath.Join(h.d.workdir, "my-app", "docs", "design", "CLAUDE.md")); err == nil {
				sawDocs = true
			}
		}
	}
	code := h.run("--level", "full", "--type", "web",
		"--lang", "backend=python", "--lang", "frontend=ts",
		"--description", "d", "--no-commit", "my-app")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", code, h.errOut.String())
	}

	if got := read(t, filepath.Join(h.d.workdir, "my-app", "docs", "design", "CLAUDE.md")); got != "design docs\n" {
		t.Errorf("docs/design/CLAUDE.md = %q", got)
	}
	if _, err := os.Stat(filepath.Join(h.d.workdir, "my-app", "docs", "design", "design-guide", ".gitkeep")); err != nil {
		t.Errorf("docs/design/design-guide/.gitkeep: %v", err)
	}
	if !sawDocs {
		t.Error("doc pack was not on disk when the first slot command ran")
	}
	if !strings.Contains(h.out.String(), filepath.Join("my-app", "docs", "design", "CLAUDE.md")) {
		t.Errorf("report lacks the doc pack:\n%s", h.out.String())
	}
}

func TestTypeWithoutDocsShipsNoDocPack(t *testing.T) {
	h := newHarness(t)
	code := h.run("--level", "full", "--type", "cli", "--lang", "go",
		"--description", "d", "--no-commit", "my-app")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", code, h.errOut.String())
	}
	if _, err := os.Stat(filepath.Join(h.d.workdir, "my-app", "docs", "design")); !os.IsNotExist(err) {
		t.Errorf("cli run created docs/design, stat err = %v", err)
	}
}

func TestDryRunListsTypeDocPack(t *testing.T) {
	h := newHarness(t)
	code := h.run("--dry-run", "--level", "full", "--type", "web",
		"--lang", "backend=python", "--lang", "frontend=ts",
		"--description", "d", "--no-commit", "my-app")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", code, h.errOut.String())
	}
	if !strings.Contains(h.out.String(), filepath.Join("my-app", "docs", "design", "CLAUDE.md")) {
		t.Errorf("dry-run plan lacks the doc pack:\n%s", h.out.String())
	}
}

// --- level split (mechanism B) -------------------------------------------

func TestStandardSkipsFullOnlyPaths(t *testing.T) {
	h := newHarness(t)
	if got := h.run("--level", "standard", "--description", "d", "--no-commit", "my-app"); got != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", got, h.errOut.String())
	}
	if _, err := os.Stat(filepath.Join(h.d.workdir, "my-app", "docs", "marketing")); !os.IsNotExist(err) {
		t.Errorf("standard run created a full-only path, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(h.d.workdir, "my-app", "docs", "CLAUDE.md")); err != nil {
		t.Errorf("standard run lacks a core file: %v", err)
	}
	if strings.Contains(h.out.String(), "marketing") {
		t.Errorf("report mentions a skipped path:\n%s", h.out.String())
	}
}

func TestFullRendersFullOnlyPaths(t *testing.T) {
	h := newHarness(t)
	code := h.run("--level", "full", "--type", "cli", "--lang", "go",
		"--description", "d", "--no-commit", "my-app")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", code, h.errOut.String())
	}
	if _, err := os.Stat(filepath.Join(h.d.workdir, "my-app", "docs", "marketing", "CLAUDE.md")); err != nil {
		t.Errorf("full run lacks the full-only path: %v", err)
	}
}

func TestDryRunLevelsDiffer(t *testing.T) {
	std := newHarness(t)
	if got := std.run("--dry-run", "--level", "standard", "--description", "d", "--no-commit", "my-app"); got != 0 {
		t.Fatalf("standard exit = %d\nstderr: %s", got, std.errOut.String())
	}
	full := newHarness(t)
	if got := full.run("--dry-run", "--level", "full", "--type", "cli", "--lang", "go",
		"--description", "d", "--no-commit", "my-app"); got != 0 {
		t.Fatalf("full exit = %d\nstderr: %s", got, full.errOut.String())
	}

	marketing := filepath.Join("docs", "marketing", "CLAUDE.md")
	if strings.Contains(std.out.String(), marketing) {
		t.Errorf("standard dry-run lists a full-only path:\n%s", std.out.String())
	}
	if !strings.Contains(full.out.String(), marketing) {
		t.Errorf("full dry-run lacks the full-only path:\n%s", full.out.String())
	}
}

// --- help ----------------------------------------------------------------

func TestHelp(t *testing.T) {
	for _, flagName := range []string{"-h", "--help"} {
		t.Run(flagName, func(t *testing.T) {
			h := newHarness(t)
			if got := h.run(flagName); got != 0 {
				t.Fatalf("exit = %d, want 0", got)
			}
			out := h.out.String()
			for _, want := range []string{"Usage: jig [flags] <name>", "--level standard|full", "--dry-run", "--version"} {
				if !strings.Contains(out, want) {
					t.Errorf("help lacks %q\ngot:\n%s", want, out)
				}
			}
			if h.errOut.Len() != 0 {
				t.Errorf("help wrote to stderr: %s", h.errOut.String())
			}
			if len(h.runner.calls) != 0 {
				t.Errorf("help ran commands: %v", h.runner.lines())
			}
			if entries := readDir(t, h.d.workdir); len(entries) != 0 {
				t.Errorf("help created files: %v", entries)
			}
		})
	}
}

// --- version -------------------------------------------------------------

func TestVersionFlag(t *testing.T) {
	h := newHarness(t)
	if got := h.run("--version"); got != 0 {
		t.Fatalf("exit = %d, want 0", got)
	}
	// The release version is baked into the source.
	if got := h.out.String(); got != "jig 0.2.0\n" {
		t.Errorf("stdout = %q, want %q", got, "jig 0.2.0\n")
	}
	if h.errOut.Len() != 0 {
		t.Errorf("version wrote to stderr: %s", h.errOut.String())
	}
	if len(h.runner.calls) != 0 {
		t.Errorf("version ran commands: %v", h.runner.lines())
	}
	if entries := readDir(t, h.d.workdir); len(entries) != 0 {
		t.Errorf("version created files: %v", entries)
	}
}

// The version is a question about the binary, not about the run: it answers
// before the rest of the command line is judged.
func TestVersionAnswersBeforeValidation(t *testing.T) {
	h := newHarness(t)
	if got := h.run("--version", "--level", "bogus"); got != 0 {
		t.Fatalf("exit = %d, want 0", got)
	}
	if !strings.Contains(h.out.String(), "jig 0.2.0") {
		t.Errorf("stdout = %q, want the version", h.out.String())
	}
}

// --- language files ------------------------------------------------------

func TestLangFilesRenderedAfterInitializer(t *testing.T) {
	h := newHarness(t)
	main := filepath.Join(h.d.workdir, "my-app", "main.go")

	// go mod init tolerates a stray .go file, uv and npm do not tolerate a
	// non-empty directory at all — so the order is the contract, not a detail.
	existedAtInit := false
	h.runner.onRun = func(_ string, argv []string) {
		if strings.Join(argv, " ") == "go mod init my-app" {
			_, err := os.Stat(main)
			existedAtInit = err == nil
		}
	}

	code := h.run("--level", "full", "--type", "cli", "--lang", "go",
		"--description", "проба", "--no-commit", "my-app")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, h.errOut.String())
	}
	if existedAtInit {
		t.Error("main.go existed when go mod init ran: files must land after the initializer")
	}
	if got := read(t, main); !strings.Contains(got, "hello from my-app") {
		t.Errorf("main.go = %q, want the name substituted", got)
	}
	if !strings.Contains(h.out.String(), "Created: my-app/main.go") {
		t.Errorf("report lacks main.go:\n%s", h.out.String())
	}
}

// The commit is last precisely so that generated files get into it.
func TestLangFilesLandBeforeTheCommit(t *testing.T) {
	h := newHarness(t)
	main := filepath.Join(h.d.workdir, "my-app", "main.go")

	existedAtAdd := false
	h.runner.onRun = func(_ string, argv []string) {
		if strings.Join(argv, " ") == "git add -A" {
			_, err := os.Stat(main)
			existedAtAdd = err == nil
		}
	}

	code := h.run("--level", "full", "--type", "cli", "--lang", "go",
		"--description", "проба", "--commit", "my-app")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, h.errOut.String())
	}
	if !existedAtAdd {
		t.Error("main.go was absent at git add -A: it would miss the first commit")
	}
}

// files is keyed by language: python ships none, and must not get main.go.
func TestNoLangFilesForPython(t *testing.T) {
	h := newHarness(t)
	code := h.run("--level", "full", "--type", "cli", "--lang", "python",
		"--description", "проба", "--no-commit", "my-app")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, h.errOut.String())
	}
	if _, err := os.Stat(filepath.Join(h.d.workdir, "my-app", "main.go")); !os.IsNotExist(err) {
		t.Error("main.go was created for the python slot")
	}
}

func TestDryRunListsLangFiles(t *testing.T) {
	h := newHarness(t)
	code := h.run("--dry-run", "--level", "full", "--type", "cli", "--lang", "go",
		"--description", "проба", "--no-commit", "my-app")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, h.errOut.String())
	}
	if !strings.Contains(h.out.String(), "Created: my-app/main.go") {
		t.Errorf("plan lacks main.go:\n%s", h.out.String())
	}
	if entries := readDir(t, h.d.workdir); len(entries) != 0 {
		t.Errorf("dry run created files: %v", entries)
	}
}

// The manifest names a folder; only the embedded tree can say it is there.
func TestEmbeddedLangPacksExist(t *testing.T) {
	types, err := manifest.List(templatesFS, typesRoot)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	found := 0
	for _, name := range types {
		typ, err := manifest.Load(templatesFS, typesRoot, name)
		if err != nil {
			t.Fatalf("load %s: %v", name, err)
		}
		for _, s := range typ.Slots {
			for lang := range s.Langs {
				for _, pack := range s.FilesFor(lang) {
					entries, err := fs.ReadDir(templatesFS, path.Join(langsRoot, pack))
					if err != nil {
						t.Errorf("%s: slot %s: lang %s: pack %q: %v", name, s.Dir, lang, pack, err)
						continue
					}
					if len(entries) == 0 {
						t.Errorf("%s: pack %q is empty", name, pack)
					}
					found++
				}
			}
		}
	}
	// cli: go → cli-go+go, python → python; library: go → library-go+go,
	// python → python; web backend: go → go, python → python.
	if found != 8 {
		t.Errorf("language packs referenced = %d, want 8", found)
	}
}

func TestHelpFollowsUILang(t *testing.T) {
	h := newHarness(t)
	if got := h.run("--ui-lang", "ru", "-h"); got != 0 {
		t.Fatalf("exit = %d, want 0", got)
	}
	if !strings.Contains(h.out.String(), "Использование: jig") {
		t.Fatalf("help is not Russian:\n%s", h.out.String())
	}
}

// --- git identity --------------------------------------------------------

func TestGitIdentityCheckedOnlyWithCommit(t *testing.T) {
	// GitAuthor reads user.name whether or not a commit is requested, so
	// user.email is the marker that tells the identity check apart.
	const marker = "git config user.email"

	t.Run("commit requested", func(t *testing.T) {
		h := newHarness(t)
		if got := h.run("--level", "standard", "--description", "d", "--commit", "my-app"); got != 0 {
			t.Fatalf("exit = %d, want 0\nstderr: %s", got, h.errOut.String())
		}
		if !h.runner.ran(marker) {
			t.Fatalf("identity not checked: %v", h.runner.lines())
		}
	})

	t.Run("no commit", func(t *testing.T) {
		h := newHarness(t)
		if got := h.run("--level", "standard", "--description", "d", "--no-commit", "my-app"); got != 0 {
			t.Fatalf("exit = %d, want 0\nstderr: %s", got, h.errOut.String())
		}
		if h.runner.ran(marker) {
			t.Fatalf("identity checked without a commit: %v", h.runner.lines())
		}
	})
}

func TestGitIdentityMissingRefusesBeforeWriting(t *testing.T) {
	h := newHarness(t)
	h.runner.output["git config user.email"] = ""

	if got := h.run("--level", "standard", "--description", "d", "--commit", "my-app"); got != 1 {
		t.Fatalf("exit = %d, want 1", got)
	}
	if _, err := os.Stat(filepath.Join(h.d.workdir, "my-app")); !os.IsNotExist(err) {
		t.Fatal("directory created despite a missing git identity")
	}
	if !strings.Contains(h.errOut.String(), "git is not configured") {
		t.Fatalf("stderr: %s", h.errOut.String())
	}
}

// --- dry run -------------------------------------------------------------

func TestDryRunWritesNothing(t *testing.T) {
	h := newHarness(t)
	code := h.run("--dry-run", "--level", "full", "--type", "web",
		"--lang", "backend=python", "--lang", "frontend=ts",
		"--description", "demo", "--commit", "my-app")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", code, h.errOut.String())
	}

	out := h.out.String()
	for _, want := range []string{
		"Dry run. Nothing written.",
		filepath.Join("my-app", "CLAUDE.md"),
		filepath.Join("my-app", ".gitignore"),
		"Ran: git init",
		"Ran: uv init",
		"Ran: npm create vite@latest . -- --template vanilla-ts",
		"Ran: git add -A",
		"Commit: " + commitMsg,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("dry run output lacks %q\ngot:\n%s", want, out)
		}
	}

	if entries := readDir(t, h.d.workdir); len(entries) != 0 {
		t.Errorf("dry run wrote files: %v", entries)
	}
	if len(h.runner.calls) != 0 {
		t.Errorf("dry run ran commands: %v", h.runner.lines())
	}
}

// --- embedded manifests --------------------------------------------------

func TestEmbeddedManifestsParse(t *testing.T) {
	types, err := manifest.List(templatesFS, typesRoot)
	if err != nil {
		t.Fatalf("list types: %v", err)
	}
	want := []string{"cli", "web", "library"}
	if !slices.Equal(types, want) {
		t.Fatalf("types = %v, want %v", types, want)
	}
	for _, name := range types {
		typ, err := manifest.Load(templatesFS, typesRoot, name)
		if err != nil {
			t.Errorf("load %s: %v", name, err)
			continue
		}
		if len(typ.Slots) == 0 {
			t.Errorf("%s: no slots", name)
		}
		// Rendered against render.Data, not a map: Command sets
		// missingkey=error, so this asserts what is scary to break — no
		// manifest argument references a variable Data does not have.
		data := render.Data{Name: "demo", Description: "проба"}
		for _, s := range typ.Slots {
			for lang := range s.Langs {
				if _, err := s.Command(lang, data); err != nil {
					t.Errorf("%s: slot %s: lang %s: %v", name, s.Dir, lang, err)
				}
			}
		}
	}
}

// --- embedded template content ---------------------------------------------

func TestEmbeddedCoreTree(t *testing.T) {
	planned, err := render.Plan(templatesFS, coreRoot, nil)
	if err != nil {
		t.Fatalf("plan core: %v", err)
	}

	// Every path the spec's tree (§3) puts into core, in rendered form. The
	// list pins both the dot-restoring rename over the real embed and the
	// .gitkeep holders without which git drops the empty directories.
	want := []string{
		".claude/settings.json",
		".claude/rules/writing-style.md",
		"CLAUDE.md",
		".gitignore",
		"docs/product/CLAUDE.md",
		"docs/product/description/.gitkeep",
		"docs/product/architecture/ARCHITECTURE.md",
		"docs/product/architecture/adr/adr-template.md",
		"docs/product/strategy/.gitkeep",
		"docs/product/release-0.1.0/prd-release/.gitkeep",
		"docs/product/release-0.1.0/spec/.gitkeep",
		"docs/product/release-0.1.0/use-case/.gitkeep",
		"docs/product/release-0.1.0/mockups/.gitkeep",
		"docs/planning/CLAUDE.md",
		"docs/planning/BACKLOG.md",
		"docs/planning/release-plan/.gitkeep",
		"docs/marketing/CLAUDE.md",
		"docs/analytics/CLAUDE.md",
		"docs/archive/.gitkeep",
	}
	for _, p := range want {
		if !slices.Contains(planned, p) {
			t.Errorf("core lacks %s", p)
		}
	}
	if len(planned) != len(want) {
		t.Errorf("core has %d files, want %d: %v", len(planned), len(want), planned)
	}
	for _, p := range planned {
		if strings.HasSuffix(p, "README.md") {
			t.Errorf("core must not ship a README: %s", p)
		}
	}
}

func TestEmbeddedFullOnlyPathsExistInCore(t *testing.T) {
	fullOnly, err := fullOnlyList(templatesFS)
	if err != nil {
		t.Fatalf("full-only list: %v", err)
	}
	if len(fullOnly) == 0 {
		t.Fatal("full-only list is empty")
	}
	all, err := render.Plan(templatesFS, coreRoot, nil)
	if err != nil {
		t.Fatalf("plan core: %v", err)
	}
	standard, err := render.Plan(templatesFS, coreRoot, fullOnly)
	if err != nil {
		t.Fatalf("plan standard: %v", err)
	}
	// A typo in full-only.json would skip nothing and silently hand the
	// standard level the full tree.
	for _, s := range fullOnly {
		covers := slices.ContainsFunc(all, func(p string) bool {
			return p == s || strings.HasPrefix(p, s+"/")
		})
		if !covers {
			t.Errorf("full-only path %q matches no core file", s)
		}
	}
	if len(standard) >= len(all) {
		t.Errorf("standard plan (%d) is not smaller than full (%d)", len(standard), len(all))
	}
}

func TestEmbeddedSettingsJSONValid(t *testing.T) {
	b, err := fs.ReadFile(templatesFS, coreRoot+"/claude/settings.json")
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	var s struct {
		EffortLevel string `json:"effortLevel"`
		Permissions struct {
			Allow []string `json:"allow"`
			Deny  []string `json:"deny"`
		} `json:"permissions"`
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&s); err != nil {
		t.Fatalf("settings.json: %v (unknown keys such as skillOverrides or defaultMode are a spec violation)", err)
	}
	for _, want := range []string{"Read(.env)", "Read(.env.*)", "Bash(cat .env*)", "Bash(echo * > .env*)"} {
		if !slices.Contains(s.Permissions.Deny, want) {
			t.Errorf("deny lacks %q", want)
		}
	}
	if len(s.Permissions.Allow) == 0 {
		t.Error("allow list is empty")
	}
	// The allow base is language-neutral: no toolchain commands baked in.
	for _, a := range s.Permissions.Allow {
		for _, tool := range []string{"go ", "uv ", "npm ", "pytest", "python"} {
			if strings.Contains(a, tool) {
				t.Errorf("allow entry %q bakes in a toolchain", a)
			}
		}
	}
}

func TestEmbeddedTypedocsDoNotOverlapCore(t *testing.T) {
	core, err := render.Plan(templatesFS, coreRoot, nil)
	if err != nil {
		t.Fatalf("plan core: %v", err)
	}
	types, err := manifest.List(templatesFS, typesRoot)
	if err != nil {
		t.Fatalf("list types: %v", err)
	}
	for _, name := range types {
		typ, err := manifest.Load(templatesFS, typesRoot, name)
		if err != nil {
			t.Fatalf("load %s: %v", name, err)
		}
		if typ.Docs == "" {
			continue
		}
		pack, err := render.Plan(templatesFS, path.Join(typedocsRoot, typ.Docs), nil)
		if err != nil {
			t.Fatalf("%s: plan doc pack %q: %v", name, typ.Docs, err)
		}
		if len(pack) == 0 {
			t.Errorf("%s: doc pack %q is empty", name, typ.Docs)
		}
		for _, f := range pack {
			target := path.Join("docs", typ.Docs, f)
			if slices.Contains(core, target) {
				t.Errorf("%s: doc pack path %s collides with core", name, target)
			}
		}
	}
}

func TestEmbeddedGitignorePacksMatchSpec(t *testing.T) {
	wantGo := "# Go\n/bin/\n*.exe\n*.test\n*.out\n"
	wantPython := "# Python\n__pycache__/\n*.py[cod]\n.venv/\nvenv/\n.pytest_cache/\n*.egg-info/\ndist/\nbuild/\n"

	if b, err := fs.ReadFile(templatesFS, langsRoot+"/go/gitignore"); err != nil {
		t.Errorf("go pack: %v", err)
	} else if string(b) != wantGo {
		t.Errorf("go gitignore = %q, want %q", b, wantGo)
	}
	if b, err := fs.ReadFile(templatesFS, langsRoot+"/python/gitignore"); err != nil {
		t.Errorf("python pack: %v", err)
	} else if string(b) != wantPython {
		t.Errorf("python gitignore = %q, want %q", b, wantPython)
	}
}

func TestEmbeddedRootClaudeMdRenders(t *testing.T) {
	dir := t.TempDir()
	data := render.Data{
		Name:        "demo",
		PackageName: "demo",
		Description: "проба",
		Slots: []render.SlotInfo{
			{Dir: "backend", Lang: "python"},
			{Dir: "frontend", Lang: "ts"},
		},
	}
	if _, err := render.Render(templatesFS, coreRoot, dir, data, nil); err != nil {
		t.Fatalf("render core: %v", err)
	}
	got := read(t, filepath.Join(dir, "CLAUDE.md"))
	for _, want := range []string{"- backend/ — python", "- frontend/ — ts", "проба", "bin/demo"} {
		if !strings.Contains(got, want) {
			t.Errorf("CLAUDE.md lacks %q:\n%s", want, got)
		}
	}

	// Standard: an empty Slots range must leave no artifact — the line before
	// it flows straight into the Go build note.
	stdDir := t.TempDir()
	data.Slots = nil
	if _, err := render.Render(templatesFS, coreRoot, stdDir, data, nil); err != nil {
		t.Fatalf("render core (standard): %v", err)
	}
	got = read(t, filepath.Join(stdDir, "CLAUDE.md"))
	if strings.Contains(got, "jig does not impose one.\n\n-") || strings.Contains(got, "—  \n") {
		t.Errorf("empty slots range left an artifact:\n%s", got)
	}
	if !strings.Contains(got, "jig does not impose one.\n- Go: build into bin/") {
		t.Errorf("standard CLAUDE.md flow broken:\n%s", got)
	}
}

// --- streams -------------------------------------------------------------

func TestErrorsGoToStderrWithPrefix(t *testing.T) {
	h := newHarness(t)
	if got := h.run("--level", "standard", "--description", "d", "--no-commit", "myApp"); got != 2 {
		t.Fatalf("exit = %d, want 2", got)
	}
	if h.out.Len() != 0 {
		t.Errorf("error leaked to stdout: %s", h.out.String())
	}
	if !strings.HasPrefix(h.errOut.String(), "jig: ") {
		t.Errorf("stderr lacks the jig: prefix: %q", h.errOut.String())
	}
	if !strings.Contains(h.errOut.String(), `"myApp"`) {
		t.Errorf("stderr does not name the bad name: %q", h.errOut.String())
	}
}

func TestReportGoesToStdout(t *testing.T) {
	h := newHarness(t)
	if got := h.run("--level", "standard", "--description", "d", "--no-commit", "my-app"); got != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", got, h.errOut.String())
	}
	out := h.out.String()
	for _, want := range []string{"Created: ", "Ran: git init", `Project "my-app" is ready.`} {
		if !strings.Contains(out, want) {
			t.Errorf("report lacks %q\ngot:\n%s", want, out)
		}
	}
	if h.errOut.Len() != 0 {
		t.Errorf("stderr not empty: %s", h.errOut.String())
	}
}

// Initializer artefacts (go.mod, …) appear in "Created:" when they exist on disk.
func TestReportListsInitArtifacts(t *testing.T) {
	h := newHarness(t)
	h.runner.onRun = func(dir string, argv []string) {
		if strings.Join(argv, " ") == "go mod init my-app" {
			write(t, filepath.Join(dir, "go.mod"), "module my-app\n")
		}
	}
	code := h.run("--level", "full", "--type", "cli", "--lang", "go",
		"--description", "d", "--no-commit", "my-app")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", code, h.errOut.String())
	}
	out := h.out.String()
	if !strings.Contains(out, "Created: my-app/go.mod") {
		t.Errorf("report lacks go.mod:\n%s", out)
	}
	if !strings.Contains(out, "Created: my-app/main.go") {
		t.Errorf("report lacks main.go:\n%s", out)
	}
}

// Dry-run warns about missing tools but still prints the plan and exits 0.
func TestDryRunWarnsOnMissingTools(t *testing.T) {
	h := newHarness(t)
	h.d.look = func(name string) (string, error) {
		if name == "uv" {
			return "", os.ErrNotExist
		}
		return "/usr/bin/" + name, nil
	}
	code := h.run("--dry-run", "--level", "full", "--type", "cli", "--lang", "python",
		"--description", "d", "--no-commit", "my-app")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", code, h.errOut.String())
	}
	if !strings.Contains(h.errOut.String(), "uv") {
		t.Errorf("stderr lacks tool warning:\n%s", h.errOut.String())
	}
	// Harness env is empty → English.
	if !strings.Contains(h.errOut.String(), "Warning:") {
		t.Errorf("stderr is not a warning:\n%s", h.errOut.String())
	}
	if !strings.Contains(h.out.String(), "Dry run") {
		t.Errorf("stdout lacks plan:\n%s", h.out.String())
	}
	if entries := readDir(t, h.d.workdir); len(entries) != 0 {
		t.Errorf("dry run created files: %v", entries)
	}
}

func TestTypeOnStandardRejectedMessage(t *testing.T) {
	h := newHarness(t)
	code := h.run("--level", "standard", "--type", "cli",
		"--description", "d", "--no-commit", "my-app")
	if code != 2 {
		t.Fatalf("exit = %d, want 2\nstderr: %s", code, h.errOut.String())
	}
	if !strings.Contains(h.errOut.String(), "--type") {
		t.Errorf("stderr does not mention --type: %q", h.errOut.String())
	}
	if !strings.Contains(h.errOut.String(), "full") {
		t.Errorf("stderr does not mention full: %q", h.errOut.String())
	}
}

func TestUnknownLangSlotMessage(t *testing.T) {
	h := newHarness(t)
	code := h.run("--level", "full", "--type", "cli", "--lang", "backend=go",
		"--description", "d", "--no-commit", "my-app")
	if code != 2 {
		t.Fatalf("exit = %d, want 2\nstderr: %s", code, h.errOut.String())
	}
	msg := h.errOut.String()
	if !strings.Contains(msg, "backend") {
		t.Errorf("stderr does not name the slot: %q", msg)
	}
	if strings.Contains(msg, "Missing flags") || strings.Contains(msg, "Не хватает") {
		t.Errorf("unknown slot reported as missing flags: %q", msg)
	}
}

func TestGoKeywordNameRejected(t *testing.T) {
	h := newHarness(t)
	code := h.run("--level", "full", "--type", "library", "--lang", "go",
		"--description", "d", "--no-commit", "type")
	if code != 2 {
		t.Fatalf("exit = %d, want 2\nstderr: %s", code, h.errOut.String())
	}
	if !strings.Contains(h.errOut.String(), "type") {
		t.Errorf("stderr does not name the project: %q", h.errOut.String())
	}
}

// A Go-keyword name is allowed when no slot uses Go.
func TestGoKeywordNameAllowedForPython(t *testing.T) {
	h := newHarness(t)
	code := h.run("--level", "full", "--type", "library", "--lang", "python",
		"--description", "d", "--no-commit", "type")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", code, h.errOut.String())
	}
}

// --- initializer failure -------------------------------------------------

func TestInitializerFailureLeavesDirAsIs(t *testing.T) {
	h := newHarness(t)
	h.runner.err["npm create vite@latest . -- --template vanilla-ts"] = io.ErrUnexpectedEOF

	code := h.run("--level", "full", "--type", "web",
		"--lang", "backend=python", "--lang", "frontend=ts",
		"--description", "d", "--commit", "my-app")
	if code != 1 {
		t.Fatalf("exit = %d, want 1\nstderr: %s", code, h.errOut.String())
	}
	if !strings.Contains(h.errOut.String(), `slot "frontend"`) {
		t.Errorf("stderr does not name the slot: %q", h.errOut.String())
	}
	// No rollback: what was written stays.
	if _, err := os.Stat(filepath.Join(h.d.workdir, "my-app", "CLAUDE.md")); err != nil {
		t.Errorf("directory was cleaned up: %v", err)
	}
	if h.runner.ran("git commit -m " + commitMsg) {
		t.Error("committed after a failed initializer")
	}
}

// --- --lang parsing ------------------------------------------------------

func TestLangFlagIsRepeatable(t *testing.T) {
	h := newHarness(t)
	code := h.run("--level", "full", "--type", "web",
		"--lang", "backend=go", "--lang", "frontend=js",
		"--description", "d", "--no-commit", "my-app")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", code, h.errOut.String())
	}
	lines := h.runner.lines()
	if !slices.Contains(lines, "go mod init my-app") {
		t.Errorf("backend=go not honoured: %v", lines)
	}
	if !slices.Contains(lines, "npm create vite@latest . -- --template vanilla") {
		t.Errorf("frontend=js not honoured: %v", lines)
	}
}

func TestBareLangGoesToSingleSlot(t *testing.T) {
	v := &langsValue{m: map[string]string{}}
	if err := v.Set("go"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := v.Set("backend=python"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if v.m["."] != "go" || v.m["backend"] != "python" {
		t.Fatalf("langs = %v", v.m)
	}
	if err := v.Set("backend="); err == nil {
		t.Fatal("empty language accepted")
	}

	// End to end: a bare --lang drives the single slot of type cli.
	h := newHarness(t)
	code := h.run("--level", "full", "--type", "cli", "--lang", "python",
		"--description", "d", "--no-commit", "my-app")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", code, h.errOut.String())
	}
	if !h.runner.ran("uv init") {
		t.Fatalf("bare --lang python not honoured: %v", h.runner.lines())
	}
}

func TestSlotInfosCarriesOnlyNamedSlots(t *testing.T) {
	dec := interview.Decision{Langs: map[string]string{
		".": "go", "backend": "python", "frontend": "ts",
	}}

	single := manifest.Type{Slots: []manifest.Slot{{Dir: "."}}}
	if got := slotInfos(single, dec); got != nil {
		t.Errorf("slotInfos for a single-dot slot = %v, want nil", got)
	}

	if got := slotInfos(manifest.Type{}, dec); got != nil {
		t.Errorf("slotInfos for no slots = %v, want nil", got)
	}

	web := manifest.Type{Slots: []manifest.Slot{{Dir: "backend"}, {Dir: "frontend"}}}
	want := []render.SlotInfo{{Dir: "backend", Lang: "python"}, {Dir: "frontend", Lang: "ts"}}
	if got := slotInfos(web, dec); !slices.Equal(got, want) {
		t.Errorf("slotInfos for web = %v, want %v", got, want)
	}
}

// --- full pipeline order -------------------------------------------------

func TestFullRunOrder(t *testing.T) {
	h := newHarness(t)
	code := h.run("--level", "full", "--type", "web",
		"--lang", "backend=python", "--lang", "frontend=ts",
		"--description", "демо", "--commit", "demo")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", code, h.errOut.String())
	}

	// git config calls are preflight noise; the pipeline is what remains.
	var got []string
	for _, l := range h.runner.lines() {
		if strings.HasPrefix(l, "git config ") {
			continue
		}
		got = append(got, l)
	}
	want := []string{
		"git init",
		"uv init",
		"npm create vite@latest . -- --template vanilla-ts",
		"git add -A",
		"git commit -m " + commitMsg,
	}
	if !slices.Equal(got, want) {
		t.Fatalf("commands =\n%v\nwant\n%v", got, want)
	}

	// Each initializer runs in its own slot directory.
	target := filepath.Join(h.d.workdir, "demo")
	dirs := map[string]string{
		"uv init": filepath.Join(target, "backend"),
		"npm create vite@latest . -- --template vanilla-ts": filepath.Join(target, "frontend"),
	}
	for _, c := range h.runner.calls {
		if want, ok := dirs[c.line()]; ok && c.dir != want {
			t.Errorf("%q ran in %s, want %s", c.line(), c.dir, want)
		}
	}

	// Substitution reached the rendered core.
	body := read(t, filepath.Join(target, "CLAUDE.md"))
	for _, want := range []string{"demo", "демо", "2026-07-14", "Test User"} {
		if !strings.Contains(body, want) {
			t.Errorf("CLAUDE.md lacks %q\ngot:\n%s", want, body)
		}
	}
	// gitignore regained its leading dot.
	if _, err := os.Stat(filepath.Join(target, ".gitignore")); err != nil {
		t.Errorf(".gitignore missing: %v", err)
	}
}

func TestStandardRunsNoInitializers(t *testing.T) {
	h := newHarness(t)
	if got := h.run("--level", "standard", "--description", "d", "--no-commit", "my-app"); got != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", got, h.errOut.String())
	}
	var got []string
	for _, l := range h.runner.lines() {
		if strings.HasPrefix(l, "git config ") {
			continue
		}
		got = append(got, l)
	}
	if !slices.Equal(got, []string{"git init"}) {
		t.Fatalf("commands = %v, want only git init", got)
	}
	// A standard run still checks git, and nothing else.
	if _, err := os.Stat(filepath.Join(h.d.workdir, "my-app", "CLAUDE.md")); err != nil {
		t.Errorf("core not rendered: %v", err)
	}
}

// --- interview wiring ----------------------------------------------------

func TestInteractiveRunAsksWhatFlagsLeftOpen(t *testing.T) {
	h := newHarness(t)
	h.d.interactive = true
	// level=full, name, description, type=cli, lang=default(go), commit=no
	h.d.stdin = strings.NewReader("2\nmy-app\nдемо\n1\n\n2\n")

	if got := run(nil, h.d); got != 0 {
		t.Fatalf("exit = %d, want 0\nstdout: %s\nstderr: %s", got, h.out.String(), h.errOut.String())
	}
	if !h.runner.ran("go mod init my-app") {
		t.Fatalf("default language not taken: %v", h.runner.lines())
	}
	if h.runner.ran("git commit -m " + commitMsg) {
		t.Fatalf("committed despite answering no: %v", h.runner.lines())
	}
	if !strings.Contains(h.out.String(), "Project name?") {
		t.Errorf("interview did not reach stdout:\n%s", h.out.String())
	}
}

func TestSetFlagRemovesItsQuestion(t *testing.T) {
	h := newHarness(t)
	h.d.interactive = true
	// Only the commit question is left open.
	h.d.stdin = strings.NewReader("2\n")

	code := run([]string{"--level", "standard", "--description", "d", "my-app"}, h.d)
	if code != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", code, h.errOut.String())
	}
	out := h.out.String()
	for _, unwanted := range []string{"What are we creating?", "Project name?", "Project description?"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("asked about a flag that was set: %q\n%s", unwanted, out)
		}
	}
	if !strings.Contains(out, "Make the first commit?") {
		t.Errorf("the open question was not asked:\n%s", out)
	}
}

func TestSlotQuestionsMapManifest(t *testing.T) {
	typ, err := manifest.Load(testFS(), typesRoot, "web")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	qs := slotQuestions(typ)
	if len(qs) != 2 {
		t.Fatalf("questions = %d, want 2", len(qs))
	}
	if qs[0].Dir != "backend" || qs[0].DefaultLang != "python" {
		t.Errorf("slot 0 = %+v", qs[0])
	}
	// The default comes first, the rest alphabetically: the user reads the
	// list top down, and 1) must agree with Enter.
	if !slices.Equal(qs[0].Langs, []string{"python", "go"}) {
		t.Errorf("langs = %v, want default first [python go]", qs[0].Langs)
	}
	if qs[1].Dir != "frontend" || qs[1].DefaultLang != "ts" {
		t.Errorf("slot 1 = %+v", qs[1])
	}
	if !slices.Equal(qs[1].Langs, []string{"ts", "js"}) {
		t.Errorf("langs = %v, want default first [ts js]", qs[1].Langs)
	}
}

// --- misc ----------------------------------------------------------------

func TestCWDName(t *testing.T) {
	tests := []struct{ workdir, want string }{
		{"/home/kir/my-app", "my-app"},
		{"/home/kir/MyApp", ""},
		{"/home/kir/1st", ""},
		{"/home/kir/.hidden", ""},
	}
	for _, tc := range tests {
		if got := cwdName(tc.workdir); got != tc.want {
			t.Errorf("cwdName(%q) = %q, want %q", tc.workdir, got, tc.want)
		}
	}
}

func TestMissingFlagsAreListed(t *testing.T) {
	h := newHarness(t)
	if got := h.run("my-app"); got != 2 {
		t.Fatalf("exit = %d, want 2", got)
	}
	msg := h.errOut.String()
	for _, want := range []string{"--level", "--description", "--commit"} {
		if !strings.Contains(msg, want) {
			t.Errorf("stderr does not list %q: %q", want, msg)
		}
	}
}

func TestUILangOverridesEnvironment(t *testing.T) {
	h := newHarness(t)
	h.d.env = func(k string) string {
		if k == "LANG" {
			return "en_US.UTF-8"
		}
		return ""
	}
	if got := h.run("--ui-lang", "ru", "--level", "standard", "--description", "d", "--no-commit", "my-app"); got != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", got, h.errOut.String())
	}
	if !strings.Contains(h.out.String(), "готов") {
		t.Fatalf("report is not Russian:\n%s", h.out.String())
	}
}

// --- helpers -------------------------------------------------------------

func mkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func readDir(t *testing.T, dir string) []fs.DirEntry {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	return entries
}
