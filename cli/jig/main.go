// jig — set up a ready-to-work local repository in a single run.
//
// main.go owns the three things only an entry point can own: flags, output and
// exit codes. Everything else lives in internal/. run takes argv and deps so
// that every test injects fakes: no globals, no real filesystem outside the
// target directory, no real command execution in unit tests.
package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"jig/internal/i18n"
	"jig/internal/interview"
	"jig/internal/manifest"
	"jig/internal/render"
	"jig/internal/toolchain"
)

//go:embed templates
var templatesFS embed.FS

// version is the release version, baked into the source. Bump it here on each
// release; a plain `go build` carries it into the binary.
var version = "0.2.0"

const (
	// Roots inside the embedded tree. //go:embed templates keeps the
	// "templates/" prefix on every path.
	typesRoot    = "templates/types"
	coreRoot     = "templates/core"
	langsRoot    = "templates/langs"
	typedocsRoot = "templates/typedocs"

	// fullOnlyPath is the manifest of core paths rendered only on the full
	// level. It lives beside templates/core, not inside it, so the core walk
	// needs no exception for it.
	fullOnlyPath = "templates/full-only.json"

	// The first commit message is always English: the interface language is a
	// property of the machine, the commit message stays in the repository.
	commitMsg = "chore: initial scaffold from jig"

	dateLayout = "2006-01-02"

	// singleSlot is the slot dir of a type that has exactly one. A bare
	// --lang go is stored under this key.
	singleSlot = "."
)

// knownInitArtifacts are the primary files initializers leave behind. After a
// successful slot command, those that exist are added to the "Created:" report
// so the user sees go.mod / pyproject.toml / package.json alongside templates.
var knownInitArtifacts = []string{"go.mod", "pyproject.toml", "package.json"}

// deps is everything run touches from the outside world.
type deps struct {
	stdin       io.Reader
	stdout      io.Writer
	stderr      io.Writer
	env         func(string) string
	runner      toolchain.Runner
	look        toolchain.Looker
	templates   fs.FS
	now         func() time.Time
	workdir     string
	interactive bool
}

func main() { os.Exit(run(os.Args[1:], realDeps())) }

func realDeps() deps {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	interactive := false
	if fi, err := os.Stdin.Stat(); err == nil && fi.Mode()&os.ModeCharDevice != 0 {
		interactive = true
	}
	return deps{
		stdin:       os.Stdin,
		stdout:      os.Stdout,
		stderr:      os.Stderr,
		env:         os.Getenv,
		runner:      toolchain.ExecRunner{Stdout: os.Stdout, Stderr: os.Stderr},
		look:        exec.LookPath,
		templates:   templatesFS,
		now:         time.Now,
		workdir:     wd,
		interactive: interactive,
	}
}

// options is the parsed command line, before the interview fills the gaps.
type options struct {
	level       string
	name        string
	description string
	hasDesc     bool
	typ         string
	langs       map[string]string
	commit      bool
	noCommit    bool
	uiLang      string
	dryRun      bool
	showVersion bool
}

// langsValue implements flag.Value for the repeatable --lang.
//
// --lang backend=go names its slot. A type with a single slot takes the bare
// form --lang go, stored under ".", because the slot name is never shown to the
// user and --lang .=go would be asking them to type an implementation detail;
// the explicit .=<lang> form is therefore rejected outright.
type langsValue struct{ m map[string]string }

func (v *langsValue) String() string { return "" }

func (v *langsValue) Set(s string) error {
	slot, lang := singleSlot, s
	if i := strings.Index(s, "="); i >= 0 {
		slot, lang = s[:i], s[i+1:]
		if slot == singleSlot {
			return fmt.Errorf("slot %q is not addressable: use the bare form --lang %s", singleSlot, lang)
		}
	}
	if slot == "" || lang == "" {
		return fmt.Errorf("want <slot>=<lang> or <lang>, got %q", s)
	}
	v.m[slot] = lang
	return nil
}

// cmd is one external command jig plans to run: its argv and the directory,
// relative to the project root, it runs in. files names the language packs
// rendered into that directory once argv succeeds.
type cmd struct {
	dir   string
	argv  []string
	files []string
}

