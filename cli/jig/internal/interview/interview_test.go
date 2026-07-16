package interview

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// fakeT renders keys without dots so tests can assert that a bare "." slot
// name never leaks into the output. Args are appended, mirroring a real
// catalogue string that ends in its verbs.
func fakeT(key string, args ...any) string {
	s := "T[" + strings.ReplaceAll(key, ".", "_") + "]"
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s+strings.Repeat(" %v", len(args)), args...)
}

func fakeTypeLabel(typeKey string) string { return "TYPE_" + typeKey }
func fakeLangLabel(lang string) string    { return "LANG_" + lang }

func fakeSlotsFor(typeKey string) ([]SlotQuestion, error) {
	switch typeKey {
	case "cli", "library":
		return []SlotQuestion{
			{Dir: ".", Langs: []string{"go", "python"}, DefaultLang: "go"},
		}, nil
	case "web":
		return []SlotQuestion{
			{Dir: "backend", Langs: []string{"go", "python"}, DefaultLang: "python"},
			{Dir: "frontend", Langs: []string{"js", "ts"}, DefaultLang: "ts"},
		}, nil
	}
	return nil, fmt.Errorf("unknown type %q", typeKey)
}

func baseInput() Input {
	return Input{
		Interactive: true,
		CWDName:     "cwd-name",
		Types:       []string{"cli", "web", "library"},
		TypeLabel:   fakeTypeLabel,
		SlotsFor:    fakeSlotsFor,
		LangLabel:   fakeLangLabel,
	}
}

func run(t *testing.T, in Input, stdin string) (Decision, string, error) {
	t.Helper()
	var w bytes.Buffer
	d, err := Ask(in, strings.NewReader(stdin), &w, fakeT)
	return d, w.String(), err
}

func TestValidateName(t *testing.T) {
	cases := []struct {
		name string
		ok   bool
	}{
		{"my-app", true},
		{"a", true},
		{"app2", true},
		{"my-app-2", true},
		{"a-b-c", true},
		{"my-", false},
		{"app-", false},
		{"my--app", false},
		{"myApp", false},
		{"my app", false},
		{"../evil", false},
		{"/etc", false},
		{".hidden", false},
		{"1st", false},
		{"", false},
		{"My-App", false},
		{"my_app", false},
		{"-app", false},
		{"app.name", false},
		{"app/name", false},
	}
	for _, c := range cases {
		err := ValidateName(c.name)
		if c.ok && err != nil {
			t.Errorf("ValidateName(%q) = %v, want nil", c.name, err)
		}
		if !c.ok && err == nil {
			t.Errorf("ValidateName(%q) = nil, want error", c.name)
		}
	}
}

func TestFlagSetSkipsQuestion(t *testing.T) {
	in := baseInput()
	in.Flags = Flags{
		Level:     "standard",
		Name:      "flag-app",
		Commit:    true,
		HasCommit: true,
		HasDesc:   false,
	}
	d, out, err := run(t, in, "some description\n")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	for _, key := range []string{"q.level", "q.name", "q.commit"} {
		if strings.Contains(out, fakeT(key)) {
			t.Errorf("output contains %q, want question skipped\n%s", key, out)
		}
	}
	if !strings.Contains(out, fakeT("q.description")) {
		t.Errorf("q.description not asked\n%s", out)
	}
	if d.Level != "standard" || d.Name != "flag-app" || d.Description != "some description" {
		t.Errorf("decision = %+v", d)
	}
	if !d.Commit {
		t.Errorf("Commit = false, want true from flag")
	}
}

func TestEmptyDescriptionFlagCountsAsAnswered(t *testing.T) {
	in := baseInput()
	in.Flags = Flags{
		Level:       "standard",
		Name:        "flag-app",
		Description: "",
		HasDesc:     true,
		Commit:      false,
		HasCommit:   true,
	}
	d, out, err := run(t, in, "")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if strings.Contains(out, fakeT("q.description")) {
		t.Errorf("q.description asked despite --description \"\"\n%s", out)
	}
	if d.Description != "" {
		t.Errorf("Description = %q, want empty", d.Description)
	}
}

