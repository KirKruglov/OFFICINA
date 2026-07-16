// Package interview turns flags plus answers into a resolved decision.
//
// The package is a leaf: it knows nothing about manifests, i18n or the
// filesystem. Everything it needs arrives through Input — slots as
// SlotQuestion values, strings through a T func. main.go does the wiring.
package interview

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// SlotQuestion is interview's narrow view of a manifest slot.
type SlotQuestion struct {
	Dir         string   // "." means single-slot: ask "Language of the project?"
	Langs       []string // offered in order: DefaultLang first, the rest alphabetical
	DefaultLang string
}

// Flags is everything the command line already answered. Zero value = unset.
type Flags struct {
	Level       string // "standard" | "full" | ""
	Name        string
	Description string
	HasDesc     bool // distinguishes --description "" from unset
	Type        string
	Langs       map[string]string // slot dir → lang; single-slot form stores under "."
	Commit      bool
	HasCommit   bool
}

// Decision is the fully resolved answer set.
type Decision struct {
	Level       string
	Name        string
	Description string
	Type        string            // "" when Level == "standard"
	Langs       map[string]string // slot dir → lang
	Commit      bool
}

// T translates a key, formatting args into it when the key carries verbs.
// main.go supplies a wrapper over i18n.Tf that binds the language.
type T func(key string, args ...any) string

// Input carries the flags and everything the interview may need to ask about.
type Input struct {
	Flags       Flags
	Interactive bool
	CWDName     string // default for the name question; "" if it fails ValidateName
	Types       []string
	TypeLabel   func(typeKey string) string
	SlotsFor    func(typeKey string) ([]SlotQuestion, error)
	LangLabel   func(lang string) string
}

// Level values. They are flag values and manifest keys, never translated.
const (
	LevelStandard = "standard"
	LevelFull     = "full"
)

// singleSlotDir marks the slot of a type that has exactly one. Its name is
// never shown: such a type asks "Language of the project?".
const singleSlotDir = "."

// MissingError lists the flags that would have been asked for.
type MissingError struct{ Flags []string }

func (e *MissingError) Error() string {
	return "missing data for a non-interactive run: " + strings.Join(e.Flags, ", ")
}

// UnknownSlotError is a --lang key that is not a slot of the chosen type.
// Raised before a MissingError so "backend" on a cli type is not reported as
// a mere missing bare --lang.
type UnknownSlotError struct{ Slot string }

func (e *UnknownSlotError) Error() string {
	return "unknown slot: " + e.Slot
}

var nameRe = regexp.MustCompile(`^[a-z]([a-z0-9]*(-[a-z0-9]+)*)?$`)

// ValidateName enforces: lowercase latin, digits, hyphen; first char a letter;
// a hyphen only between two alphanumerics — no trailing or doubled hyphens.
// The name becomes both a directory and a module path, so this one rule also
// blocks path traversal, spaces, case and dotted names.
func ValidateName(name string) error {
	if !nameRe.MatchString(name) {
		return fmt.Errorf("invalid project name %q: lowercase latin letters, digits and hyphen only, first character a letter", name)
	}
	return nil
}

// Ask resolves the decision, asking only for what the flags left open.
// A non-interactive run never prompts: missing data returns a *MissingError.
func Ask(in Input, r io.Reader, w io.Writer, t T) (Decision, error) {
	if !in.Interactive {
		return resolveNonInteractive(in)
	}
	s := &session{sc: bufio.NewScanner(r), w: w, t: t}
	return s.run(in)
}

// resolveNonInteractive collects every flag the interview would have asked
// for and refuses in one go, instead of failing one question at a time.
func resolveNonInteractive(in Input) (Decision, error) {
	f := in.Flags
	var missing []string
	if f.Level == "" {
		missing = append(missing, "--level")
	}
	if f.Name == "" {
		missing = append(missing, "<name>")
	}
	if !f.HasDesc {
		missing = append(missing, "--description")
	}

	d := Decision{
		Level:       f.Level,
		Name:        f.Name,
		Description: f.Description,
		Langs:       map[string]string{},
		Commit:      f.Commit,
	}

	if f.Level == LevelFull {
		if f.Type == "" {
			missing = append(missing, "--type")
		} else {
			d.Type = f.Type
			slots, err := in.SlotsFor(f.Type)
			if err != nil {
				return Decision{}, err
			}
			if err := rejectUnknownLangSlots(f.Langs, slots); err != nil {
				return Decision{}, err
			}
			langMissing := false
			for _, slot := range slots {
				lang, ok := f.Langs[slot.Dir]
				if !ok || lang == "" {
					langMissing = true
					continue
				}
				d.Langs[slot.Dir] = lang
			}
			if langMissing {
				missing = append(missing, "--lang")
			}
		}
	}

	if !f.HasCommit {
		missing = append(missing, "--commit")
	}
	if len(missing) > 0 {
		return Decision{}, &MissingError{Flags: missing}
	}
	return d, nil
}

// session holds the reader, the writer and the translator for one interview.
type session struct {
	sc *bufio.Scanner
	w  io.Writer
	t  T
}

