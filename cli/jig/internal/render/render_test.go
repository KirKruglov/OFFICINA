package render

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

// sampleData is a fully populated Data value used by most tests.
func sampleData() Data {
	return Data{
		Name:        "demo",
		PackageName: "demo",
		Description: "демо проект",
		Date:        "2026-07-14",
		Author:      "Test User",
		Slots: []SlotInfo{
			{Dir: "cmd", Lang: "go"},
			{Dir: "scripts", Lang: "python"},
		},
	}
}

// coreFS mirrors the example tree from the spec's "Контракт ядра шаблона".
func coreFS() fstest.MapFS {
	return fstest.MapFS{
		"templates/core/CLAUDE.md.tmpl": &fstest.MapFile{
			Data: []byte("# {{.Name}}\n{{.Description}}\nАвтор: {{.Author}}\nСоздан: {{.Date}}\n" +
				"Пакет: {{.PackageName}}\nСлоты: {{range .Slots}}{{.Dir}}={{.Lang}} {{end}}\n"),
		},
		"templates/core/gitignore": &fstest.MapFile{
			Data: []byte("/bin\n"),
		},
		"templates/core/README.md": &fstest.MapFile{
			Data: []byte("literal {{.Name}} stays\n"),
		},
		"templates/core/.claude/settings.json": &fstest.MapFile{
			Data: []byte("{\n  \"permissions\": { \"allow\": [\"Bash({{.Name}})\"] }\n}\n"),
		},
		"templates/core/sub/gitignore": &fstest.MapFile{
			Data: []byte("node_modules\n"),
		},
		"templates/core/docs/adr/.gitkeep": &fstest.MapFile{
			Data: []byte(""),
		},
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func TestRenderTmplSuffixDroppedAndVarsSubstituted(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "out")
	if _, err := Render(coreFS(), "templates/core", dir, sampleData(), nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md.tmpl")); !os.IsNotExist(err) {
		t.Errorf("CLAUDE.md.tmpl must not exist in output, stat err = %v", err)
	}

	got := read(t, filepath.Join(dir, "CLAUDE.md"))
	want := "# demo\nдемо проект\nАвтор: Test User\nСоздан: 2026-07-14\n" +
		"Пакет: demo\nСлоты: cmd=go scripts=python \n"
	if got != want {
		t.Errorf("CLAUDE.md =\n%q\nwant\n%q", got, want)
	}
}

func TestRenderStandardLevelHasNoSlots(t *testing.T) {
	fsys := fstest.MapFS{
		"core/CLAUDE.md.tmpl": &fstest.MapFile{
			Data: []byte("Слоты: {{range .Slots}}{{.Dir}}={{.Lang}} {{end}}\n"),
		},
	}
	dir := filepath.Join(t.TempDir(), "out")

	if _, err := Render(fsys, "core", dir, Data{Name: "demo", PackageName: "demo"}, nil); err != nil {
		t.Fatalf("Render with empty Slots: %v", err)
	}

	got := read(t, filepath.Join(dir, "CLAUDE.md"))
	want := "Слоты: \n"
	if got != want {
		t.Errorf("CLAUDE.md = %q, want %q", got, want)
	}
}

func TestRenderTypeVariableIsGone(t *testing.T) {
	fsys := fstest.MapFS{
		"core/CLAUDE.md.tmpl": &fstest.MapFile{Data: []byte("{{.Type}}\n")},
	}
	dir := filepath.Join(t.TempDir(), "out")

	_, err := Render(fsys, "core", dir, sampleData(), nil)
	if err == nil {
		t.Fatal("Render with {{.Type}}: want error, got nil")
	}
	if !strings.Contains(err.Error(), "Type") {
		t.Errorf("error should mention the removed field, got: %v", err)
	}
}

func TestPackageName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"probe-lib", "probelib"},
		{"demo", "demo"},
		{"a-b-c", "abc"},
	}
	for _, tt := range tests {
		if got := PackageName(tt.name); got != tt.want {
			t.Errorf("PackageName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestIsGoKeyword(t *testing.T) {
	for _, word := range []string{"type", "func", "var", "go", "package", "interface", "map"} {
		if !IsGoKeyword(word) {
			t.Errorf("IsGoKeyword(%q) = false, want true", word)
		}
		// PackageName leaves a bare keyword unchanged.
		if !IsGoKeyword(PackageName(word)) {
			t.Errorf("IsGoKeyword(PackageName(%q)) = false, want true", word)
		}
	}
	for _, word := range []string{"demo", "probelib", "mytype", "golang"} {
		if IsGoKeyword(word) {
			t.Errorf("IsGoKeyword(%q) = true, want false", word)
		}
	}
	// Hyphenated names that are not themselves keywords stay free after strip.
	if IsGoKeyword(PackageName("my-type")) {
		t.Error("PackageName(\"my-type\") must not be treated as a keyword")
	}
}

func TestRenderNonTmplCopiedByteForByte(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "out")
	if _, err := Render(coreFS(), "templates/core", dir, sampleData(), nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	got := read(t, filepath.Join(dir, "README.md"))
	want := "literal {{.Name}} stays\n"
	if got != want {
		t.Errorf("README.md = %q, want %q (must not be rendered)", got, want)
	}
}

func TestRenderGitignoreAtRootAndNested(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "out")
	if _, err := Render(coreFS(), "templates/core", dir, sampleData(), nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	if got := read(t, filepath.Join(dir, ".gitignore")); got != "/bin\n" {
		t.Errorf("root .gitignore = %q", got)
	}
	if got := read(t, filepath.Join(dir, "sub", ".gitignore")); got != "node_modules\n" {
		t.Errorf("sub/.gitignore = %q", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "gitignore")); !os.IsNotExist(err) {
		t.Errorf("plain gitignore must not exist, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "sub", "gitignore")); !os.IsNotExist(err) {
		t.Errorf("plain sub/gitignore must not exist, stat err = %v", err)
	}
}

func TestRenderNestingPreserved(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "out")
	if _, err := Render(coreFS(), "templates/core", dir, sampleData(), nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "docs", "adr", ".gitkeep")); err != nil {
		t.Errorf("docs/adr/.gitkeep: %v", err)
	}
}

func TestRenderDotClaudeSettingsCopiedAsIs(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "out")
	if _, err := Render(coreFS(), "templates/core", dir, sampleData(), nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	got := read(t, filepath.Join(dir, ".claude", "settings.json"))
	want := "{\n  \"permissions\": { \"allow\": [\"Bash({{.Name}})\"] }\n}\n"
	if got != want {
		t.Errorf(".claude/settings.json = %q, want %q", got, want)
	}
}

func TestRenderEmptyOptionalValuesAreNotAnError(t *testing.T) {
	fsys := fstest.MapFS{
		"core/CLAUDE.md.tmpl": &fstest.MapFile{
			Data: []byte("Автор: {{.Author}}\nОписание: {{.Description}}\n"),
		},
	}
	dir := filepath.Join(t.TempDir(), "out")

	if _, err := Render(fsys, "core", dir, Data{Name: "demo", Date: "2026-07-14"}, nil); err != nil {
		t.Fatalf("Render with empty Description/Author: %v", err)
	}

	got := read(t, filepath.Join(dir, "CLAUDE.md"))
	want := "Автор: \nОписание: \n"
	if got != want {
		t.Errorf("CLAUDE.md = %q, want %q", got, want)
	}
}

func TestRenderUndefinedVariableIsError(t *testing.T) {
	fsys := fstest.MapFS{
		"core/CLAUDE.md.tmpl": &fstest.MapFile{Data: []byte("{{.Bogus}}\n")},
	}
	dir := filepath.Join(t.TempDir(), "out")

	_, err := Render(fsys, "core", dir, sampleData(), nil)
	if err == nil {
		t.Fatal("Render with {{.Bogus}}: want error, got nil")
	}
	if !strings.Contains(err.Error(), "Bogus") {
		t.Errorf("error should mention the bad field, got: %v", err)
	}
}

func TestRenderReturnsDestinationRelativePaths(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "out")
	got, err := Render(coreFS(), "templates/core", dir, sampleData(), nil)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// fs.WalkDir order is lexical by SOURCE path, so ".gitignore" lands
	// where "gitignore" sorted, not where ".gitignore" would.
	want := []string{
		".claude/settings.json",
		"CLAUDE.md",
		"README.md",
		"docs/adr/.gitkeep",
		".gitignore",
		"sub/.gitignore",
	}
	if len(got) != len(want) {
		t.Fatalf("Render returned %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("path[%d] = %q, want %q (full: %v)", i, got[i], want[i], got)
		}
	}
}

