# World of Hyatt CLI

Scriptable World of Hyatt award availability for humans and agents.

This CLI turns Hyatt hotel metadata and points-calendar pages into structured JSON. It can resolve a city into Hyatt spirit codes, scan hotels by room category, and keep one-night versus multi-night availability separate.

## Current Transport Model

Hyatt commonly returns HTTP 403 to raw programmatic requests for the hotel metadata and rate-calendar pages. The working default is browser-backed:

```text
hyatt-cli command -> browser-use loads Hyatt page -> CLI extracts page JSON/window.STORE -> normalized JSON output
```

That means live searches require `browser-use` on `PATH`. By default the CLI uses one reusable headed browser session named `hyatt-cli`; it navigates the existing tab between Hyatt URLs instead of opening and closing a new browser for every hotel. On macOS, the CLI also minimizes Hyatt Chrome windows after navigation so the session stays out of the way.

```bash
pipx install browser-use
hyatt-cli doctor --json
```

Raw HTTP is still available for debugging, but it is not the recommended agent path:

```bash
HYATT_TRANSPORT=http hyatt-cli calendars --spirit-code KULAL --start-date 2026-09-01 --end-date 2026-09-02
```

Useful transport env vars:

| Name | Default | Use |
| --- | --- | --- |
| `HYATT_TRANSPORT` | `browser` | Set `http` or `direct` only to debug raw HTTP. |
| `HYATT_BROWSER_SESSION` | `hyatt-cli` | Browser-use session name. |
| `HYATT_BROWSER_PROFILE` | unset | Optional browser-use profile. Leave unset unless that profile is known to work. |
| `HYATT_BROWSER_BACKGROUND` | enabled | On macOS, minimize Hyatt Chrome windows after navigation. Set `0` to leave the window visible. |
| `HYATT_BROWSER_HEADLESS` | unset | Set `true` to try hidden/headless browser mode. Hyatt may block this. |
| `HYATT_BROWSER_FALLBACK` | enabled | Set `0` to disable browser fallback. |
| `HYATT_HOTELS_CACHE_MAX_AGE` | `24h` | Freshness window for the slow-changing Hyatt hotel metadata cache. Set `0` to force live refreshes in `auto` mode. |
| `HYATT_COOKIES` | unset | Optional raw Cookie header for direct HTTP debugging. Not required for normal browser-backed searches. |

Close the reusable browser session when done:

```bash
browser-use --session hyatt-cli close
```

Headless note: basic headless Chromium hit Hyatt's "browser did something unexpected" page during verification, so the default remains headed. The UX fix is session reuse plus minimizing/backgrounding the Hyatt window.

Hotel metadata note: `https://www.hyatt.com/explore-hotels/service/hotels` changes slowly, so `hyatt-cli hotels` and city resolution cache normalized hotel rows in the local SQLite store. Default `auto` mode reuses that cache for 24 hours before opening Hyatt again. Use `--data-source live` to force a refresh, `--no-cache` to bypass cache reads/writes, or `HYATT_HOTELS_CACHE_MAX_AGE=0` to disable this freshness shortcut.

## Install

```bash
git clone https://github.com/jiahongc/hyatt-cli.git
cd hyatt-cli
go build -o bin/hyatt-cli ./cmd/hyatt-cli
go build -o bin/hyatt-mcp ./cmd/hyatt-mcp
```

Or install with Go:

```bash
go install github.com/jiahongc/hyatt-cli/cmd/hyatt-cli@latest
go install github.com/jiahongc/hyatt-cli/cmd/hyatt-mcp@latest
```

## Quick Start

Check runtime readiness:

```bash
hyatt-cli doctor --json
hyatt-cli agent-context --pretty
```

Resolve a city into Hyatt hotels and spirit codes:

```bash
hyatt-cli hotels resolve-city \
  --city "New York City" \
  --json --no-input --no-color --yes \
  --select name,spiritCode,city,state,country,category,brand
```

