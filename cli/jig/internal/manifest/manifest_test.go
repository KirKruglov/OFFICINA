package manifest

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

const validJSON = `{
  "name": { "ru": "Веб-приложение", "en": "Web application" },
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
}`

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"types/web.json":     {Data: []byte(validJSON)},
		"types/cli.json":     {Data: []byte(validJSON)},
		"types/library.json": {Data: []byte(validJSON)},
	}
}

func TestParseValid(t *testing.T) {
	typ, err := Parse([]byte(validJSON))
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if typ.Name.RU != "Веб-приложение" {
		t.Errorf("Name.RU = %q, want %q", typ.Name.RU, "Веб-приложение")
	}
	if typ.Name.EN != "Web application" {
		t.Errorf("Name.EN = %q, want %q", typ.Name.EN, "Web application")
	}
	if len(typ.Slots) != 2 {
		t.Fatalf("len(Slots) = %d, want 2", len(typ.Slots))
	}
	// Slots must preserve manifest order: backend before frontend.
	if typ.Slots[0].Dir != "backend" {
		t.Errorf("Slots[0].Dir = %q, want %q", typ.Slots[0].Dir, "backend")
	}
	if typ.Slots[1].Dir != "frontend" {
		t.Errorf("Slots[1].Dir = %q, want %q", typ.Slots[1].Dir, "frontend")
	}
	if typ.Slots[0].DefaultLang != "python" {
		t.Errorf("Slots[0].DefaultLang = %q, want %q", typ.Slots[0].DefaultLang, "python")
	}
	if got := typ.Slots[0].Langs["python"]; !slices.Equal(got, []string{"uv", "init"}) {
		t.Errorf("Slots[0].Langs[python] = %v, want [uv init]", got)
	}
}

func TestParseInvalid(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{
			name: "unknown key",
			json: `{"name":{"ru":"А","en":"A"},"version":2,
			        "slots":[{"dir":"x","default_lang":"go","langs":{"go":["go"]}}]}`,
		},
		{
			name: "unknown key in slot",
			json: `{"name":{"ru":"А","en":"A"},
			        "slots":[{"dir":"x","default_lang":"go","langs":{"go":["go"]},"extra":1}]}`,
		},
		{
			name: "name missing ru",
			json: `{"name":{"en":"A"},
			        "slots":[{"dir":"x","default_lang":"go","langs":{"go":["go"]}}]}`,
		},
		{
			name: "name missing en",
			json: `{"name":{"ru":"А"},
			        "slots":[{"dir":"x","default_lang":"go","langs":{"go":["go"]}}]}`,
		},
		{
			name: "name absent entirely",
			json: `{"slots":[{"dir":"x","default_lang":"go","langs":{"go":["go"]}}]}`,
		},
		{
			name: "slot without dir",
			json: `{"name":{"ru":"А","en":"A"},
			        "slots":[{"default_lang":"go","langs":{"go":["go"]}}]}`,
		},
		{
			name: "default_lang empty",
			json: `{"name":{"ru":"А","en":"A"},
			        "slots":[{"dir":"x","default_lang":"","langs":{"go":["go"]}}]}`,
		},
		{
			name: "default_lang absent",
			json: `{"name":{"ru":"А","en":"A"},
			        "slots":[{"dir":"x","langs":{"go":["go"]}}]}`,
		},
		{
			name: "default_lang not in langs",
			json: `{"name":{"ru":"А","en":"A"},
			        "slots":[{"dir":"x","default_lang":"rust","langs":{"go":["go"]}}]}`,
		},
		{
			name: "files key not in langs",
			json: `{"name":{"ru":"А","en":"A"},
			        "slots":[{"dir":"x","default_lang":"go","langs":{"go":["go"]},
			                  "files":{"rust":["rust-cli"]}}]}`,
		},
		{
			name: "empty slots",
			json: `{"name":{"ru":"А","en":"A"},"slots":[]}`,
		},
		{
			name: "slots absent",
			json: `{"name":{"ru":"А","en":"A"}}`,
		},
		{
			name: "malformed json",
			json: `{"name":`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Parse([]byte(tt.json)); err == nil {
				t.Fatalf("Parse(%s) error = nil, want error", tt.name)
			}
		})
	}
}