func TestPlanMatchesRenderAndWritesNothing(t *testing.T) {
	base := t.TempDir()
	planDir := filepath.Join(base, "plan-out")

	planned, err := Plan(coreFS(), "templates/core", nil)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if _, err := os.Stat(planDir); !os.IsNotExist(err) {
		t.Errorf("Plan must not create anything, stat err = %v", err)
	}
	if entries, err := os.ReadDir(base); err != nil || len(entries) != 0 {
		t.Errorf("temp dir not empty after Plan: entries=%v err=%v", entries, err)
	}

	rendered, err := Render(coreFS(), "templates/core", filepath.Join(base, "render-out"), sampleData(), nil)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(planned) != len(rendered) {
		t.Fatalf("Plan %v != Render %v", planned, rendered)
	}
	for i := range rendered {
		if planned[i] != rendered[i] {
			t.Errorf("path[%d]: Plan %q != Render %q", i, planned[i], rendered[i])
		}
	}
}

func TestRenderCreatesDestinationDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "deep", "out")
	if _, err := Render(coreFS(), "templates/core", dir, sampleData(), nil); err != nil {
		t.Fatalf("Render: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("%s is not a directory", dir)
	}
}

func TestRenderMissingRootIsError(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "out")
	if _, err := Render(coreFS(), "templates/nope", dir, sampleData(), nil); err == nil {
		t.Fatal("want error for missing root, got nil")
	}
	if _, err := Plan(coreFS(), "templates/nope", nil); err == nil {
		t.Fatal("Plan: want error for missing root, got nil")
	}
}

