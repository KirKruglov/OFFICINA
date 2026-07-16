// Package render walks a template tree and materialises it into a directory.
//
// The package knows no file and no folder name inside the tree it renders. The
// whole contract is:
//
//   - a file with the .tmpl suffix is executed as a Go template and loses the
//     suffix; a file without it is copied byte for byte;
//   - dot-names are stored dot-less because go:embed skips dotted paths, and
//     the walk restores them: "gitignore" and "gitkeep", at any level, become
//     ".gitignore" and ".gitkeep"; the directory "claude" at the tree root
//     becomes ".claude"; nothing else is renamed;
//   - empty directories are not carried over — only directories holding a file
//     are created.
package render

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

// SlotInfo describes one slot of the chosen project type: the directory it
// creates and the language it holds. Templates reach for .Dir and .Lang.
type SlotInfo struct {
	Dir  string
	Lang string
}

// Data is the complete set of template variables. Because it is a struct
// (not a map), text/template makes any other field reference an Execute
// error for free.
//
// Empty Description and Author are legal: "Author: {{.Author}}" renders as
// "Author: " — ugly, not broken. Slots carries only the named sub-slots (web:
// backend, frontend); the "standard" level and single-"."-slot types leave it
// nil, and a {{range .Slots}} then renders nothing.
type Data struct {
	Name        string
	PackageName string
	Description string
	Date        string
	Author      string
	Slots       []SlotInfo
}

// PackageName turns a project name into a Go package name by dropping the
// hyphens: "probe-lib" becomes "probelib". Dropping them is enough because the
// name has already passed interview.ValidateName (lowercase latin, digits,
// hyphens only between alphanumerics, leading letter), so the only character
// left that a Go identifier rejects is the hyphen. No import of interview
// here: the rule is stated, not borrowed, to keep the packages apart.
func PackageName(name string) string {
	return strings.ReplaceAll(name, "-", "")
}

// goKeywords is the set of Go reserved words that cannot be a package clause.
// Kept as a map so IsGoKeyword is O(1) and the list stays in one place.
var goKeywords = map[string]struct{}{
	"break": {}, "case": {}, "chan": {}, "const": {}, "continue": {},
	"default": {}, "defer": {}, "else": {}, "fallthrough": {}, "for": {},
	"func": {}, "go": {}, "goto": {}, "if": {}, "import": {},
	"interface": {}, "map": {}, "package": {}, "range": {}, "return": {},
	"select": {}, "struct": {}, "switch": {}, "type": {}, "var": {},
}

// IsGoKeyword reports whether s is a Go reserved word. Callers pass
// PackageName(projectName): the hyphen-stripped form is what lands in
// "package …" for library-go.
func IsGoKeyword(s string) bool {
	_, ok := goKeywords[s]
	return ok
}

// Render walks root in fsys and writes into dir, creating dir as needed.
// skip lists root-relative source paths excluded together with their
// subtrees; nil skips nothing. It returns the destination-relative paths it
// created, in fs.WalkDir order.
func Render(fsys fs.FS, root, dir string, data Data, skip []string) ([]string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	var created []string
	err := walk(fsys, root, skip, func(src, target string) error {
		created = append(created, target)

		content, err := fileContent(fsys, src, data)
		if err != nil {
			return err
		}

		dst := filepath.Join(dir, filepath.FromSlash(target))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dst, content, 0o644)
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// RenderMerge renders a language pack into dir over files that may already be
// there — the core tree or an initializer ran first. An existing ".gitignore"
// target gets the pack block appended after one blank line; any other existing
// target is left untouched; fresh files are written as by Render. The returned
// paths are those written or appended — a skipped file is not reported.
func RenderMerge(fsys fs.FS, root, dir string, data Data) ([]string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	var written []string
	err := walk(fsys, root, nil, func(src, target string) error {
		content, err := fileContent(fsys, src, data)
		if err != nil {
			return err
		}

		dst := filepath.Join(dir, filepath.FromSlash(target))
		existing, err := os.ReadFile(dst)
		switch {
		case err != nil && !os.IsNotExist(err):
			return err
		case err == nil && path.Base(target) != ".gitignore":
			return nil // never overwrite what the initializer created
		case err == nil:
			// Append the ignore block after one blank line, normalising the
			// missing final newline of the existing file.
			joined := bytes.TrimRight(existing, "\n")
			content = append(append(joined, "\n\n"...), content...)
		default:
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				return err
			}
		}

		written = append(written, target)
		return os.WriteFile(dst, content, 0o644)
	})
	if err != nil {
		return nil, err
	}
	return written, nil
}

// fileContent reads one template file and, for a .tmpl source, executes it.
func fileContent(fsys fs.FS, src string, data Data) ([]byte, error) {
	content, err := fs.ReadFile(fsys, src)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(src, ".tmpl") {
		return execute(src, content, data)
	}
	return content, nil
}

// Plan returns exactly what Render would create, writing nothing. It takes
// the same skip list, so a dry-run plan and a real run never diverge.
func Plan(fsys fs.FS, root string, skip []string) ([]string, error) {
	var planned []string
	err := walk(fsys, root, skip, func(_, target string) error {
		planned = append(planned, target)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return planned, nil
}

// walk visits every file under root, calling fn with the source path in fsys
// and the destination-relative target path. Directories are skipped, so empty
// ones never reach fn — that is how rule 6 (no empty dirs) holds. Files whose
// root-relative path falls under an entry of skip are left out.
func walk(fsys fs.FS, root string, skip []string, fn func(src, target string) error) error {
	return fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := relTo(root, p)
		if err != nil {
			return err
		}
		if skipped(rel, skip) {
			return nil
		}
		return fn(p, targetName(rel))
	})
}

// skipped reports whether rel equals an entry of skip or lies inside one.
// Matching is by path segment: "docs/product/release-0.1.0-notes" does not
// fall under "docs/product/release-0.1.0".
func skipped(rel string, skip []string) bool {
	for _, s := range skip {
		if rel == s || strings.HasPrefix(rel, s+"/") {
			return true
		}
	}
	return false
}

// relTo returns p relative to root, both being slash-separated fs.FS paths.
func relTo(root, p string) (string, error) {
	root = path.Clean(root)
	if root == "." {
		return p, nil
	}
	rel := strings.TrimPrefix(p, root+"/")
	if rel == p {
		return "", fmt.Errorf("render: %q is not under %q", p, root)
	}
	return rel, nil
}

// execute renders content as a Go template. A reference to any field Data does
// not have fails here, which is the whole point of Data being a struct.
func execute(name string, content []byte, data Data) ([]byte, error) {
	tmpl, err := template.New(path.Base(name)).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("render: parse %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render: execute %s: %w", name, err)
	}
	return buf.Bytes(), nil
}

// targetName maps a template's relative path to its output path: it strips the
// .tmpl suffix and restores the dots that go:embed cannot carry — "gitignore"
// and "gitkeep" at any level, the "claude" directory at the tree root.
func targetName(rel string) string {
	target := strings.TrimSuffix(rel, ".tmpl")
	if target == "claude" || strings.HasPrefix(target, "claude/") {
		target = "." + target
	}
	dir, base := path.Split(target)
	switch base {
	case "gitignore":
		base = ".gitignore"
	case "gitkeep":
		base = ".gitkeep"
	}
	return dir + base
}
