"""Unit tests for auto-commit pure logic + entry-point smoke test."""
import importlib.util
import pathlib
import subprocess
import sys
from importlib.machinery import SourceFileLoader

_ENTRY = pathlib.Path(__file__).parent / "auto-commit"
_loader = SourceFileLoader("auto_commit", str(_ENTRY))
_spec = importlib.util.spec_from_loader("auto_commit", _loader)
ac = importlib.util.module_from_spec(_spec)
_loader.exec_module(ac)


# --- file_dir ---

def test_file_dir_nested():
    assert ac.file_dir("cli/auto-commit/main") == "cli/auto-commit"


def test_file_dir_root():
    assert ac.file_dir("README.md") == ""


# --- check_gate ---

def test_gate_ok_single_dir():
    assert ac.check_gate(["cli/auto-commit/a", "cli/auto-commit/b"]) is None


def test_gate_empty():
    assert "no changes" in ac.check_gate([])


def test_gate_too_many_files():
    files = [f"cli/x/f{i}" for i in range(4)]
    assert "limit" in ac.check_gate(files)


def test_gate_two_top_dirs():
    assert "directories" in ac.check_gate(["cli/x/a", "planning/y/b"])


def test_gate_two_subdirs_same_top():
    # one top "cli" but different parent directories → refuse (strictly "one directory")
    assert "directories" in ac.check_gate(["cli/a/x", "cli/b/y"])


def test_gate_two_root_files_one_group():
    # both files at the root → single top "" → no directory-based refusal
    assert ac.check_gate(["A.md", "B.md"]) is None


def test_gate_empty_with_untracked():
    # tracked empty but new untracked present → message points to them, not "no changes"
    msg = ac.check_gate([], ["planning/new.md"])
    assert "untracked" in msg
    assert "committing-changes" in msg


def test_gate_empty_no_untracked():
    # tracked empty and no untracked → the plain message
    assert ac.check_gate([], []) == "no changes to commit"


# --- derive_verb ---

def test_verb_all_modified():
    assert ac.derive_verb(["M", "M"]) == "update"


def test_verb_all_added():
    assert ac.derive_verb(["A"]) == "add"


def test_verb_all_deleted():
    assert ac.derive_verb(["D", "D"]) == "remove"


def test_verb_rename():
    assert ac.derive_verb(["R100"]) == "rename"


def test_verb_mixed_falls_to_update():
    assert ac.derive_verb(["A", "D"]) == "update"


# --- detect_scope ---

def test_scope_common_dir():
    assert ac.detect_scope(["cli/auto-commit/a", "cli/auto-commit/b"]) == "auto-commit"


def test_scope_partial_common():
    assert ac.detect_scope(["cli/a/x", "cli/b/y"]) == "cli"


def test_scope_root_none():
    assert ac.detect_scope(["A.md", "B.md"]) is None


def test_scope_single_file():
    assert ac.detect_scope(["cli/auto-commit/main"]) == "auto-commit"


# --- build_subject / compose ---

def test_build_subject_update():
    assert ac.build_subject("update", ["cli/x/a.py", "cli/x/b.py"]) == "update a.py, b.py"


def test_build_subject_add():
    assert ac.build_subject("add", ["cli/x/parser.py"]) == "add parser.py"


def test_compose_with_scope():
    assert ac.compose("auto-commit", "update a.py") == "auto-commit: update a.py"


def test_compose_no_scope():
    assert ac.compose(None, "update a.py") == "update a.py"


def test_compose_truncates_to_limit():
    head = ac.compose("scope", "update " + "x" * 100)
    assert len(head) <= ac.SUBJECT_LIMIT


# --- scan_secrets ---

def test_secret_filename():
    hits = ac.scan_secrets("", [".env"])
    assert any(".env" in h for h in hits)


def test_secret_pem_file():
    hits = ac.scan_secrets("", ["certs/server.pem"])
    assert hits


def test_secret_content_added_line():
    diff = "+API = 'sk-abcdefghijklmnopqrstuvwxyz'"
    hits = ac.scan_secrets(diff, ["cli/x/a.py"])
    assert any("OpenAI/Anthropic" in h for h in hits)


def test_secret_ignores_context_lines():
    diff = " API = 'sk-abcdefghijklmnopqrstuvwxyz'"  # not an added line
    assert ac.scan_secrets(diff, ["cli/x/a.py"]) == []


def test_clean_diff_no_secrets():
    assert ac.scan_secrets("+x = 1", ["cli/x/a.py"]) == []


def test_secret_env_example_allowed():
    # .env.example is allowed by the guide — the name must not block
    assert ac.scan_secrets("", [".env.example"]) == []


def test_secret_doc_name_allowed():
    # documentation with "secret"/"credentials" in the name — not blocked by name
    assert ac.scan_secrets("", ["docs/secret-rotation.md"]) == []


def test_secret_content_in_doc_still_caught():
    # but a real key in .md content is caught by the content pattern
    diff = "+token = 'ghp_abcdefghijklmnopqrstuvwxyz12'"
    assert ac.scan_secrets(diff, ["docs/notes.md"])


# --- entry-point smoke test ---

def test_help_smoke():
    r = subprocess.run([sys.executable, str(_ENTRY), "--help"], capture_output=True, text=True)
    assert r.returncode == 0
    assert "auto-commit" in r.stdout