Check one hotel calendar:

```bash
hyatt-cli calendars \
  --spirit-code KULAL \
  --start-date 2026-09-01 \
  --end-date 2026-09-02 \
  --room-category STANDARD_ROOM \
  --json --no-input --no-color --yes \
  --select spiritCode,nights,roomCategory,days
```

Scan specific hotels:

```bash
hyatt-cli scan hotel \
  --hotels KULAL,KUAGH \
  --start 2026-09-01 \
  --end 2026-09-05 \
  --nights 2 \
  --room-categories STANDARD_ROOM,STANDARD_SUITE \
  --json --no-input --no-color --yes \
  --select spiritCode,date,checkinDate,checkoutDate,nights,roomCategory,isStandardRoom,available,pointsValue,pointsLevel
```

Scan every matching hotel in a city:

```bash
hyatt-cli scan city \
  --city "Kuala Lumpur" \
  --start 2026-09-01 \
  --end 2026-09-02 \
  --nights 1 \
  --room-categories STANDARD_ROOM \
  --json --no-input --no-color --yes \
  --timeout 360s \
  --select spiritCode,date,nights,roomCategory,isStandardRoom,available,pointsValue,pointsLevel
```

## Agent Contract

This repo is intended to be usable by Claude, Codex, Hermes, OpenClaw, and other shell-capable agents.

Agent rules:

- Start with `hyatt-cli doctor --json` and `hyatt-cli agent-context --pretty`.
- Prefer `--json --no-input --no-color --yes --select ...` for exact, stable output.
- `--agent` is a shorthand for `--json --compact --no-input --no-color --yes`; use explicit `--select` when nested calendar rows matter.
- Parse data from `.results` when output is wrapped with provenance metadata.
- Treat `.meta.source == "browser"` as normal for live Hyatt searches.
- Do not chase `HYATT_COOKIES` unless explicitly debugging raw HTTP.
- Let `auto` mode reuse cached hotel metadata; force `--data-source live` only when you need Hyatt's latest hotel list.
- Avoid broad city scans when a user already supplied spirit codes; city scans open one browser-backed calendar page per hotel and room category.
- Expect one reusable browser tab for live scans; do not close it between calls unless you need to reset the session.
- On macOS, the CLI minimizes Hyatt Chrome windows by default. Set `HYATT_BROWSER_BACKGROUND=0` if you need to watch the browser.
- Use `--timeout 120s` or higher for live browser-backed scans; use `--timeout 360s` for city scans.

Fast path for agents:

1. Keep the default `HYATT_BROWSER_SESSION=hyatt-cli` warm across related calls.
2. If the user gave a city, run `hotels resolve-city` once; it uses cached hotel metadata when fresh.
3. Prefer one `scan hotel --hotels code1,code2,...` call over many one-hotel calls.
4. Select only needed fields, usually `spiritCode,date,nights,roomCategory,isStandardRoom,available,pointsValue,pointsLevel`.
5. Do not close the browser-use session until the task is done.

Measured on this Mac during verification: a cold two-hotel scan took about 5.3 seconds; a warm-session two-hotel scan took about 1.2 seconds.

Fast command choice:

| User intent | Best command |
| --- | --- |
| "What Hyatt hotels are in this city?" | `hyatt-cli hotels resolve-city --city "<city>" ...` |
| "Check this hotel code/date" | `hyatt-cli calendars --spirit-code <code> --start-date <in> --end-date <out> ...` |
| "Check several known hotel codes" | `hyatt-cli scan hotel --hotels <codes> --start <date> --end <date> --nights <n> ...` |
| "I only know the city" | `hyatt-cli scan city --city "<city>" --start <date> --end <date> --nights <n> ...` |
| "Is it standard room or suite?" | Include `--room-categories STANDARD_ROOM,STANDARD_SUITE` and select `roomCategory,isStandardRoom`. |