func TestSkipMatchesBySegment(t *testing.T) {
	skip := []string{"docs/product/release-0.1.0", "docs/marketing"}
	tests := []struct {
		rel  string
		want bool
	}{
		{"docs/product/release-0.1.0", true},
		{"docs/product/release-0.1.0/spec/x.md", true},
		{"docs/marketing/CLAUDE.md", true},
		{"docs/product/release-0.1.0-notes", false},
		{"docs/product/release-0.1.0-notes/x.md", false},
		{"docs/marketing-extra/x.md", false},
		{"docs/product", false},
		{"CLAUDE.md", false},
	}
	for _, tt := range tests {
		if got := skipped(tt.rel, skip); got != tt.want {
			t.Errorf("skipped(%q, %v) = %v, want %v", tt.rel, skip, got, tt.want)
		}
	}
	if skipped("anything", nil) {
		t.Error("skipped with nil list must be false")
	}
}

func skipFS() fstest.MapFS {
	return fstest.MapFS{
		"core/CLAUDE.md":                &fstest.MapFile{Data: []byte("keep\n")},
		"core/docs/planning/BACKLOG.md": &fstest.MapFile{Data: []byte("keep\n")},
		"core/docs/marketing/CLAUDE.md": &fstest.MapFile{Data: []byte("full only\n")},
		"core/docs/marketing-extra/x":   &fstest.MapFile{Data: []byte("keep\n")},
	}
}

func TestRenderSkipsListedPaths(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "out")
	created, err := Render(skipFS(), "core", dir, sampleData(), []string{"docs/marketing"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "docs", "marketing")); !os.IsNotExist(err) {
		t.Errorf("docs/marketing must not exist, stat err = %v", err)
	}
	for _, p := range []string{"CLAUDE.md", "docs/planning/BACKLOG.md", "docs/marketing-extra/x"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(p))); err != nil {
			t.Errorf("%s: %v", p, err)
		}
	}
	if slices.Contains(created, "docs/marketing/CLAUDE.md") {
		t.Errorf("created reports a skipped path: %v", created)
	}
}

func TestPlanMatchesRenderWithSkip(t *testing.T) {
	skip := []string{"docs/marketing"}
	planned, err := Plan(skipFS(), "core", skip)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	rendered, err := Render(skipFS(), "core", filepath.Join(t.TempDir(), "out"), sampleData(), skip)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !slices.Equal(planned, rendered) {
		t.Errorf("Plan %v != Render %v", planned, rendered)
	}
}

// packFS is a language pack: an ignore block plus an entry-point template.
func packFS() fstest.MapFS {
	return fstest.MapFS{
		"langs/go/gitignore":    &fstest.MapFile{Data: []byte("# Go\n/bin/\n")},
		"langs/go/main.go.tmpl": &fstest.MapFile{Data: []byte("package main // {{.Name}}\n")},
	}
}