func TestStandardSkipsTypeAndLangButAsksCommit(t *testing.T) {
	in := baseInput()
	in.Flags = Flags{Name: "app", Description: "", HasDesc: true}
	// level → 1 (standard), commit → empty (default yes)
	d, out, err := run(t, in, "1\n\n")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	for _, key := range []string{"q.type", "q.lang.single", "q.lang.slot"} {
		if strings.Contains(out, fakeT(key)) {
			t.Errorf("output contains %q on standard level\n%s", key, out)
		}
	}
	if !strings.Contains(out, fakeT("q.commit")) {
		t.Errorf("q.commit not asked on standard level\n%s", out)
	}
	if d.Level != "standard" {
		t.Errorf("Level = %q, want standard", d.Level)
	}
	if d.Type != "" {
		t.Errorf("Type = %q, want empty on standard", d.Type)
	}
	if len(d.Langs) != 0 {
		t.Errorf("Langs = %v, want empty on standard", d.Langs)
	}
	if !d.Commit {
		t.Errorf("Commit = false, want default yes")
	}
}

func TestSingleSlotAsksSingleLangQuestion(t *testing.T) {
	in := baseInput()
	in.Flags = Flags{Level: "full", Name: "app", Description: "desc", HasDesc: true, Type: "cli", HasCommit: true, Commit: true}
	// lang → 2 (python)
	d, out, err := run(t, in, "2\n")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if !strings.Contains(out, fakeT("q.lang.single")) {
		t.Errorf("q.lang.single not asked\n%s", out)
	}
	if strings.Contains(out, fakeT("q.lang.slot")) {
		t.Errorf("q.lang.slot asked for single-slot type\n%s", out)
	}
	if strings.Contains(out, ".") {
		t.Errorf("output leaks a bare %q slot name\n%s", ".", out)
	}
	if d.Langs["."] != "python" {
		t.Errorf("Langs = %v, want {.: python}", d.Langs)
	}
}

func TestSingleSlotWithOneLanguageIsStillAsked(t *testing.T) {
	in := baseInput()
	in.SlotsFor = func(string) ([]SlotQuestion, error) {
		return []SlotQuestion{{Dir: ".", Langs: []string{"go"}, DefaultLang: "go"}}, nil
	}
	in.Flags = Flags{Level: "full", Name: "app", HasDesc: true, Type: "library", HasCommit: true}
	d, out, err := run(t, in, "1\n")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if !strings.Contains(out, fakeT("q.lang.single")) {
		t.Errorf("slot with a single language was not asked\n%s", out)
	}
	if d.Langs["."] != "go" {
		t.Errorf("Langs = %v", d.Langs)
	}
}

func TestWebAsksTwoSlotQuestionsInManifestOrder(t *testing.T) {
	in := baseInput()
	in.Flags = Flags{Level: "full", Name: "app", HasDesc: true, Type: "web", HasCommit: true}
	// backend → 1 (go), frontend → 1 (js)
	d, out, err := run(t, in, "1\n1\n")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	backend := strings.Index(out, fakeT("q.lang.slot", "backend"))
	frontend := strings.Index(out, fakeT("q.lang.slot", "frontend"))
	if backend < 0 || frontend < 0 {
		t.Fatalf("slot questions missing: backend=%d frontend=%d\n%s", backend, frontend, out)
	}
	if backend > frontend {
		t.Errorf("frontend asked before backend\n%s", out)
	}
	if d.Langs["backend"] != "go" || d.Langs["frontend"] != "js" {
		t.Errorf("Langs = %v, want {backend: go, frontend: js}", d.Langs)
	}
}

