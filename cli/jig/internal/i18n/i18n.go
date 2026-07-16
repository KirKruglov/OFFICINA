// Package i18n holds the bilingual interface strings of jig.
//
// The language is passed as a function parameter. There is no global state:
// the environment is read once in main and handed down.
package i18n

import (
	"fmt"
	"sort"
)

// Lang is an interface language.
type Lang string

// Supported languages. English is the fallback: the repository is public and
// most users do not have a Russian locale.
const (
	RU Lang = "ru"
	EN Lang = "en"
)

// pair holds one string in both languages.
type pair struct{ ru, en string }

// strings maps a key to its Russian and English text.
//
// Level keys (standard, full), type keys (cli, web, library) and language keys
// (go, python, ts, js) are never translated: they are flag values and manifest
// keys at the same time.
//
// The name shadows the stdlib strings package on purpose — the tests reach this
// map in-package. Nothing here imports strings.
var strings = map[string]pair{
	// Questions.
	"q.level": {
		ru: "Что создаём?",
		en: "What are we creating?",
	},
	"q.name": {
		ru: "Имя проекта?",
		en: "Project name?",
	},
	"q.description": {
		ru: "Описание проекта?",
		en: "Project description?",
	},
	"q.type": {
		ru: "Тип проекта?",
		en: "Project type?",
	},
	"q.lang.slot": {
		ru: "Язык для «%s»?",
		en: "Language for \"%s\"?",
	},
	"q.lang.single": {
		ru: "Язык проекта?",
		en: "Project language?",
	},
	"q.commit": {
		ru: "Сделать первый коммит?",
		en: "Make the first commit?",
	},

	// Options.
	"opt.level.standard": {
		ru: "Заготовку — папки, настройка под Claude, git",
		en: "A skeleton — folders, Claude setup, git",
	},
	"opt.level.full": {
		ru: "Готовый проект — то же плюс окружение языка",
		en: "A ready project — the same plus the language environment",
	},
	"opt.commit.yes": {
		ru: "Да",
		en: "Yes",
	},
	"opt.commit.no": {
		ru: "Нет",
		en: "No",
	},
	"opt.lang.go": {
		ru: "Go",
		en: "Go",
	},
	"opt.lang.python": {
		ru: "Python",
		en: "Python",
	},
	"opt.lang.ts": {
		ru: "TypeScript",
		en: "TypeScript",
	},
	"opt.lang.js": {
		ru: "JavaScript",
		en: "JavaScript",
	},
	"hint.default": {
		ru: "по умолчанию",
		en: "default",
	},
	"prompt.empty": {
		ru: "Выберите вариант.",
		en: "Choose an option.",
	},
	"prompt.invalid": {
		ru: "Такого варианта нет. Повторите ввод.",
		en: "No such option. Enter it again.",
	},

	// Errors.
	"err.name.invalid": {
		ru: "Имя «%s» недопустимо. Разрешены строчные латинские буквы, цифры и дефис; первый символ — буква.",
		en: "Name \"%s\" is not allowed. Use lowercase latin letters, digits and hyphen; the first character must be a letter.",
	},
	"err.name.go_keyword": {
		ru: "Имя «%s» нельзя использовать с Go: это зарезервированное слово, из него не получится имя пакета.",
		en: "Name \"%s\" cannot be used with Go: it is a reserved word and cannot be a package name.",
	},
	"err.dir.notempty": {
		ru: "Папка «%s» не пуста.",
		en: "Directory \"%s\" is not empty.",
	},
	"err.tool.missing": {
		ru: "Команда «%s» не найдена в PATH.",
		en: "Command \"%s\" not found in PATH.",
	},
	"warn.tool.missing": {
		ru: "Предупреждение: команда «%s» не найдена в PATH (реальный запуск упадёт).",
		en: "Warning: command \"%s\" not found in PATH (a real run would fail).",
	},
	"err.git.identity": {
		ru: "git не настроен. Задайте user.name и user.email.",
		en: "git is not configured. Set user.name and user.email.",
	},
	"err.type.unknown": {
		ru: "Неизвестный тип проекта: «%s».",
		en: "Unknown project type: \"%s\".",
	},
	"err.type.fullonly": {
		ru: "Флаг --type допустим только с --level full.",
		en: "--type is only valid with --level full.",
	},
	"err.lang.unknown": {
		ru: "Неизвестный язык: «%s».",
		en: "Unknown language: \"%s\".",
	},
	"err.level.unknown": {
		ru: "Недопустимое значение «%s» для --level: возможны standard или full.",
		en: "Invalid value \"%s\" for --level: use standard or full.",
	},
	"err.commit.conflict": {
		ru: "Флаги --commit и --no-commit несовместимы.",
		en: "--commit and --no-commit are mutually exclusive.",
	},
	"err.slot.nolang": {
		ru: "Для слота «%s» язык не задан.",
		en: "No language set for slot \"%s\".",
	},
	"err.slot.unknown": {
		ru: "Неизвестный слот «%s» для этого типа проекта.",
		en: "Unknown slot \"%s\" for this project type.",
	},
	"err.noninteractive": {
		ru: "Ввод не терминал, спросить нечем. Не хватает флагов: %s",
		en: "Input is not a terminal, nothing to ask with. Missing flags: %s",
	},
	"err.init.failed": {
		ru: "Инициализация слота «%s» остановлена: %s\nПапка оставлена как есть.",
		en: "Initialization of slot \"%s\" stopped: %s\nThe directory is left as is.",
	},
	"err.manifest.invalid": {
		ru: "Манифест типа «%s» повреждён: %s",
		en: "Manifest of type \"%s\" is broken: %s",
	},

	// Help.
	"help.usage": {
		ru: "Использование: jig [флаги] <name>",
		en: "Usage: jig [flags] <name>",
	},
	"help.summary": {
		ru: "jig разворачивает готовый к работе локальный репозиторий за один вызов.",
		en: "jig sets up a ready-to-work local repository in a single run.",
	},
	"help.flags": {
		ru: "Флаги:",
		en: "Flags:",
	},
	"help.flag.level": {
		ru: "  --level standard|full           что разворачиваем",
		en: "  --level standard|full           what to set up",
	},
	"help.flag.description": {
		ru: "  --description <текст>           одна строка о проекте",
		en: "  --description <text>            one line about the project",
	},
	"help.flag.type": {
		ru: "  --type cli|web|library          только для full",
		en: "  --type cli|web|library          full only",
	},
	"help.flag.lang": {
		ru: "  --lang <slot>=<lang>            повторяемый: --lang backend=go --lang frontend=js",
		en: "  --lang <slot>=<lang>            repeatable: --lang backend=go --lang frontend=js",
	},
	"help.flag.commit": {
		ru: "  --commit                        сделать первый коммит",
		en: "  --commit                        make the first commit",
	},
	"help.flag.nocommit": {
		ru: "  --no-commit                     обойтись без коммита",
		en: "  --no-commit                     skip the first commit",
	},
	"help.flag.uilang": {
		ru: "  --ui-lang ru|en                 язык интерфейса; перекрывает окружение",
		en: "  --ui-lang ru|en                 interface language; overrides the environment",
	},
	"help.flag.dryrun": {
		ru: "  --dry-run                       показать план, ничего не писать",
		en: "  --dry-run                       show the plan, write nothing",
	},
	"help.flag.version": {
		ru: "  --version                       версия и выход",
		en: "  --version                       print the version and exit",
	},

	// Report.
	"report.created": {
		ru: "Создано: %s",
		en: "Created: %s",
	},
	"report.ran": {
		ru: "Выполнено: %s",
		en: "Ran: %s",
	},
	"report.commit": {
		ru: "Коммит: %s",
		en: "Commit: %s",
	},
	"report.done": {
		ru: "Проект «%s» готов.",
		en: "Project \"%s\" is ready.",
	},
	"dryrun.header": {
		ru: "Пробный запуск. Ничего не записано.",
		en: "Dry run. Nothing written.",
	},
}

// Detect resolves the interface language: flagVal wins, then lcAll, then
// langEnv. A value starting with "ru" means Russian, anything else — English.
// An empty or unknown value means English.
func Detect(flagVal, lcAll, langEnv string) Lang {
	for _, v := range []string{flagVal, lcAll, langEnv} {
		if v == "" {
			continue
		}
		if hasRuPrefix(v) {
			return RU
		}
		return EN
	}
	return EN
}

// hasRuPrefix reports whether v starts with "ru". Written by hand because the
// package-level strings map shadows the stdlib strings package.
func hasRuPrefix(v string) bool {
	return len(v) >= 2 && v[0] == 'r' && v[1] == 'u'
}

// T returns the string for key in lang. An unknown key yields "!"+key, an
// unknown language falls back to English. T never panics.
func T(lang Lang, key string) string {
	p, ok := strings[key]
	if !ok {
		return "!" + key
	}
	if lang == RU {
		return p.ru
	}
	return p.en
}

// Tf returns T(lang, key) formatted with args.
func Tf(lang Lang, key string, args ...any) string {
	return fmt.Sprintf(T(lang, key), args...)
}

// Keys returns every key, sorted.
func Keys() []string {
	keys := make([]string, 0, len(strings))
	for k := range strings {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