const filesJSON = `{
  "name": { "ru": "Инструмент", "en": "Tool" },
  "slots": [
    {
      "dir": ".",
      "default_lang": "go",
      "langs": {
        "go": ["go", "mod", "init", "{{.Name}}"],
        "python": ["uv", "init", "--app"]
      },
      "files": { "go": ["cli-go"] }
    }
  ]
}`

func TestParseFiles(t *testing.T) {
	typ, err := Parse([]byte(filesJSON))
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	slot := typ.Slots[0]

	if got := slot.FilesFor("go"); !slices.Equal(got, []string{"cli-go"}) {
		t.Errorf("FilesFor(go) = %v, want [cli-go]", got)
	}
	// A language that ships no files is the norm, not an error.
	if got := slot.FilesFor("python"); got != nil {
		t.Errorf("FilesFor(python) = %v, want nil", got)
	}
	if got := slot.FilesFor("rust"); got != nil {
		t.Errorf("FilesFor(rust) = %v, want nil", got)
	}
}

func TestParseDocsField(t *testing.T) {
	withDocs := strings.Replace(validJSON, `"name":`, `"docs": "design",
  "name":`, 1)
	typ, err := Parse([]byte(withDocs))
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if typ.Docs != "design" {
		t.Errorf("Docs = %q, want %q", typ.Docs, "design")
	}
}

// A type without "docs" ships no doc pack: the field is optional.
func TestParseWithoutDocsIsValid(t *testing.T) {
	typ, err := Parse([]byte(validJSON))
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if typ.Docs != "" {
		t.Errorf("Docs = %q, want empty", typ.Docs)
	}
}

// A slot without "files" is the common case: the field is optional.
func TestParseWithoutFilesIsValid(t *testing.T) {
	typ, err := Parse([]byte(validJSON))
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if got := typ.Slots[0].FilesFor("go"); got != nil {
		t.Errorf("FilesFor(go) = %v, want nil", got)
	}
}

// The manifest is parsed once and reused, so a caller must not be able to
// mutate it through the returned slice.
func TestFilesForDoesNotMutateManifest(t *testing.T) {
	typ, err := Parse([]byte(filesJSON))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	slot := typ.Slots[0]

	got := slot.FilesFor("go")
	got[0] = "overwritten"

	if again := slot.FilesFor("go"); !slices.Equal(again, []string{"cli-go"}) {
		t.Errorf("FilesFor(go) = %v after mutation, want [cli-go]", again)
	}
}

func TestSlotCommand(t *testing.T) {
	typ, err := Parse([]byte(validJSON))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	slot := typ.Slots[0]
	data := struct{ Name string }{"demo"}

	got, err := slot.Command("go", data)
	if err != nil {
		t.Fatalf("Command(go) error = %v, want nil", err)
	}
	if want := []string{"go", "mod", "init", "demo"}; !slices.Equal(got, want) {
		t.Errorf("Command(go) = %v, want %v", got, want)
	}

	got, err = slot.Command("python", data)
	if err != nil {
		t.Fatalf("Command(python) error = %v, want nil", err)
	}
	if want := []string{"uv", "init"}; !slices.Equal(got, want) {
		t.Errorf("Command(python) = %v, want %v", got, want)
	}
}

func TestSlotCommandUnknownLang(t *testing.T) {
	typ, err := Parse([]byte(validJSON))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if _, err := typ.Slots[0].Command("nope", struct{ Name string }{"demo"}); err == nil {
		t.Fatal("Command(nope) error = nil, want error")
	}
}

func TestSlotCommandDoesNotMutateManifest(t *testing.T) {
	typ, err := Parse([]byte(validJSON))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	slot := typ.Slots[0]
	if _, err := slot.Command("go", struct{ Name string }{"demo"}); err != nil {
		t.Fatalf("Command(go) error = %v", err)
	}
	// Rendering must not write through to the stored argv template.
	want := []string{"go", "mod", "init", "{{.Name}}"}
	if got := slot.Langs["go"]; !slices.Equal(got, want) {
		t.Errorf("Langs[go] after render = %v, want %v", got, want)
	}
}