func run(argv []string, d deps) int {
	o, lang, stop, code := parseFlags(argv, d)
	if stop {
		return code
	}

	types, err := manifest.List(d.templates, typesRoot)
	if err != nil {
		return fail(d, 1, err.Error())
	}
	if o.typ != "" && !slices.Contains(types, o.typ) {
		return fail(d, 2, i18n.Tf(lang, "err.type.unknown", o.typ))
	}

	dec, code := decide(o, lang, types, d)
	if code != 0 {
		return code
	}

	// --type is full-only: on standard it would otherwise be silently dropped.
	if dec.Level != interview.LevelFull && o.typ != "" {
		return fail(d, 2, i18n.T(lang, "err.type.fullonly"))
	}

	typ, code := resolveType(dec, lang, types, d)
	if code != 0 {
		return code
	}

	// Unknown --lang slots must fail before anything is written. Interview
	// already dropped them from dec.Langs, so the check uses the raw flags.
	if code := validateLangSlots(o.langs, typ, lang, d); code != 0 {
		return code
	}

	// library-go (and any future pack) puts PackageName into "package …".
	// A Go keyword there is not a valid package clause.
	if usesGo(dec) && render.IsGoKeyword(render.PackageName(dec.Name)) {
		return fail(d, 2, i18n.Tf(lang, "err.name.go_keyword", dec.Name))
	}

	data := render.Data{
		Name:        dec.Name,
		Description: dec.Description,
		Date:        d.now().Format(dateLayout),
		PackageName: render.PackageName(dec.Name),
		Slots:       slotInfos(typ, dec),
	}

	// Slot commands are rendered before anything is written: a broken argument
	// template must not leave a half-built directory behind.
	slotCmds := make([]cmd, 0, len(typ.Slots))
	for _, s := range typ.Slots {
		argvSlot, err := s.Command(dec.Langs[s.Dir], data)
		if err != nil {
			return fail(d, 1, err.Error())
		}
		slotCmds = append(slotCmds, cmd{
			dir:   s.Dir,
			argv:  argvSlot,
			files: s.FilesFor(dec.Langs[s.Dir]),
		})
	}

	fullOnly, err := fullOnlyList(d.templates)
	if err != nil {
		return fail(d, 1, err.Error())
	}
	skip := coreSkip(dec.Level, fullOnly)

	if o.dryRun {
		return dryRun(d, lang, dec, typ, slotCmds, skip)
	}

	target := filepath.Join(d.workdir, dec.Name)
	if code := preflight(d, lang, dec, target, slotCmds); code != 0 {
		return code
	}

	// GitAuthor runs after the identity check and never fails the run: an empty
	// .Author is legal whenever no commit is requested.
	author, _ := toolchain.GitAuthor(d.runner, d.workdir)
	data.Author = author

	return build(d, lang, dec, typ, target, data, slotCmds, skip)
}

// fullOnlyList reads the full-only manifest from the embedded tree.
func fullOnlyList(fsys fs.FS) ([]string, error) {
	b, err := fs.ReadFile(fsys, fullOnlyPath)
	if err != nil {
		return nil, err
	}
	var m struct {
		FullOnly []string `json:"full_only"`
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("%s: %w", fullOnlyPath, err)
	}
	return m.FullOnly, nil
}

// coreSkip decides, once for both the real run and dry-run, which core paths
// the level leaves out: full renders everything, standard skips the full-only
// list. One function so the plan and the written tree can never diverge.
func coreSkip(level string, fullOnly []string) []string {
	if level == interview.LevelFull {
		return nil
	}
	return fullOnly
}