func TestPerSlotLangFlagSkipsOnlyThatSlot(t *testing.T) {
	in := baseInput()
	in.Flags = Flags{
		Level: "full", Name: "app", HasDesc: true, Type: "web", HasCommit: true,
		Langs: map[string]string{"backend": "go"},
	}
	// only frontend is asked → 2 (ts)
	d, out, err := run(t, in, "2\n")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if strings.Contains(out, fakeT("q.lang.slot", "backend")) {
		t.Errorf("backend asked despite --lang backend=go\n%s", out)
	}
	if !strings.Contains(out, fakeT("q.lang.slot", "frontend")) {
		t.Errorf("frontend not asked\n%s", out)
	}
	if d.Langs["backend"] != "go" || d.Langs["frontend"] != "ts" {
		t.Errorf("Langs = %v", d.Langs)
	}
}

func TestEmptyAnswerTakesDefault(t *testing.T) {
	in := baseInput()
	in.Flags = Flags{Level: "full", Type: "web", HasDesc: true}
	// name → empty (cwd-name), backend → empty (python), frontend → empty (ts), commit → empty (yes)
	d, _, err := run(t, in, "\n\n\n\n")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if d.Name != "cwd-name" {
		t.Errorf("Name = %q, want cwd-name", d.Name)
	}
	if d.Langs["backend"] != "python" {
		t.Errorf("backend = %q, want python", d.Langs["backend"])
	}
	if d.Langs["frontend"] != "ts" {
		t.Errorf("frontend = %q, want ts", d.Langs["frontend"])
	}
	if !d.Commit {
		t.Errorf("Commit = false, want default yes")
	}
}

func TestEmptyAnswerTakesLevelDefault(t *testing.T) {
	in := baseInput()
	in.Flags = Flags{Name: "app", HasDesc: true, HasCommit: true}
	d, _, err := run(t, in, "\n")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if d.Level != "standard" {
		t.Errorf("Level = %q, want standard by default", d.Level)
	}
}

func TestInvalidChoiceReprompts(t *testing.T) {
	in := baseInput()
	in.Flags = Flags{Name: "app", HasDesc: true, HasCommit: true}
	// level: 9 → invalid, then 2 → full; type: 2 → web; langs: 1, 1
	d, out, err := run(t, in, "9\n2\n2\n1\n1\n")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if !strings.Contains(out, fakeT("prompt.invalid")) {
		t.Errorf("prompt.invalid not printed\n%s", out)
	}
	if strings.Count(out, fakeT("q.level")) < 2 {
		t.Errorf("q.level not re-asked\n%s", out)
	}
	if d.Level != "full" || d.Type != "web" {
		t.Errorf("decision = %+v, want full/web", d)
	}
}

func TestInvalidChoiceKindsReprompt(t *testing.T) {
	in := baseInput()
	in.Flags = Flags{Name: "app", HasDesc: true, HasCommit: true}
	// "abc", "0", "-1" all invalid, then 1 → standard
	d, out, err := run(t, in, "abc\n0\n-1\n1\n")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if strings.Count(out, fakeT("prompt.invalid")) != 3 {
		t.Errorf("want 3 invalid prompts, got %d\n%s", strings.Count(out, fakeT("prompt.invalid")), out)
	}
	if d.Level != "standard" {
		t.Errorf("Level = %q", d.Level)
	}
}

func TestEmptyChoiceWithoutDefaultRepromptsAsEmpty(t *testing.T) {
	in := baseInput()
	in.Flags = Flags{Level: "full", Name: "app", HasDesc: true, HasCommit: true}
	// type has no default: empty → prompt.empty, then 2 → web; langs: 1, 1
	d, out, err := run(t, in, "\n2\n1\n1\n")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if !strings.Contains(out, fakeT("prompt.empty")) {
		t.Errorf("prompt.empty not printed for an empty answer without a default\n%s", out)
	}
	if strings.Contains(out, fakeT("prompt.invalid")) {
		t.Errorf("empty answer reported as an invalid option\n%s", out)
	}
	if d.Type != "web" {
		t.Errorf("Type = %q, want web", d.Type)
	}
}

