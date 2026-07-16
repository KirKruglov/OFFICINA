<p align="center">
  <a href="README.md">English</a> ·
  <strong>Deutsch</strong> ·
  <a href="README.es.md">Español</a> ·
  <a href="README.fr.md">Français</a> ·
  <a href="README.ru.md">Русский</a>
</p>

<h1 align="center">OFFICINA</h1>

<p align="center">
  <em>lateinisch</em> <strong>officina</strong> — eine Werkstatt, eine Schmiede, in der Werkzeuge entstehen.
</p>

<p align="center">
  <img alt="Plattform: macOS" src="https://img.shields.io/badge/platform-macOS-000000?logo=apple&logoColor=white">
  <img alt="Kern: Claude Code" src="https://img.shields.io/badge/core-Claude%20Code-da7756">
  <img alt="Lizenz: MIT" src="https://img.shields.io/badge/license-MIT-green">
  <a href="https://github.com/KirKruglov/OFFICINA/commits/main"><img alt="Letzter Commit" src="https://img.shields.io/github/last-commit/KirKruglov/OFFICINA?logo=github&label=letzter%20Commit"></a>
</p>

<p align="center">
  <strong>Methode und Werkzeuge für die systematische Arbeit mit einem KI-Entwicklungsassistenten.</strong>
  Skills, Subagents, CLI und Methodik — zusammengestellt und in echten Projekten erprobt.
</p>

---

> <sub>Übersetzt aus der maßgeblichen <a href="README.md">englischen README</a>; kann dem Original hinterherhinken.</sub>

## Warum OFFICINA

OFFICINA ist eine Werkstatt für die Arbeit mit KI-Werkzeugen in der Produktentwicklung. Im Kern steht
die **Methode**: wie man die Arbeit mit einem KI-Assistenten so aufbaut, dass sie vorhersehbar und
reproduzierbar bleibt. Darum herum — Skills, Subagents, CLI und Konfigurationen, bereit zur
Wiederverwendung.

- **Methode zuerst.** Wiederverwendbare Regeln und Leitfäden — Schreibstil, CLI-Erstellung, Skill- &
  Subagent-Erstellung, Loop-Engineering — jeweils für sich lesbar.
- **Portable Skills, Claude Code im Kern.** Skills sind nicht an ein Werkzeug gebunden — sie laufen in
  Claude Code und kompatiblen Umgebungen. Subagents und Einstellungen sind auf Claude Code
  zugeschnitten. Beides ist handverlesen statt in Masse übernommen.
- **Umgebung um den Kern.** VS Code als angewandter Bereich, der den Kern stützt, statt das
  Aushängeschild zu sein.

## Für wen

- **Sie arbeiten mit einem KI-Assistenten und wollen System** — eine Methode, keine einmaligen Prompts.
- **Sie schreiben eigene Skills, Subagents, CLI** — greifen Sie auf erprobte Muster und Leitfäden zurück.
- **Sie richten Ihre Arbeitsumgebung ein** für KI-Entwicklung auf macOS.

Wenn Sie eine Plug-and-Play-Distribution zum „Installieren und Vergessen" suchen — das ist es nicht:
Bei OFFICINA geht es um Methode und bewusste Auswahl, nicht um eine fertige schlüsselfertige
Zusammenstellung.

## Was ist enthalten

| Bereich | Was er enthält |
|---|---|
| [`skills/`](skills/) | Portable Skills — wiederverwendbare Modi und Abläufe (Claude Code und kompatible Umgebungen) |
| [`claude/`](claude/) | Claude-Code-Schicht — Subagents und Einstellungen |
| [`methodology/`](methodology/) | Regeln und Leitfäden — *wie man arbeitet* |
| [`cli/`](cli/) | Persönliche CLI-Tools und gemeinsame Bibliotheken |
| [`vscode/`](vscode/) | VS-Code-Einstellungen, Tastenkürzel, kuratierte Erweiterungen |
| [`install/`](install/) | Installations- und Layout-Skripte |

## Wie die Methode funktioniert

Ausgangspunkt ist eine Frage — womit die Aufgabe zu lösen ist:

- Eine deterministische Aktion, kein Modell nötig — **CLI** (z. B. richtet [`jig`](cli/jig/) ein neues Repository mit einem Befehl ein)
- Kontext verstehen und je nach Situation entscheiden — **Skill**
- Eine isolierte Teilaufgabe oder eine Persona-Rolle — **Subagent**
- Ein wiederkehrender Zyklus nach Zeitplan — **Loop**

Danach folgt der passende Leitfaden in [`methodology/`](methodology/): Struktur, Konventionen, eine
Checkliste vor dem Deploy. Die Methode ist gemeinsam; das Artefakt lässt sich zwischen Projekten
übertragen.

[`jig`](cli/jig/) ist die Methode als Werkzeug: mit einem Befehl legt es genau die Dokumentstruktur und
Claude-Einrichtung an, die die Leitfäden in [`methodology/`](methodology/) beschreiben.

## Schnellstart

Erfordert **macOS**. Klonen Sie das Repository und öffnen Sie den Bereich, der zu Ihrem Bedarf passt —
jeder bringt seine eigene README mit und, wo sinnvoll, einen Installer mit nur einem Befehl unter
[`install/`](install/):

```bash
git clone https://github.com/KirKruglov/OFFICINA.git
cd OFFICINA
```

Der erste fertige Installer richtet die VS-Code-Umgebung ein:

```bash
./install/vscode.sh   # Einstellungen, Tastenkürzel, MesloLGS Nerd Font, kuratierte Erweiterungen
```

## Philosophie

- **Das richtige Werkzeug für die Aufgabe.** Deterministisches wandert in ein Skript, kontextabhängige
  Entscheidungen ans Modell. Ein überflüssiges Modell im Loop bringt Kosten und Unsicherheit.
- **Verlässlichkeit beruht auf Grenzen.** Grenzen, Budgets und Prüfungen halten das System zusammen.
  Ein autonomer Loop braucht etwas, das „nein" sagen kann.
- **Wissen steckt im Artefakt.** Ein Leitfaden, Skill oder Subagent wird wiederverwendet; einmal
  Hergeleitetes wird nicht jedes Mal neu hergeleitet.
- **Handverlesen.** Was ins Repository gelangt, ist in der Praxis erprobt und bewusst ausgewählt.

## Mitwirken

Issues, Fragen und kleine Korrekturen sind willkommen — siehe [CONTRIBUTING.md](CONTRIBUTING.md).
OFFICINA ist ein offenes Projekt.

Nützlich? Vergeben Sie einen Stern — so findet der nächste Entwickler das Projekt. Forken Sie es,
nehmen Sie, was passt, und passen Sie es an Ihr Setup an.

## Lizenz

Veröffentlicht unter der [MIT-Lizenz](LICENSE) — © 2026 Kir Kruglov. Frei zu verwenden, zu ändern und
zu verbreiten.