// parseFlags reads argv. It returns stop=true when run must not continue:
// either --help was asked for (code 0) or the command line is unusable (2).
func parseFlags(argv []string, d deps) (options, i18n.Lang, bool, int) {
	o := options{langs: map[string]string{}}

	fset := flag.NewFlagSet("jig", flag.ContinueOnError)
	fset.SetOutput(io.Discard) // errors and usage are ours, and bilingual
	fset.Usage = func() {}
	fset.StringVar(&o.level, "level", "", "")
	fset.StringVar(&o.description, "description", "", "")
	fset.StringVar(&o.typ, "type", "", "")
	fset.Var(&langsValue{m: o.langs}, "lang", "")
	fset.BoolVar(&o.commit, "commit", false, "")
	fset.BoolVar(&o.noCommit, "no-commit", false, "")
	fset.StringVar(&o.uiLang, "ui-lang", "", "")
	fset.BoolVar(&o.dryRun, "dry-run", false, "")
	fset.BoolVar(&o.showVersion, "version", false, "")

	err := fset.Parse(argv)

	// The language is resolved even on a parse error: --ui-lang may well have
	// been read before the bad flag, and the error text is bilingual too.
	lang := i18n.Detect(o.uiLang, d.env("LC_ALL"), d.env("LANG"))

	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printUsage(d.stdout, lang)
			return o, lang, true, 0
		}
		fmt.Fprintf(d.stderr, "jig: %v\n", err)
		printUsage(d.stderr, lang)
		return o, lang, true, 2
	}

	// The version is a value, not interface text: it is not translated, and it
	// answers before any other flag is judged.
	if o.showVersion {
		fmt.Fprintf(d.stdout, "jig %s\n", version)
		return o, lang, true, 0
	}

	// --description "" is a value; an unset --description is a question.
	fset.Visit(func(f *flag.Flag) {
		if f.Name == "description" {
			o.hasDesc = true
		}
	})

	if o.commit && o.noCommit {
		fmt.Fprintln(d.stderr, "jig: "+i18n.T(lang, "err.commit.conflict"))
		return o, lang, true, 2
	}
	if o.level != "" && o.level != interview.LevelStandard && o.level != interview.LevelFull {
		fmt.Fprintln(d.stderr, "jig: "+i18n.Tf(lang, "err.level.unknown", o.level))
		return o, lang, true, 2
	}

	rest := fset.Args()
	if len(rest) > 1 {
		fmt.Fprintln(d.stderr, "jig: "+i18n.T(lang, "help.usage"))
		return o, lang, true, 2
	}
	if len(rest) == 1 {
		o.name = rest[0]
		// Ask validates interview answers only; a name from the command line
		// is checked here, before anything else looks at it.
		if err := interview.ValidateName(o.name); err != nil {
			fmt.Fprintln(d.stderr, "jig: "+i18n.Tf(lang, "err.name.invalid", o.name))
			return o, lang, true, 2
		}
	}

	return o, lang, false, 0
}

// decide runs the interview: it asks for everything the flags left open.
func decide(o options, lang i18n.Lang, types []string, d deps) (interview.Decision, int) {
	in := interview.Input{
		Flags: interview.Flags{
			Level:       o.level,
			Name:        o.name,
			Description: o.description,
			HasDesc:     o.hasDesc,
			Type:        o.typ,
			Langs:       o.langs,
			Commit:      o.commit,
			HasCommit:   o.commit || o.noCommit,
		},
		Interactive: d.interactive,
		CWDName:     cwdName(d.workdir),
		Types:       types,
		TypeLabel: func(key string) string {
			t, err := manifest.Load(d.templates, typesRoot, key)
			if err != nil {
				return key
			}
			return t.Name.For(string(lang))
		},
		SlotsFor: func(key string) ([]interview.SlotQuestion, error) {
			t, err := manifest.Load(d.templates, typesRoot, key)
			if err != nil {
				return nil, err
			}
			return slotQuestions(t), nil
		},
		LangLabel: func(l string) string { return i18n.T(lang, "opt.lang."+l) },
	}

	// The empty-args branch is load-bearing: Tf on a key holding %s and no
	// arguments renders "%!s(MISSING)". Most keys are asked for as they are.
	dec, err := interview.Ask(in, d.stdin, d.stdout, func(k string, args ...any) string {
		if len(args) == 0 {
			return i18n.T(lang, k)
		}
		return i18n.Tf(lang, k, args...)
	})
	if err != nil {
		var missing *interview.MissingError
		if errors.As(err, &missing) {
			return interview.Decision{}, fail(d, 2, i18n.Tf(lang, "err.noninteractive", strings.Join(missing.Flags, ", ")))
		}
		var unknown *interview.UnknownSlotError
		if errors.As(err, &unknown) {
			return interview.Decision{}, fail(d, 2, i18n.Tf(lang, "err.slot.unknown", unknown.Slot))
		}
		return interview.Decision{}, fail(d, 1, err.Error())
	}
	if err := interview.ValidateName(dec.Name); err != nil {
		return interview.Decision{}, fail(d, 2, i18n.Tf(lang, "err.name.invalid", dec.Name))
	}
	return dec, 0
}

