package i18n

import (
	str "strings"
	"testing"
)

func TestDetect(t *testing.T) {
	cases := []struct {
		name    string
		flagVal string
		lcAll   string
		langEnv string
		want    Lang
	}{
		{"flag overrides env to ru", "ru", "en_US.UTF-8", "en_US.UTF-8", RU},
		{"flag overrides env to en", "en", "ru_RU.UTF-8", "", EN},
		{"lc_all beats lang", "", "ru_RU.UTF-8", "en_US.UTF-8", RU},
		{"lang used when lc_all empty", "", "", "ru_RU.UTF-8", RU},
		{"english lang", "", "", "en_US.UTF-8", EN},
		{"everything empty", "", "", "", EN},
		{"c locale", "", "", "C.UTF-8", EN},
		{"unknown locale", "", "", "de_DE.UTF-8", EN},
		{"bare ru", "", "", "ru", RU},
		{"unknown flag value falls back to en", "de", "ru_RU.UTF-8", "", EN},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Detect(c.flagVal, c.lcAll, c.langEnv)
			if got != c.want {
				t.Errorf("Detect(%q, %q, %q) = %q, want %q", c.flagVal, c.lcAll, c.langEnv, got, c.want)
			}
		})
	}
}

func TestKeysSortedAndComplete(t *testing.T) {
	keys := Keys()
	const want = 51
	if len(keys) != want {
		t.Errorf("Keys() length = %d, want %d", len(keys), want)
	}
	if len(keys) != len(strings) {
		t.Errorf("Keys() length = %d, strings map length = %d", len(keys), len(strings))
	}
	for i := 1; i < len(keys); i++ {
		if keys[i-1] >= keys[i] {
			t.Errorf("Keys() not sorted: %q >= %q", keys[i-1], keys[i])
		}
	}
	for _, k := range keys {
		if _, ok := strings[k]; !ok {
			t.Errorf("Keys() returned %q which is absent from the strings map", k)
		}
	}
}

func TestEveryKeyHasBothLanguages(t *testing.T) {
	for _, k := range Keys() {
		p := strings[k]
		if p.ru == "" {
			t.Errorf("key %q: ru is empty", k)
		}
		if p.en == "" {
			t.Errorf("key %q: en is empty", k)
		}
	}
}

func TestExpectedKeySet(t *testing.T) {
	want := []string{
		// questions
		"q.level", "q.name", "q.description", "q.type",
		"q.lang.slot", "q.lang.single", "q.commit",
		// options
		"opt.level.standard", "opt.level.full",
		"opt.commit.yes", "opt.commit.no",
		"opt.lang.go", "opt.lang.python", "opt.lang.ts", "opt.lang.js",
		"hint.default", "prompt.empty", "prompt.invalid",
		// errors
		"err.name.invalid", "err.name.go_keyword",
		"err.dir.notempty",
		"err.tool.missing", "warn.tool.missing", "err.git.identity",
		"err.type.unknown", "err.type.fullonly",
		"err.lang.unknown", "err.slot.nolang", "err.slot.unknown",
		"err.noninteractive", "err.init.failed", "err.manifest.invalid",
		"err.level.unknown", "err.commit.conflict",
		// help
		"help.usage", "help.summary", "help.flags",
		"help.flag.level", "help.flag.description", "help.flag.type",
		"help.flag.lang", "help.flag.commit", "help.flag.nocommit",
		"help.flag.uilang", "help.flag.dryrun", "help.flag.version",
		// report
		"report.created", "report.ran", "report.commit", "report.done",
		"dryrun.header",
	}
	for _, k := range want {
		if _, ok := strings[k]; !ok {
			t.Errorf("missing key %q", k)
		}
	}
	if len(strings) != len(want) {
		t.Errorf("strings map has %d keys, expected set has %d", len(strings), len(want))
	}
}

func TestFormatVerbsMatchAcrossLanguages(t *testing.T) {
	for _, k := range Keys() {
		p := strings[k]
		if got, want := str.Count(p.ru, "%s"), str.Count(p.en, "%s"); got != want {
			t.Errorf("key %q: ru has %d %%s, en has %d %%s", k, got, want)
		}
	}
}

func TestLevelOptionsHideLevelKeys(t *testing.T) {
	for _, k := range []string{"opt.level.standard", "opt.level.full"} {
		p := strings[k]
		for _, text := range []string{p.ru, p.en} {
			if str.Contains(text, "standard") || str.Contains(text, "full") {
				t.Errorf("key %q: option text %q must not expose the level key", k, text)
			}
		}
	}
}

func TestT(t *testing.T) {
	if got := T(RU, "q.commit"); got != "Сделать первый коммит?" {
		t.Errorf("T(RU, q.commit) = %q", got)
	}
	if got := T(EN, "q.commit"); got != "Make the first commit?" {
		t.Errorf("T(EN, q.commit) = %q", got)
	}
}

func TestTUnknownKey(t *testing.T) {
	if got := T(RU, "unknown.key"); got != "!unknown.key" {
		t.Errorf("T(RU, unknown.key) = %q, want %q", got, "!unknown.key")
	}
	if got := T(EN, "unknown.key"); got != "!unknown.key" {
		t.Errorf("T(EN, unknown.key) = %q, want %q", got, "!unknown.key")
	}
}

func TestTUnknownLangFallsBackToEnglish(t *testing.T) {
	if got, want := T(Lang("de"), "q.commit"), T(EN, "q.commit"); got != want {
		t.Errorf("T(de, q.commit) = %q, want %q", got, want)
	}
}

func TestTf(t *testing.T) {
	got := Tf(RU, "q.lang.slot", "backend")
	if got != "Язык для «backend»?" {
		t.Errorf("Tf(RU, q.lang.slot, backend) = %q", got)
	}
	if got := Tf(EN, "q.lang.slot", "backend"); !str.Contains(got, "backend") {
		t.Errorf("Tf(EN, q.lang.slot, backend) = %q, want it to contain the slot", got)
	}
}

func TestTfUnknownKey(t *testing.T) {
	if got := Tf(RU, "unknown.key", "x"); !str.HasPrefix(got, "!unknown.key") {
		t.Errorf("Tf(RU, unknown.key) = %q, want prefix %q", got, "!unknown.key")
	}
}

func TestSingleSlotQuestionOmitsSlotName(t *testing.T) {
	for _, l := range []Lang{RU, EN} {
		if got := T(l, "q.lang.single"); str.Contains(got, "%s") || str.Contains(got, ".") && str.Contains(got, "«.»") {
			t.Errorf("T(%s, q.lang.single) = %q must not reference the slot name", l, got)
		}
	}
}