func TestSlotCommandBadTemplate(t *testing.T) {
	slot := Slot{
		Dir:         "x",
		DefaultLang: "go",
		Langs:       map[string][]string{"go": {"go", "{{.Nope}}"}},
	}
	if _, err := slot.Command("go", struct{ Name string }{"demo"}); err == nil {
		t.Fatal("Command with unknown field error = nil, want error")
	}
}

func TestList(t *testing.T) {
	got, err := List(testFS(), "types")
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	want := []string{"cli", "library", "web"}
	if !slices.Equal(got, want) {
		t.Errorf("List() = %v, want %v", got, want)
	}
}

func TestListIgnoresNonJSON(t *testing.T) {
	fsys := testFS()
	fsys["types/README.md"] = &fstest.MapFile{Data: []byte("# types")}
	fsys["types/notes.txt"] = &fstest.MapFile{Data: []byte("scratch")}
	fsys["types/jsonish"] = &fstest.MapFile{Data: []byte("{}")}

	got, err := List(fsys, "types")
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	want := []string{"cli", "library", "web"}
	if !slices.Equal(got, want) {
		t.Errorf("List() = %v, want %v", got, want)
	}
}

func TestListSortsByOrderThenName(t *testing.T) {
	withOrder := func(n int) []byte {
		return []byte(strings.Replace(validJSON, `"name":`,
			fmt.Sprintf(`"order": %d, "name":`, n), 1))
	}
	fsys := fstest.MapFS{
		"types/cli.json":     {Data: withOrder(1)},
		"types/web.json":     {Data: withOrder(2)},
		"types/library.json": {Data: withOrder(3)},
		// No order → 0, ahead of the numbered ones; ties break by name.
		"types/beta.json":  {Data: []byte(validJSON)},
		"types/alpha.json": {Data: []byte(validJSON)},
	}

	got, err := List(fsys, "types")
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	want := []string{"alpha", "beta", "cli", "web", "library"}
	if !slices.Equal(got, want) {
		t.Errorf("List() = %v, want %v", got, want)
	}
}

func TestListInvalidManifestFails(t *testing.T) {
	fsys := testFS()
	fsys["types/broken.json"] = &fstest.MapFile{Data: []byte(`{"slots":[]}`)}

	if _, err := List(fsys, "types"); err == nil {
		t.Fatal("List with a broken manifest: error = nil, want error")
	}
}

func TestListMissingRoot(t *testing.T) {
	if _, err := List(testFS(), "nope"); err == nil {
		t.Fatal("List(nope) error = nil, want error")
	}
}

func TestLoad(t *testing.T) {
	typ, err := Load(testFS(), "types", "web")
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if typ.Name.EN != "Web application" {
		t.Errorf("Name.EN = %q, want %q", typ.Name.EN, "Web application")
	}
	if len(typ.Slots) != 2 {
		t.Errorf("len(Slots) = %d, want 2", len(typ.Slots))
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(testFS(), "types", "nope")
	if err == nil {
		t.Fatal("Load(nope) error = nil, want error")
	}
	// The error must name the type, otherwise the CLI cannot report it usefully.
	if !strings.Contains(err.Error(), "nope") {
		t.Errorf("Load() error = %q, want it to mention %q", err, "nope")
	}
}

func TestLoadInvalidManifest(t *testing.T) {
	fsys := fstest.MapFS{
		"types/broken.json": {Data: []byte(`{"name":{"ru":"А","en":"A"},"slots":[]}`)},
	}
	if _, err := Load(fsys, "types", "broken"); err == nil {
		t.Fatal("Load(broken) error = nil, want error")
	}
}

func TestLabelFor(t *testing.T) {
	l := Label{RU: "Веб-приложение", EN: "Web application"}
	tests := []struct {
		lang string
		want string
	}{
		{"ru", "Веб-приложение"},
		{"en", "Web application"},
		{"xx", "Web application"},
		{"", "Web application"},
	}
	for _, tt := range tests {
		if got := l.For(tt.lang); got != tt.want {
			t.Errorf("For(%q) = %q, want %q", tt.lang, got, tt.want)
		}
	}
}
