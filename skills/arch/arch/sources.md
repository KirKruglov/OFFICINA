# sources

A registry of verified primary sources for the technologies in `house-stack.md`. It sets the trust
boundary for verifying any fact about a technology: versions, release notes, breaking changes,
compatibility, platform requirements.

> **Paired with `house-stack.md` — replace it alongside your own stack, before the first run.** The
> registry below covers one author's technologies. The rules and the fetch notes are reusable as-is;
> the table is not.
>
> A technology absent from this table cannot be verified without a `❓` gate — the skill has to find
> the official source and get your confirmation before it may record a version. An empty or foreign
> registry turns every layer of a run into that gate.

Last updated: 2026-07-19. Every domain and version hint was checked by auto-fetch on that date.

## Rules

- **Proof** is only information from the domains of a technology's official perimeter and from
  ecosystem registries: `pypi.org`, `registry.npmjs.org`, `pkg.go.dev`.
- Blogs, Stack Overflow, forums, and articles may point the way; they are not proof.
- The **version hint** is a verified fast entry point for checking the current stable version. It is
  an optimization, not a restriction: any page inside the perimeter is equally trusted.
- A technology outside the registry gets its entry through the `arch-new` procedure (find the
  official source → user confirmation → record it here).

## Auto-fetch limitations

- HTML pages on `npmjs.com` return 403 to automated requests — for npm packages use the JSON
  endpoint `registry.npmjs.org/<package>/latest` (the `version` field).
- `wails.io` blocks auto-fetch (403) — for automated verification use `github.com/wailsapp/wails`
  and `pkg.go.dev`.
- The root of `react.dev` may return 404 to a fetcher — enter through `react.dev/versions`.

## Registry

| Technology | Official perimeter | Version hint |
|---|---|---|
| Go | `go.dev` | `go.dev/dl/` |
| Python | `python.org` | `python.org/downloads/` |
| TypeScript | `typescriptlang.org`, `github.com/microsoft/TypeScript` | `registry.npmjs.org/typescript/latest` |
| Node.js | `nodejs.org` | `nodejs.org/en/about/previous-releases`; for production — the active LTS |
| cobra | `cobra.dev`, `github.com/spf13/cobra` | `pkg.go.dev/github.com/spf13/cobra` |
| typer | `typer.tiangolo.com`, `github.com/fastapi/typer` | `pypi.org/project/typer/` |
| pandas | `pandas.pydata.org` | `pypi.org/project/pandas/` |
| React | `react.dev`, `github.com/facebook/react` | `registry.npmjs.org/react/latest` |
| Vite | `vite.dev`, `github.com/vitejs/vite` | `registry.npmjs.org/vite/latest` |
| FastAPI | `fastapi.tiangolo.com`, `github.com/fastapi/fastapi` | `pypi.org/project/fastapi/` |
| SQLite | `sqlite.org` | `sqlite.org/changes.html` |
| PostgreSQL | `postgresql.org` | `postgresql.org/support/versioning/` |
| Wails | `wails.io` (auto-fetch blocked), `github.com/wailsapp/wails` | `pkg.go.dev/github.com/wailsapp/wails/v2` |
| Electron | `electronjs.org`, `github.com/electron/electron` | `registry.npmjs.org/electron/latest` |
| React Native | `reactnative.dev`, `github.com/facebook/react-native` | `registry.npmjs.org/react-native/latest` |
| Expo | `expo.dev`, `docs.expo.dev`, `github.com/expo/expo` | `registry.npmjs.org/expo/latest` |
| uv | `docs.astral.sh/uv`, `github.com/astral-sh/uv` | `pypi.org/project/uv/` |
| pnpm | `pnpm.io`, `github.com/pnpm/pnpm` | `registry.npmjs.org/pnpm/latest` |
| react-router | `reactrouter.com`, `github.com/remix-run/react-router` | `registry.npmjs.org/react-router/latest` |
| MDX | `mdxjs.com`, `github.com/mdx-js/mdx` | `registry.npmjs.org/@mdx-js/rollup/latest` |
| uvicorn | `uvicorn.org`, `github.com/encode/uvicorn` | `pypi.org/project/uvicorn/` |
| SQLAlchemy | `sqlalchemy.org` | `pypi.org/project/sqlalchemy/` |
| psycopg | `psycopg.org`, `github.com/psycopg/psycopg` | `pypi.org/project/psycopg/` |
| alembic | `alembic.sqlalchemy.org` | `pypi.org/project/alembic/` |