// resolveType loads the manifest of the chosen type and checks that every slot
// got a language the manifest knows. A standard run has no type: the zero Type
// with no slots is the right answer for it.
func resolveType(dec interview.Decision, lang i18n.Lang, types []string, d deps) (manifest.Type, int) {
	if dec.Level != interview.LevelFull {
		return manifest.Type{}, 0
	}
	if !slices.Contains(types, dec.Type) {
		return manifest.Type{}, fail(d, 2, i18n.Tf(lang, "err.type.unknown", dec.Type))
	}
	typ, err := manifest.Load(d.templates, typesRoot, dec.Type)
	if err != nil {
		return manifest.Type{}, fail(d, 1, i18n.Tf(lang, "err.manifest.invalid", dec.Type, err))
	}
	for _, s := range typ.Slots {
		l := dec.Langs[s.Dir]
		if l == "" {
			return manifest.Type{}, fail(d, 2, i18n.Tf(lang, "err.slot.nolang", s.Dir))
		}
		if _, ok := s.Langs[l]; !ok {
			return manifest.Type{}, fail(d, 2, i18n.Tf(lang, "err.lang.unknown", l))
		}
	}
	return typ, 0
}

// validateLangSlots rejects --lang keys that do not name a slot of typ.
// On standard, typ has no slots, so any --lang is an error. Sorted keys keep
// the error order stable when several slots are wrong at once.
func validateLangSlots(langs map[string]string, typ manifest.Type, lang i18n.Lang, d deps) int {
	if len(langs) == 0 {
		return 0
	}
	known := make(map[string]bool, len(typ.Slots))
	for _, s := range typ.Slots {
		known[s.Dir] = true
	}
	keys := make([]string, 0, len(langs))
	for slot := range langs {
		keys = append(keys, slot)
	}
	slices.Sort(keys)
	for _, slot := range keys {
		if !known[slot] {
			return fail(d, 2, i18n.Tf(lang, "err.slot.unknown", slot))
		}
	}
	return 0
}

// usesGo reports whether any chosen slot language is Go.
func usesGo(dec interview.Decision) bool {
	for _, l := range dec.Langs {
		if l == "go" {
			return true
		}
	}
	return false
}

// preflight checks everything that can be known before a single file is
// written: the directory is free, the tools are on PATH, git has an identity.
func preflight(d deps, lang i18n.Lang, dec interview.Decision, target string, slotCmds []cmd) int {
	entries, err := os.ReadDir(target)
	switch {
	case err == nil && len(entries) > 0:
		return fail(d, 1, i18n.Tf(lang, "err.dir.notempty", dec.Name))
	case err != nil && !errors.Is(err, fs.ErrNotExist):
		return fail(d, 1, err.Error())
	}

	tools := []string{"git"}
	for _, c := range slotCmds {
		if len(c.argv) > 0 && !slices.Contains(tools, c.argv[0]) {
			tools = append(tools, c.argv[0])
		}
	}
	if missing := toolchain.CheckPath(d.look, tools); len(missing) > 0 {
		for _, m := range missing {
			fmt.Fprintln(d.stderr, "jig: "+i18n.Tf(lang, "err.tool.missing", m))
		}
		return 1
	}

	// Only a requested commit needs an identity. The check runs against the
	// working directory: the target does not exist yet.
	if dec.Commit {
		if err := toolchain.GitIdentity(d.runner, d.workdir); err != nil {
			return fail(d, 1, i18n.T(lang, "err.git.identity"))
		}
	}
	return 0
}