func TestInvalidNameReprompts(t *testing.T) {
	in := baseInput()
	in.CWDName = ""
	in.Flags = Flags{Level: "standard", HasDesc: true, HasCommit: true}
	d, out, err := run(t, in, "My App\n../evil\nmy-app\n")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if n := strings.Count(out, fakeT("err.name.invalid")); n != 2 {
		t.Errorf("want 2 name-rule messages, got %d\n%s", n, out)
	}
	// The rejected name itself must reach the user, not just the rule.
	for _, bad := range []string{"My App", "../evil"} {
		if !strings.Contains(out, fakeT("err.name.invalid", bad)) {
			t.Errorf("message for %q does not name it\n%s", bad, out)
		}
	}
	if d.Name != "my-app" {
		t.Errorf("Name = %q, want my-app", d.Name)
	}
}

func TestNoCWDNameMeansEmptyAnswerReprompts(t *testing.T) {
	in := baseInput()
	in.CWDName = ""
	in.Flags = Flags{Level: "standard", HasDesc: true, HasCommit: true}
	d, out, err := run(t, in, "\nmy-app\n")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if !strings.Contains(out, fakeT("prompt.empty")) {
		t.Errorf("empty answer accepted without a default\n%s", out)
	}
	if strings.Contains(out, fakeT("hint.default")) {
		t.Errorf("default hint shown without a default\n%s", out)
	}
	if d.Name != "my-app" {
		t.Errorf("Name = %q", d.Name)
	}
}

func TestNameDefaultsToCWDName(t *testing.T) {
	in := baseInput()
	in.CWDName = "cwd-name"
	in.Flags = Flags{Level: "standard", HasDesc: true, HasCommit: true}
	d, out, err := run(t, in, "\n")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if !strings.Contains(out, "cwd-name") {
		t.Errorf("CWDName not offered as default\n%s", out)
	}
	if d.Name != "cwd-name" {
		t.Errorf("Name = %q, want cwd-name", d.Name)
	}
}

func TestDescriptionAcceptsEmptyAnswer(t *testing.T) {
	in := baseInput()
	in.Flags = Flags{Level: "standard", Name: "app", HasCommit: true}
	d, _, err := run(t, in, "\n")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if d.Description != "" {
		t.Errorf("Description = %q, want empty", d.Description)
	}
}

func TestDescriptionAskedOnBothLevels(t *testing.T) {
	for _, level := range []string{"standard", "full"} {
		in := baseInput()
		in.Flags = Flags{Level: level, Name: "app", Type: "cli", HasCommit: true,
			Langs: map[string]string{".": "go"}}
		_, out, err := run(t, in, "a description\n")
		if err != nil {
			t.Fatalf("level %s: Ask: %v", level, err)
		}
		if !strings.Contains(out, fakeT("q.description")) {
			t.Errorf("level %s: q.description not asked\n%s", level, out)
		}
	}
}

func TestNonInteractiveMissingData(t *testing.T) {
	cases := []struct {
		name  string
		flags Flags
		want  []string
	}{
		{
			name:  "nothing given",
			flags: Flags{},
			want:  []string{"--level", "<name>", "--description", "--commit"},
		},
		{
			name:  "full without type",
			flags: Flags{Level: "full", Name: "app", HasDesc: true, HasCommit: true},
			want:  []string{"--type"},
		},
		{
			name:  "full with type, no langs",
			flags: Flags{Level: "full", Name: "app", HasDesc: true, HasCommit: true, Type: "web"},
			want:  []string{"--lang"},
		},
		{
			name: "full with type, one lang missing",
			flags: Flags{Level: "full", Name: "app", HasDesc: true, HasCommit: true, Type: "web",
				Langs: map[string]string{"backend": "go"}},
			want: []string{"--lang"},
		},
		{
			name:  "standard needs no type or lang",
			flags: Flags{Level: "standard", HasDesc: true, HasCommit: true},
			want:  []string{"<name>"},
		},
		{
			name:  "only commit missing",
			flags: Flags{Level: "standard", Name: "app", HasDesc: true},
			want:  []string{"--commit"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := baseInput()
			in.Interactive = false
			in.Flags = c.flags
			_, out, err := run(t, in, "")
			var me *MissingError
			if !errors.As(err, &me) {
				t.Fatalf("err = %v, want *MissingError", err)
			}
			if strings.Join(me.Flags, ",") != strings.Join(c.want, ",") {
				t.Errorf("missing = %v, want %v", me.Flags, c.want)
			}
			if out != "" {
				t.Errorf("wrote %q to writer, want nothing", out)
			}
			if me.Error() == "" {
				t.Errorf("Error() is empty")
			}
		})
	}
}