func (s *session) run(in Input) (Decision, error) {
	f := in.Flags
	d := Decision{
		Level:       f.Level,
		Name:        f.Name,
		Description: f.Description,
		Langs:       map[string]string{},
		Commit:      f.Commit,
	}

	// 1. What are we creating?
	if d.Level == "" {
		lvl, err := s.askChoice(s.t("q.level"), []option{
			{value: LevelStandard, label: s.t("opt.level.standard")},
			{value: LevelFull, label: s.t("opt.level.full")},
		}, 0)
		if err != nil {
			return Decision{}, err
		}
		d.Level = lvl
	}

	// 2. Project name.
	if d.Name == "" {
		name, err := s.askLine(s.t("q.name"), in.CWDName, false, ValidateName)
		if err != nil {
			return Decision{}, err
		}
		d.Name = name
	}

	// 3. Description — asked at both levels, empty answer allowed.
	if !f.HasDesc {
		desc, err := s.askLine(s.t("q.description"), "", true, nil)
		if err != nil {
			return Decision{}, err
		}
		d.Description = desc
	}

	if d.Level == LevelFull {
		// 4. Project type.
		if d.Type = f.Type; d.Type == "" {
			opts := make([]option, 0, len(in.Types))
			for _, key := range in.Types {
				opts = append(opts, option{value: key, label: in.TypeLabel(key)})
			}
			typ, err := s.askChoice(s.t("q.type"), opts, -1)
			if err != nil {
				return Decision{}, err
			}
			d.Type = typ
		}

		// 5..N. One question per slot of the chosen type.
		slots, err := in.SlotsFor(d.Type)
		if err != nil {
			return Decision{}, err
		}
		if err := rejectUnknownLangSlots(f.Langs, slots); err != nil {
			return Decision{}, err
		}
		for _, slot := range slots {
			if lang, ok := f.Langs[slot.Dir]; ok && lang != "" {
				d.Langs[slot.Dir] = lang
				continue
			}
			lang, err := s.askSlot(in, slot)
			if err != nil {
				return Decision{}, err
			}
			d.Langs[slot.Dir] = lang
		}
	}

	// 7. First commit.
	if !f.HasCommit {
		answer, err := s.askChoice(s.t("q.commit"), []option{
			{value: "yes", label: s.t("opt.commit.yes")},
			{value: "no", label: s.t("opt.commit.no")},
		}, 0)
		if err != nil {
			return Decision{}, err
		}
		d.Commit = answer == "yes"
	}

	return d, nil
}

// rejectUnknownLangSlots fails on --lang keys that are not among the type's
// slots. Keys are checked in sorted order so the first error is stable.
func rejectUnknownLangSlots(langs map[string]string, slots []SlotQuestion) error {
	if len(langs) == 0 {
		return nil
	}
	known := make(map[string]bool, len(slots))
	for _, s := range slots {
		known[s.Dir] = true
	}
	keys := make([]string, 0, len(langs))
	for slot := range langs {
		keys = append(keys, slot)
	}
	slices.Sort(keys)
	for _, slot := range keys {
		if !known[slot] {
			return &UnknownSlotError{Slot: slot}
		}
	}
	return nil
}

// askSlot asks one slot's language. A slot with a single language is still
// asked: language packs grow, and a "show or not" branch would rot.
func (s *session) askSlot(in Input, slot SlotQuestion) (string, error) {
	question := s.t("q.lang.single")
	if slot.Dir != singleSlotDir {
		question = s.t("q.lang.slot", slot.Dir)
	}
	opts := make([]option, 0, len(slot.Langs))
	def := -1
	for i, lang := range slot.Langs {
		opts = append(opts, option{value: lang, label: in.LangLabel(lang)})
		if lang == slot.DefaultLang {
			def = i
		}
	}
	return s.askChoice(question, opts, def)
}

type option struct {
	value string
	label string
}

// askChoice prints numbered options and reads a number. defaultIdx < 0 means
// there is no default: an empty answer re-prompts.
func (s *session) askChoice(question string, opts []option, defaultIdx int) (string, error) {
	for {
		fmt.Fprintln(s.w, question)
		for i, o := range opts {
			line := fmt.Sprintf("  %d) %s", i+1, o.label)
			if i == defaultIdx {
				line += " (" + s.t("hint.default") + ")"
			}
			fmt.Fprintln(s.w, line)
		}
		answer, err := s.read()
		if err != nil {
			return "", err
		}
		if answer == "" {
			if defaultIdx >= 0 {
				return opts[defaultIdx].value, nil
			}
			// Nothing was offered to default to, so the answer is not a wrong
			// option — it is a missing one.
			fmt.Fprintln(s.w, s.t("prompt.empty"))
			continue
		}
		n, err := strconv.Atoi(answer)
		if err == nil && n >= 1 && n <= len(opts) {
			return opts[n-1].value, nil
		}
		fmt.Fprintln(s.w, s.t("prompt.invalid"))
	}
}

// askLine reads a single line. An empty answer takes def when there is one;
// allowEmpty accepts an empty answer as a value in its own right.
func (s *session) askLine(question, def string, allowEmpty bool, validate func(string) error) (string, error) {
	for {
		fmt.Fprintln(s.w, question)
		if def != "" {
			fmt.Fprintf(s.w, "  %s (%s)\n", def, s.t("hint.default"))
		}
		answer, err := s.read()
		if err != nil {
			return "", err
		}
		if answer == "" {
			if def != "" {
				return def, nil
			}
			if allowEmpty {
				return "", nil
			}
			fmt.Fprintln(s.w, s.t("prompt.empty"))
			continue
		}
		if validate != nil {
			if err := validate(answer); err != nil {
				// err is untranslated; the key states the rule instead.
				fmt.Fprintln(s.w, s.t("err.name.invalid", answer))
				continue
			}
		}
		return answer, nil
	}
}

// read prints the prompt and returns one trimmed line.
func (s *session) read() (string, error) {
	fmt.Fprint(s.w, "> ")
	if !s.sc.Scan() {
		if err := s.sc.Err(); err != nil {
			return "", err
		}
		return "", io.ErrUnexpectedEOF
	}
	return strings.TrimSpace(s.sc.Text()), nil
}
