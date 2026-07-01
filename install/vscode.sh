#!/usr/bin/env bash
#
# vscode.sh — install the VS Code configuration (macOS).
# Copies settings.json / keybindings.json and installs the extensions from extensions.txt.
# Model: copy the config to another machine (as opposed to symlinking for yourself). macOS only.
#
# Usage:
#   ./install/vscode.sh
#
set -euo pipefail

# --- Paths -----------------------------------------------------------------
# This script's directory → the source files live next to it in ../vscode.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC="$(cd "$SCRIPT_DIR/../vscode" && pwd)"
DEST="$HOME/Library/Application Support/Code/User"

# --- Environment checks ----------------------------------------------------
if [[ "$(uname)" != "Darwin" ]]; then
  echo "Error: this script targets macOS only." >&2
  exit 1
fi

if ! command -v code >/dev/null 2>&1; then
  echo "Error: 'code' CLI not found on PATH." >&2
  echo "In VS Code: Cmd+Shift+P → Shell Command: Install 'code' command in PATH." >&2
  exit 1
fi

for f in settings.json keybindings.json extensions.txt; do
  if [[ ! -f "$SRC/$f" ]]; then
    echo "Error: source file not found: $SRC/$f" >&2
    exit 1
  fi
done

mkdir -p "$DEST"

# --- Font ------------------------------------------------------------------
# settings.json references MesloLGS Nerd Font.
if ! command -v brew >/dev/null 2>&1; then
  echo "Warning: Homebrew not found — skipping font installation."
  echo "  Install MesloLGS Nerd Font manually, otherwise the terminal will render incorrectly."
elif brew list --cask font-meslo-lg-nerd-font >/dev/null 2>&1; then
  echo "MesloLGS Nerd Font is already installed."
else
  echo "Installing MesloLGS Nerd Font..."
  brew install --cask font-meslo-lg-nerd-font
fi

# --- Settings and keybindings ----------------------------------------------
# Existing files are saved to *.bak-<timestamp> rather than overwritten silently.
stamp="$(date +%Y%m%d-%H%M%S)"
for f in settings.json keybindings.json; do
  if [[ -f "$DEST/$f" ]]; then
    backup="$DEST/$f.bak-$stamp"
    cp "$DEST/$f" "$backup"
    echo "Existing $f backed up: $backup"
  fi
  cp "$SRC/$f" "$DEST/$f"
  echo "Copied $f → $DEST/$f"
done

# --- Extensions ------------------------------------------------------------
echo "Installing extensions from extensions.txt..."
while IFS= read -r ext || [[ -n "$ext" ]]; do
  [[ -z "$ext" ]] && continue
  code --install-extension "$ext" --force
done < "$SRC/extensions.txt"

echo
echo "Done. Fully restart VS Code (quit and reopen)."