## Core Commands

### `hotels`

Fetch Hyatt hotel metadata:

```bash
hyatt-cli hotels --json --select name,spiritCode,city,state,country,category,brand
```

Resolve a city:

```bash
hyatt-cli hotels resolve-city --city "New York City" --json --select name,spiritCode,city,state,country,category,brand
```

### `calendars`

Fetch and normalize one Hyatt points calendar page:

```bash
hyatt-cli calendars \
  --spirit-code KUAGH \
  --start-date 2027-01-10 \
  --end-date 2027-01-12 \
  --room-category STANDARD_ROOM \
  --json --select spiritCode,nights,roomCategory,days
```

Length of stay matters. A one-night request and a two-night request are different Hyatt searches because checkout date changes.

### `scan hotel`

Scan one or more spirit codes over a date range:

```bash
hyatt-cli scan hotel \
  --hotels KULAL \
  --start 2026-09-01 \
  --end 2026-09-05 \
  --nights 2 \
  --room-categories STANDARD_ROOM \
  --json --select spiritCode,date,checkinDate,checkoutDate,nights,roomCategory,isStandardRoom,available,pointsValue
```

### `scan city`

Resolve a city, then scan all matched spirit codes:

```bash
hyatt-cli scan city \
  --city "Kuala Lumpur" \
  --start 2026-09-01 \
  --end 2026-09-02 \
  --nights 1 \
  --room-categories STANDARD_ROOM \
  --json --timeout 360s \
  --select spiritCode,date,nights,roomCategory,isStandardRoom,available,pointsValue
```

City scans are slower than hotel scans because each matched hotel/category needs its own browser-backed calendar fetch.

## Room Categories

Common values:

- `STANDARD_ROOM`
- `STANDARD_SUITE`
- `PREMIUM_SUITE`
- `CLUB`

The CLI emits:

- `roomCategory`: the Hyatt category returned for the row.
- `isStandardRoom`: `true` only for `STANDARD_ROOM`.
- `nights`: the requested length of stay.
- `checkinDate` / `checkoutDate`: row-specific stay dates.

## MCP

This repository includes `hyatt-mcp` for Claude Desktop and other MCP clients.

```bash
go install github.com/jiahongc/hyatt-cli/cmd/hyatt-mcp@latest
```

Claude Desktop example:

```json
{
  "mcpServers": {
    "hyatt": {
      "command": "hyatt-mcp",
      "env": {
        "HYATT_TRANSPORT": "browser"
      }
    }
  }
}
```

The MCP server mirrors the CLI command surface. For availability work, command-mirror tools using `hyatt-cli --agent --select ...` are often easier for agents than raw endpoint tools.

## Troubleshooting

### `HTTP 403`

Hyatt blocked raw HTTP. Use the default browser transport:

```bash
unset HYATT_TRANSPORT
hyatt-cli doctor --json
```

If `browser-use` is missing:

```bash
pipx install browser-use
```

### Empty or missing local data

Local analysis commands read SQLite snapshots. If a command says snapshots have not been synced, either run a live `scan ...` command or sync the needed resources after browser transport is working.

### City scans are slow

First resolve the city:

```bash
hyatt-cli hotels resolve-city --city "New York City" --json --select name,spiritCode,category
```

Then scan only the spirit codes you actually care about:

```bash
hyatt-cli scan hotel --hotels NYCAM,NYCGH --start 2026-11-05 --end 2026-11-08 --nights 2 --room-categories STANDARD_ROOM --json
```

## Verification

Latest live endpoint smoke results are in:

```text
artifacts/verification/hyatt-endpoint-smoke-2026-06-21/results.md
```

The current verified state:

- Hotel metadata works through browser transport.
- City resolution works for US and Asia examples.
- Calendar lookup works for standard rooms and suites.
- One-night and multi-night scans are separate.
- `scan city` works, but is slower because it opens multiple browser-backed pages.
