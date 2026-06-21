# Agent Readiness Audit - 2026-06-21

## Scope

Reviewed the README, local agent instructions, skill metadata, runtime `agent-context`, MCP context, manifest metadata, doctor output, and the live Hyatt transport path.

Target agents: Claude, Codex, Hermes, OpenClaw, and other shell-capable agents.

## Findings Fixed

- README examples were stale. They referenced `calendars get` and `hotels sync`, which are not the primary current command forms.
- Docs over-emphasized `HYATT_COOKIES`, even though the working live path is browser-backed.
- `agent-context` and MCP context marked `HYATT_COOKIES` as required, which would send agents down the wrong setup path.
- Live Hyatt commands wasted time by trying direct HTTP first, then falling back after the expected 403.
- MCP typed endpoint tools still used raw HTTP, while the CLI command mirror had the working browser-backed behavior.
- `doctor` treated missing `HYATT_COOKIES` as an error even when browser transport was available.
- Skill instructions had stale transport guidance and an outdated calendar recipe.

## Changes Made

- Hyatt hotel metadata and rate-calendar paths are browser-first by default.
- `HYATT_TRANSPORT=http` / `direct` remains available for debugging direct HTTP.
- `HYATT_BROWSER_FALLBACK=0` still disables browser fallback.
- `agent-context` now reports browser auth mode and optional env vars.
- MCP context now tells clients to use browser transport and command-mirror tools for availability workflows.
- MCP typed `calendars_get` and `hotels_list` now shell through the companion CLI, so they use the same browser-first transport and normalization as CLI users.
- README now focuses on the current browser-first workflow, exact JSON flags, and fast command selection.
- `AGENTS.md` now gives a short universal playbook for repo-local agents.
- `SKILL.md` now requires `browser-use` and removes the stale cookie-first mental model.
- `doctor` now reports missing cookies as OK when browser transport is the default path.

## Agent Performance Recommendations

- Prefer `scan hotel` over `scan city` when spirit codes are known.
- Run `hotels resolve-city` first when the user only gave a city, then narrow to the few hotel codes that matter.
- Always use `--select` for agent runs.
- Use `--timeout 120s` for hotel scans and `--timeout 360s` for city scans.
- Treat `.meta.source == "browser"` as expected, not a fallback failure.

## Verification

- `go test ./...` passed.
- `go build -o build/stage/bin/hyatt-cli ./cmd/hyatt-cli` passed.
- `go build -o build/stage/bin/hyatt-mcp ./cmd/hyatt-mcp` passed.
- Live `hyatt-cli hotels` returned 2,979 rows with `meta.reason = hyatt_browser_first`.
- `hyatt-cli agent-context --pretty` reports `auth.mode = browser`.
- `hyatt-cli doctor --json` reports browser transport and does not require `HYATT_COOKIES`.