// build creates the directory and carries out the plan. There is no rollback:
// on a failed initializer the directory stays as it is and jig says where it
// stopped.
func build(d deps, lang i18n.Lang, dec interview.Decision, typ manifest.Type, target string, data render.Data, slotCmds []cmd, skip []string) int {
	if err := os.MkdirAll(target, 0o755); err != nil {
		return fail(d, 1, err.Error())
	}
	files, err := render.Render(d.templates, coreRoot, target, data, skip)
	if err != nil {
		return fail(d, 1, err.Error())
	}

	if err := toolchain.GitInit(d.runner, target); err != nil {
		return fail(d, 1, err.Error())
	}
	ran := []string{"git init"}

	// The type's doc pack goes into docs/<id>/ before the slots run: it cannot
	// collide with a slot's cwd, so uv and npm never trip over it.
	if typ.Docs != "" {
		docsFiles, err := render.Render(d.templates, path.Join(typedocsRoot, typ.Docs),
			filepath.Join(target, "docs", typ.Docs), data, nil)
		if err != nil {
			return fail(d, 1, err.Error())
		}
		for _, f := range docsFiles {
			files = append(files, filepath.Join("docs", typ.Docs, f))
		}
	}

	for _, c := range slotCmds {
		dir := filepath.Join(target, c.dir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fail(d, 1, err.Error())
		}
		if err := d.runner.Run(dir, c.argv); err != nil {
			return fail(d, 1, i18n.Tf(lang, "err.init.failed", c.dir, err))
		}
		ran = append(ran, strings.Join(c.argv, " "))
		files = append(files, existingInitArtifacts(target, c.dir)...)

		// Language files land after the initializer: uv and npm refuse a
		// non-empty directory. The commit is still ahead, so they get into it.
		// RenderMerge, not Render: a pack must not overwrite what the core or
		// the initializer put there — its .gitignore block is appended.
		for _, pack := range c.files {
			written, err := render.RenderMerge(d.templates, path.Join(langsRoot, pack), dir, data)
			if err != nil {
				return fail(d, 1, err.Error())
			}
			for _, f := range written {
				files = append(files, filepath.Join(c.dir, f))
			}
		}
	}

	// The commit is last: otherwise what the initializers generated would not
	// be in it.
	if dec.Commit {
		if err := toolchain.GitCommit(d.runner, target, commitMsg); err != nil {
			return fail(d, 1, err.Error())
		}
		ran = append(ran, "git add -A")
	}

	// Compact after sort: an appended root .gitignore is reported once, not as
	// a core line plus a pack line.
	sorted := slices.Clone(files)
	slices.Sort(sorted)
	sorted = slices.Compact(sorted)
	for _, f := range sorted {
		fmt.Fprintln(d.stdout, i18n.Tf(lang, "report.created", filepath.Join(dec.Name, f)))
	}
	for _, r := range ran {
		fmt.Fprintln(d.stdout, i18n.Tf(lang, "report.ran", r))
	}
	if dec.Commit {
		fmt.Fprintln(d.stdout, i18n.Tf(lang, "report.commit", commitMsg))
	}
	fmt.Fprintln(d.stdout, i18n.Tf(lang, "report.done", dec.Name))
	return 0
}

// dryRun prints the files and the commands and writes nothing.
// Missing tools are warned on stderr but do not fail the plan: dry-run is for
// inspecting the plan, not for asserting the machine is ready.
func dryRun(d deps, lang i18n.Lang, dec interview.Decision, typ manifest.Type, slotCmds []cmd, skip []string) int {
	tools := []string{"git"}
	for _, c := range slotCmds {
		if len(c.argv) > 0 && !slices.Contains(tools, c.argv[0]) {
			tools = append(tools, c.argv[0])
		}
	}
	for _, m := range toolchain.CheckPath(d.look, tools) {
		fmt.Fprintln(d.stderr, "jig: "+i18n.Tf(lang, "warn.tool.missing", m))
	}

	files, err := render.Plan(d.templates, coreRoot, skip)
	if err != nil {
		return fail(d, 1, err.Error())
	}
	if typ.Docs != "" {
		planned, err := render.Plan(d.templates, path.Join(typedocsRoot, typ.Docs), nil)
		if err != nil {
			return fail(d, 1, err.Error())
		}
		for _, f := range planned {
			files = append(files, filepath.Join("docs", typ.Docs, f))
		}
	}
	for _, c := range slotCmds {
		for _, pack := range c.files {
			planned, err := render.Plan(d.templates, path.Join(langsRoot, pack), nil)
			if err != nil {
				return fail(d, 1, err.Error())
			}
			for _, f := range planned {
				files = append(files, filepath.Join(c.dir, f))
			}
		}
	}
	// Plan lists source-walk order, where a root "gitignore" surfaces late as
	// ".gitignore". Sorting turns it into a readable tree; Compact folds the
	// pack's append into the core's line, as the real run reports it.
	slices.Sort(files)
	files = slices.Compact(files)

	fmt.Fprintln(d.stdout, i18n.T(lang, "dryrun.header"))
	for _, f := range files {
		fmt.Fprintln(d.stdout, i18n.Tf(lang, "report.created", filepath.Join(dec.Name, f)))
	}
	fmt.Fprintln(d.stdout, i18n.Tf(lang, "report.ran", "git init"))
	for _, c := range slotCmds {
		fmt.Fprintln(d.stdout, i18n.Tf(lang, "report.ran", strings.Join(c.argv, " ")))
	}
	if dec.Commit {
		fmt.Fprintln(d.stdout, i18n.Tf(lang, "report.ran", "git add -A"))
		fmt.Fprintln(d.stdout, i18n.Tf(lang, "report.commit", commitMsg))
	}
	return 0
}

