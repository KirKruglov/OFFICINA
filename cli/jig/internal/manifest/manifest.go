// Package manifest reads and validates project type manifests.
//
// A manifest is a JSON file describing one project type: its display name and
// the slots (directories) it is made of. Each slot lists the initializer
// command for every language it supports. A command is an array of arguments,
// never a shell string: no shell, no escaping, no injection.
package manifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
	"text/template"
)

// Label is a display string in both supported interface languages.
type Label struct {
	RU string `json:"ru"`
	EN string `json:"en"`
}

// For returns the label in lang ("ru" or "en"). Any other value yields EN.
func (l Label) For(lang string) string {
	if lang == "ru" {
		return l.RU
	}
	return l.EN
}

// Slot is one directory of a project type together with the initializer
// command for each language it supports.
type Slot struct {
	Dir         string              `json:"dir"`
	DefaultLang string              `json:"default_lang"`
	Langs       map[string][]string `json:"langs"`
	// Files is optional: language key → folders under templates/langs whose
	// contents are rendered into the slot after the initializer succeeds. It
	// exists because go mod init has no flags to express a type, while uv and
	// vite do (--lib/--app, vanilla-ts/vanilla).
	Files map[string][]string `json:"files,omitempty"`
}

// Type is a parsed project type manifest.
type Type struct {
	Name Label `json:"name"`
	// Order is optional: List and the interview sort types by it, then by
	// name. Types without it (zero) come first, alphabetically.
	Order int `json:"order,omitempty"`
	// Docs is optional: the id of a doc pack under templates/typedocs whose
	// contents are rendered into the project's docs/ directory. Empty means
	// the type ships no docs of its own.
	Docs  string `json:"docs,omitempty"`
	Slots []Slot `json:"slots"`
}

// Parse decodes and validates a single manifest.
//
// Unknown JSON keys are an error rather than silently ignored. Note that
// DisallowUnknownFields constrains structs only, so the langs map still
// accepts arbitrary language keys — that is by design.
func Parse(data []byte) (Type, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var typ Type
	if err := dec.Decode(&typ); err != nil {
		return Type{}, fmt.Errorf("decode manifest: %w", err)
	}
	if err := typ.validate(); err != nil {
		return Type{}, fmt.Errorf("invalid manifest: %w", err)
	}
	return typ, nil
}

// validate enforces the rules the spec mandates for a manifest.
func (t Type) validate() error {
	if t.Name.RU == "" {
		return errors.New(`field "name" is missing key "ru"`)
	}
	if t.Name.EN == "" {
		return errors.New(`field "name" is missing key "en"`)
	}
	if len(t.Slots) == 0 {
		return errors.New(`field "slots" must contain at least one slot`)
	}
	for i, s := range t.Slots {
		if err := s.validate(); err != nil {
			return fmt.Errorf("%s: %w", s.label(i), err)
		}
	}
	return nil
}

// label identifies a slot in error messages by index, plus its dir when known.
func (s Slot) label(i int) string {
	if s.Dir == "" {
		return fmt.Sprintf("slot %d", i)
	}
	return fmt.Sprintf("slot %d (%s)", i, s.Dir)
}

func (s Slot) validate() error {
	if s.Dir == "" {
		return errors.New(`field "dir" is required`)
	}
	if s.DefaultLang == "" {
		return errors.New(`field "default_lang" is required`)
	}
	if len(s.Langs) == 0 {
		return errors.New(`field "langs" must contain at least one language`)
	}
	if _, ok := s.Langs[s.DefaultLang]; !ok {
		return fmt.Errorf(`default_lang %q is absent from "langs" (have: %s)`,
			s.DefaultLang, strings.Join(s.languages(), ", "))
	}
	// A slot cannot ship files for a language it does not support. The same
	// rule as default_lang: a typo must not be ignored into a wrong result.
	for _, lang := range s.fileLangs() {
		if _, ok := s.Langs[lang]; !ok {
			return fmt.Errorf(`files key %q is absent from "langs" (have: %s)`,
				lang, strings.Join(s.languages(), ", "))
		}
	}
	return nil
}

// fileLangs returns the language keys of Files, sorted, so that a manifest
// with two bad keys always fails on the same one.
func (s Slot) fileLangs() []string {
	langs := make([]string, 0, len(s.Files))
	for lang := range s.Files {
		langs = append(langs, lang)
	}
	slices.Sort(langs)
	return langs
}

// FilesFor returns the template folders lang ships, or nil when it ships none.
// The result is a copy: the parsed manifest is reused across calls.
func (s Slot) FilesFor(lang string) []string {
	return slices.Clone(s.Files[lang])
}

// languages returns the slot's language keys, sorted.
func (s Slot) languages() []string {
	langs := make([]string, 0, len(s.Langs))
	for lang := range s.Langs {
		langs = append(langs, lang)
	}
	slices.Sort(langs)
	return langs
}

// Command renders the initializer argv for lang, substituting template
// placeholders such as {{.Name}} from data into each argument separately.
// An unknown lang is an error.
func (s Slot) Command(lang string, data any) ([]string, error) {
	argv, ok := s.Langs[lang]
	if !ok {
		return nil, fmt.Errorf("slot %s: unknown language %q (have: %s)",
			s.Dir, lang, strings.Join(s.languages(), ", "))
	}

	// Render into a fresh slice: the manifest's argv is a template, reused
	// across calls, and must not be overwritten with rendered values.
	rendered := make([]string, len(argv))
	for i, arg := range argv {
		tmpl, err := template.New("arg").Option("missingkey=error").Parse(arg)
		if err != nil {
			return nil, fmt.Errorf("slot %s: language %s: argument %d: parse template: %w",
				s.Dir, lang, i, err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return nil, fmt.Errorf("slot %s: language %s: argument %d: render template: %w",
				s.Dir, lang, i, err)
		}
		rendered[i] = buf.String()
	}
	return rendered, nil
}

// Load reads root/typeName.json from fsys and parses it.
func Load(fsys fs.FS, root, typeName string) (Type, error) {
	name := path.Join(root, typeName+".json")
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return Type{}, fmt.Errorf("read manifest for type %q: %w", typeName, err)
	}
	typ, err := Parse(data)
	if err != nil {
		return Type{}, fmt.Errorf("type %q: %w", typeName, err)
	}
	return typ, nil
}

// List returns the available type keys — the basenames of the *.json files
// directly under root — sorted by the manifests' order field, then by name.
// It loads every manifest to read the order, so a broken one fails the list
// rather than surfacing later at a random point.
func List(fsys fs.FS, root string) ([]string, error) {
	entries, err := fs.ReadDir(fsys, root)
	if err != nil {
		return nil, fmt.Errorf("list types in %q: %w", root, err)
	}
	type keyed struct {
		name  string
		order int
	}
	types := make([]keyed, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || path.Ext(e.Name()) != ".json" {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".json")
		typ, err := Load(fsys, root, name)
		if err != nil {
			return nil, err
		}
		types = append(types, keyed{name: name, order: typ.Order})
	}
	slices.SortFunc(types, func(a, b keyed) int {
		if a.order != b.order {
			return a.order - b.order
		}
		return strings.Compare(a.name, b.name)
	})
	names := make([]string, len(types))
	for i, t := range types {
		names[i] = t.name
	}
	return names, nil
}