func TestRenderMergeAppendsToExistingGitignore(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(".DS_Store\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	written, err := RenderMerge(packFS(), "langs/go", dir, sampleData())
	if err != nil {
		t.Fatalf("RenderMerge: %v", err)
	}

	got := read(t, filepath.Join(dir, ".gitignore"))
	want := ".DS_Store\n\n# Go\n/bin/\n"
	if got != want {
		t.Errorf(".gitignore = %q, want %q", got, want)
	}
	if !slices.Contains(written, ".gitignore") {
		t.Errorf("appended .gitignore missing from report: %v", written)
	}
}

func TestRenderMergeAppendSeparatesWithBlankLine(t *testing.T) {
	dir := t.TempDir()
	// No trailing newline: the merge must still leave one blank line between.
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(".DS_Store"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := RenderMerge(packFS(), "langs/go", dir, sampleData()); err != nil {
		t.Fatalf("RenderMerge: %v", err)
	}

	got := read(t, filepath.Join(dir, ".gitignore"))
	want := ".DS_Store\n\n# Go\n/bin/\n"
	if got != want {
		t.Errorf(".gitignore = %q, want %q", got, want)
	}
}

func TestRenderMergeSkipsOtherExistingFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("original\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	written, err := RenderMerge(packFS(), "langs/go", dir, sampleData())
	if err != nil {
		t.Fatalf("RenderMerge: %v", err)
	}

	if got := read(t, filepath.Join(dir, "main.go")); got != "original\n" {
		t.Errorf("existing main.go overwritten: %q", got)
	}
	if slices.Contains(written, "main.go") {
		t.Errorf("skipped file reported as written: %v", written)
	}
	// The pack's .gitignore had no target yet — created as usual.
	if got := read(t, filepath.Join(dir, ".gitignore")); got != "# Go\n/bin/\n" {
		t.Errorf(".gitignore = %q", got)
	}
}

func TestRenderMergeWritesFreshFilesLikeRender(t *testing.T) {
	dir := t.TempDir()
	written, err := RenderMerge(packFS(), "langs/go", dir, sampleData())
	if err != nil {
		t.Fatalf("RenderMerge: %v", err)
	}

	if got := read(t, filepath.Join(dir, "main.go")); got != "package main // demo\n" {
		t.Errorf("main.go = %q", got)
	}
	if got := read(t, filepath.Join(dir, ".gitignore")); got != "# Go\n/bin/\n" {
		t.Errorf(".gitignore = %q", got)
	}
	want := []string{".gitignore", "main.go"}
	slices.Sort(written)
	if !slices.Equal(written, want) {
		t.Errorf("written = %v, want %v", written, want)
	}
}

func TestRenderRestoresDotNames(t *testing.T) {
	fsys := fstest.MapFS{
		"core/claude/settings.json":  &fstest.MapFile{Data: []byte("{}\n")},
		"core/claude/rules/style.md": &fstest.MapFile{Data: []byte("style\n")},
		"core/docs/archive/gitkeep":  &fstest.MapFile{Data: []byte("")},
	}
	dir := filepath.Join(t.TempDir(), "out")
	if _, err := Render(fsys, "core", dir, sampleData(), nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	for _, p := range []string{
		".claude/settings.json",
		".claude/rules/style.md",
		"docs/archive/.gitkeep",
	} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(p))); err != nil {
			t.Errorf("%s: %v", p, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "claude")); !os.IsNotExist(err) {
		t.Errorf("plain claude/ must not exist, stat err = %v", err)
	}
}

func TestTargetName(t *testing.T) {
	tests := []struct {
		rel  string
		want string
	}{
		{"CLAUDE.md.tmpl", "CLAUDE.md"},
		{"gitignore", ".gitignore"},
		{"sub/gitignore", "sub/.gitignore"},
		{"README.md", "README.md"},
		{".claude/settings.json", ".claude/settings.json"},
		{"gitignore.tmpl", ".gitignore"},
		{"docs/adr/.gitkeep", "docs/adr/.gitkeep"},
		{".env.example", ".env.example"},
		{"a/b/c/CLAUDE.md.tmpl", "a/b/c/CLAUDE.md"},
		{"gitignore-sample", "gitignore-sample"},
		{"sub/my.gitignore", "sub/my.gitignore"},
		{"claude/settings.json", ".claude/settings.json"},
		{"claude/rules/writing-style.md", ".claude/rules/writing-style.md"},
		{"gitkeep", ".gitkeep"},
		{"docs/archive/gitkeep", "docs/archive/.gitkeep"},
		{"gitkeep.tmpl", ".gitkeep"},
		{"sub/claude/x.md", "sub/claude/x.md"},
		{"claude-notes/x.md", "claude-notes/x.md"},
		{"sub/my.gitkeep", "sub/my.gitkeep"},
		{"gitkeep-sample", "gitkeep-sample"},
	}
	for _, tt := range tests {
		if got := targetName(tt.rel); got != tt.want {
			t.Errorf("targetName(%q) = %q, want %q", tt.rel, got, tt.want)
		}
	}
}