// existingInitArtifacts returns project-relative paths of known initializer
// outputs that are present under target/slotDir after a successful command.
func existingInitArtifacts(target, slotDir string) []string {
	var out []string
	for _, name := range knownInitArtifacts {
		abs := filepath.Join(target, slotDir, name)
		if _, err := os.Stat(abs); err != nil {
			continue
		}
		rel := name
		if slotDir != singleSlot && slotDir != "" {
			rel = filepath.Join(slotDir, name)
		}
		out = append(out, rel)
	}
	return out
}

// slotQuestions is the whole bridge between manifest and interview: the
// interview package is a leaf and must not know what a manifest is.
// Languages are offered default first, then the rest alphabetically. The user
// reads the list top down, and every other question of the interview puts its
// default at 1); a plain alphabetical order made "1)" and Enter disagree.
// manifest.validate guarantees DefaultLang is one of Langs.
func slotQuestions(t manifest.Type) []interview.SlotQuestion {
	qs := make([]interview.SlotQuestion, 0, len(t.Slots))
	for _, s := range t.Slots {
		rest := make([]string, 0, len(s.Langs))
		for l := range s.Langs {
			if l != s.DefaultLang {
				rest = append(rest, l)
			}
		}
		slices.Sort(rest)
		langs := append([]string{s.DefaultLang}, rest...)
		qs = append(qs, interview.SlotQuestion{Dir: s.Dir, Langs: langs, DefaultLang: s.DefaultLang})
	}
	return qs
}

// slotInfos is what the template gets to know about the slots of the project:
// only the named sub-slots. A standard run has none, and the single "." slot
// of cli/library is not a named subarea either — nil for both, so a template
// {{range .Slots}} renders exactly when the code has named parts.
func slotInfos(t manifest.Type, dec interview.Decision) []render.SlotInfo {
	var infos []render.SlotInfo
	for _, s := range t.Slots {
		if s.Dir == singleSlot {
			continue
		}
		infos = append(infos, render.SlotInfo{Dir: s.Dir, Lang: dec.Langs[s.Dir]})
	}
	return infos
}

// cwdName offers the current folder's name as the default project name — but
// only when it would pass the name check. Otherwise there is no default.
func cwdName(workdir string) string {
	base := filepath.Base(workdir)
	if interview.ValidateName(base) != nil {
		return ""
	}
	return base
}

// fail writes one prefixed line to stderr and hands back the exit code, so a
// caller can `return fail(...)`.
func fail(d deps, code int, msg string) int {
	fmt.Fprintf(d.stderr, "jig: %s\n", msg)
	return code
}

// printUsage writes the help. The help.flag.* strings are pre-formatted whole
// lines: they are printed as they are, or the columns break.
func printUsage(w io.Writer, lang i18n.Lang) {
	fmt.Fprintln(w, i18n.T(lang, "help.usage"))
	fmt.Fprintln(w)
	fmt.Fprintln(w, i18n.T(lang, "help.summary"))
	fmt.Fprintln(w)
	fmt.Fprintln(w, i18n.T(lang, "help.flags"))
	for _, key := range []string{
		"help.flag.level",
		"help.flag.description",
		"help.flag.type",
		"help.flag.lang",
		"help.flag.commit",
		"help.flag.nocommit",
		"help.flag.uilang",
		"help.flag.dryrun",
		"help.flag.version",
	} {
		fmt.Fprintln(w, i18n.T(lang, key))
	}
	fmt.Fprintln(w, "  -h, --help")
}
