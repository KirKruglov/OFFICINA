# Security Policy

OFFICINA is a curated showcase of a personal development setup — configs, shell and install scripts,
and documentation. It ships no networked service, but the scripts here run on your machine and the
files are meant to stay free of any secrets or personal data. This policy covers both.

## Reporting a vulnerability

**Please do not open a public issue for security problems.**

Report privately through GitHub:

1. Go to the repository's **Security** tab → **Report a vulnerability**.
2. Describe the issue, the affected file(s), and how to reproduce or observe it.

This uses GitHub's private vulnerability reporting, so the details stay confidential until a fix is
out. You'll get a best-effort response — this is a personal project maintained in spare time.

## What to report

- **Leaked secrets or personal data** — a committed token, credential, private key, or a
  machine-specific path (`/Users/<name>/...`) that slipped past sanitization.
- **Unsafe script behavior** — anything in [`install/`](install/) or `shell/` that could destroy
  data, execute untrusted input, or modify files outside its intended scope.
- **Misleading guidance** — a documented setting or command that meaningfully weakens the security
  of a machine that applies it.

## What is not in scope

- Vulnerabilities in third-party tools or VS Code extensions referenced here — report those to their
  respective maintainers.
- Platforms other than **macOS**, which is the only supported target.

## Supported versions

Only the current state of the `main` branch is maintained. There are no released versions or
backports; fixes land on `main`.

## Before you run anything

These scripts are meant to be read before they are run. Review [`install/`](install/) scripts and any
shell config, and confirm they match your expectations, before executing them on your machine.
