# World of Hyatt CLI Agent Guide

This repository contains the standalone World of Hyatt award availability CLI and MCP server. Keep local edits narrow, verify behavior before shipping, and avoid unrelated cleanup.

## Local Operating Contract

Start by asking the CLI for current runtime truth:

```bash
hyatt-cli doctor --json
hyatt-cli agent-context --pretty
```

Live Hyatt metadata and rate-calendar calls are browser-backed by default. Do not start by debugging cookies or raw HTTP. Make sure `browser-use` is on `PATH`; only use `HYATT_TRANSPORT=http` when explicitly debugging direct HTTP behavior.

The default browser transport reuses one headed `browser-use` session named `hyatt-cli` and navigates the existing tab between Hyatt URLs. On macOS it minimizes Hyatt Chrome windows after navigation by default; set `HYATT_BROWSER_BACKGROUND=0` only when you need to watch the browser. Do not close that session between calls unless you need to reset it. `HYATT_BROWSER_HEADLESS=true` is available as an opt-in experiment, but Hyatt may block basic headless mode.

Hyatt hotel metadata from `/explore-hotels/service/hotels` is cache-backed because it changes slowly. In default `auto` mode, reuse the local cache for city resolution and `hyatt-cli hotels`; only pass `--data-source live` when you intentionally need a fresh hotel list. `--no-cache` bypasses reads and writes, and `HYATT_HOTELS_CACHE_MAX_AGE=0` disables the hotel metadata freshness shortcut.

Use runtime discovery instead of relying on a copied command list:

```bash
hyatt-cli which "<capability>" --json
hyatt-cli <command> --help
```

Use exact machine flags when you know the fields you need:

```bash
hyatt-cli <command> --json --no-input --no-color --yes --select field1,field2
```

Add `--agent` when you want the shorthand for JSON, compact output, non-interactive defaults, no color, and confirmation-safe scripting:

```bash
hyatt-cli <command> --agent
```

For nested calendar data, prefer explicit `--select` over relying on compact defaults. Parse wrapped command output from `.results`; `.meta.source == "browser"` is normal for live Hyatt searches.

Fast command choice:

- City to hotel codes: `hyatt-cli hotels resolve-city --city "<city>" --json --select name,spiritCode,city,state,country,category,brand`
- One hotel/date window: `hyatt-cli calendars --spirit-code <code> --start-date <in> --end-date <out> --room-category STANDARD_ROOM --json --select spiritCode,nights,roomCategory,days`
- Known hotel codes: `hyatt-cli scan hotel --hotels <codes> --start <date> --end <date> --nights <n> --room-categories STANDARD_ROOM --json --select spiritCode,date,nights,roomCategory,isStandardRoom,available,pointsValue`
- City-wide scan: `hyatt-cli scan city --city "<city>" --start <date> --end <date> --nights <n> --room-categories STANDARD_ROOM --json --timeout 360s --select spiritCode,date,nights,roomCategory,isStandardRoom,available,pointsValue`

Speed rules:

- Keep the `hyatt-cli` browser-use session warm across related calls.
- Resolve a city once with cached hotel metadata, then scan selected spirit codes with `scan hotel`.
- Batch hotel codes in one command instead of looping in the agent.
- Always pass `--select`; avoid asking for full calendar payloads unless needed.

Before running an unfamiliar command that may mutate remote state, inspect its help and prefer a dry run:

```bash
hyatt-cli <command> --help
hyatt-cli <command> --dry-run --agent
```

Use `--yes --no-input` only after the target, arguments, and side effects are clear.

For install, auth, examples, and longer product guidance, read `README.md` and `SKILL.md`. This file intentionally stays small so repo-local agents get invariant local guidance without duplicating the user docs.

## Release Notes

Use `CHANGELOG.md` for user-facing release notes when behavior changes. Do not bump versions or create release tags unless the user explicitly asks.