func TestNonInteractiveUnknownSlot(t *testing.T) {
	in := baseInput()
	in.Interactive = false
	in.Flags = Flags{
		Level: "full", Name: "app", Description: "desc", HasDesc: true,
		Type: "cli", Langs: map[string]string{"backend": "go"},
		Commit: true, HasCommit: true,
	}
	_, _, err := run(t, in, "")
	var ue *UnknownSlotError
	if !errors.As(err, &ue) {
		t.Fatalf("err = %v, want *UnknownSlotError", err)
	}
	if ue.Slot != "backend" {
		t.Errorf("slot = %q, want backend", ue.Slot)
	}
	var me *MissingError
	if errors.As(err, &me) {
		t.Error("unknown slot must not also be a MissingError")
	}
}

func TestNonInteractiveAllFlagsGiven(t *testing.T) {
	in := baseInput()
	in.Interactive = false
	in.Flags = Flags{
		Level: "full", Name: "app", Description: "desc", HasDesc: true,
		Type: "web", Langs: map[string]string{"backend": "go", "frontend": "js"},
		Commit: true, HasCommit: true,
	}
	d, out, err := run(t, in, "")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if out != "" {
		t.Errorf("wrote %q to writer, want nothing", out)
	}
	want := Decision{
		Level: "full", Name: "app", Description: "desc", Type: "web",
		Langs: map[string]string{"backend": "go", "frontend": "js"}, Commit: true,
	}
	if d.Level != want.Level || d.Name != want.Name || d.Description != want.Description ||
		d.Type != want.Type || d.Commit != want.Commit ||
		d.Langs["backend"] != "go" || d.Langs["frontend"] != "js" {
		t.Errorf("decision = %+v, want %+v", d, want)
	}
}

func TestNonInteractiveSingleSlotLangUnderDot(t *testing.T) {
	in := baseInput()
	in.Interactive = false
	in.Flags = Flags{
		Level: "full", Name: "app", HasDesc: true, Type: "cli",
		Langs: map[string]string{".": "python"}, HasCommit: true, Commit: false,
	}
	d, _, err := run(t, in, "")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if d.Langs["."] != "python" {
		t.Errorf("Langs = %v, want {.: python}", d.Langs)
	}
	if d.Commit {
		t.Errorf("Commit = true, want false from --no-commit")
	}
}

func TestAllFlagsGivenWritesNothing(t *testing.T) {
	in := baseInput()
	in.Flags = Flags{
		Level: "full", Name: "app", Description: "desc", HasDesc: true,
		Type: "cli", Langs: map[string]string{".": "go"}, Commit: true, HasCommit: true,
	}
	d, out, err := run(t, in, "")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if out != "" {
		t.Errorf("Ask wrote %q, want nothing", out)
	}
	if d.Type != "cli" || d.Langs["."] != "go" {
		t.Errorf("decision = %+v", d)
	}
}

func TestUnknownTypeFromSlotsFor(t *testing.T) {
	in := baseInput()
	in.Interactive = false
	in.Flags = Flags{Level: "full", Name: "app", HasDesc: true, Type: "desktop", HasCommit: true}
	_, _, err := run(t, in, "")
	if err == nil {
		t.Fatalf("want error for unknown type")
	}
	var me *MissingError
	if errors.As(err, &me) {
		t.Errorf("got *MissingError, want the SlotsFor error")
	}
}

func TestEOFDuringInterview(t *testing.T) {
	in := baseInput()
	in.Flags = Flags{Level: "standard", Name: "app", HasDesc: true}
	_, _, err := run(t, in, "")
	if err == nil {
		t.Fatalf("want error on EOF")
	}
}
