---
name: hyatt
description: "Hyatt award availability as a scriptable, local, agent-readable CLI. Trigger phrases: `check Hyatt points availability`, `find Hyatt award nights`, `Hyatt hotels in New York City with points`, `search Hyatt award rooms by city`, `scan World of Hyatt hotels`, `Hyatt certificate fit`, `use hyatt`, `run hyatt`."
author: "Jiahong Chen"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  openclaw:
    requires:
      bins:
        - hyatt-cli
        - browser-use
    install:
      - kind: go
        bins: [hyatt-cli]
        module: github.com/jiahongc/hyatt-cli/cmd/hyatt-cli
---

# World of Hyatt CLI

## Prerequisites: Install the CLI

This skill drives the `hyatt-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install from this repository:
   ```bash
   go install github.com/jiahongc/hyatt-cli/cmd/hyatt-cli@latest
   ```
2. Verify: `hyatt-cli --version`
3. Ensure the reported install directory is on `$PATH` for the agent/runtime that will invoke this skill.

This installs into `$GOPATH/bin` (default `$HOME/go/bin`), so add that directory to `$PATH` if needed.

If `--version` reports "command not found" after install, the runtime cannot see the binary directory on `$PATH`. Do not proceed with skill commands until verification succeeds.

Live Hyatt hotel and calendar pages are browser-backed. Verify `browser-use` is installed and on `$PATH` before live searches:

```bash
pipx install browser-use
hyatt-cli doctor --json
```

Hyatt's calendar is useful but browser-bound and property-by-property. This CLI resolves cities into Hyatt hotels and spirit codes, turns points-calendar pages into structured local data, and separates standard-room availability from other room categories. It also treats length of stay as an availability-changing input, so one-night results are never reused as proof of multi-night award space.

## When to Use This CLI

Use this CLI when checking World of Hyatt points availability across properties, dates, room categories, or watchlists. It is best for repeatable award-search workflows where local snapshots, structured output, and agent-readable filters matter.

## Anti-triggers

Do not use this CLI for:
- Do not use this CLI to make or cancel Hyatt reservations.
- Do not use this CLI for durable logged-in account benefits or elite-status planning.
- Do not use this CLI as a replacement for Hyatt customer service when availability or booking rules are disputed.

## Unique Capabilities

These capabilities aren't available in any other tool for this API.

### Award decision helpers
- **`awards certificate-fit`** — Find available Hyatt award nights that fit a Cat 1-4 or Cat 1-7 certificate before it expires.

  _Use this when the user needs to burn a Hyatt certificate without manually checking category and date fit._

  ```bash
  hyatt-cli awards certificate-fit --cert cat1-4 --expires 2026-12-31 --start 2026-09-01 --end 2026-12-31 --agent
  ```
- **`awards offpeak`** — Find off-peak or unusually low-point Hyatt award clusters across synced calendar data.

  _Use this when the user wants low-points timing rather than a specific hotel._

  ```bash
  hyatt-cli awards offpeak --country US --start 2026-06-01 --end 2027-05-31 --min-nights 2 --agent
  ```
- **`awards room-ladder`** — Compare standard, club, and suite award availability side by side for one Hyatt stay window.

  _Use this when the user cares whether better room categories are open, not just standard rooms._

  ```bash
  hyatt-cli awards room-ladder --hotel CHIRH --start 2026-10-01 --end 2026-10-07 --agent
  ```

### Flexible-date search
- **`awards density`** — Show which dates or weeks have the most Hyatt award options across a hotel shortlist.

  _Use this before picking travel dates when flexibility matters more than a specific property._

  ```bash
  hyatt-cli awards density --hotels CHIRH,NYCUA,PARPH --start 2026-07-01 --end 2026-09-30 --bucket week --agent
  ```
- **`awards split-stay`** — Build viable multi-hotel Hyatt itineraries when no one property has the full stay.

  _Use this when the user will switch hotels to make an award trip work._

  ```bash
  hyatt-cli awards split-stay --hotels CHIRH,CHIJD --start 2026-08-01 --end 2026-08-15 --nights 5 --max-switches 1 --agent
  ```

### Watchlist intelligence
- **`watch volatility`** — Rank watched Hyatt hotels by how often award space opens, closes, or changes price.

  _Use this to decide which hard-to-book Hyatt searches deserve closer monitoring._

  ```bash
  hyatt-cli watch volatility --since 30d --limit 20 --agent
  ```

### Reachability mitigation
- **`awards coverage`** — Show stale, missing, or uneven local Hyatt calendar coverage before trusting a scan.

  _Use this when browser protection or expired cookies may have left gaps in cached availability._

  ```bash
  hyatt-cli awards coverage --hotels CHIRH,NYCUA --start 2026-07-01 --end 2026-12-31 --agent
  ```

## HTTP Transport

Hyatt commonly returns HTTP 403 to raw HTTP clients. For hotel metadata and points-calendar availability, this CLI uses `browser-use` by default, extracts the loaded page JSON or `window.STORE`, and normalizes that into CLI JSON.

Only set `HYATT_TRANSPORT=http` when explicitly debugging direct HTTP. `HYATT_COOKIES` is optional and is not required for the normal browser-backed search path.

## Discovery Signals

This CLI was generated with browser-observed traffic context.
- Capture coverage: 1 API entries from 2 total network entries
- Protocols: html-embedded-state (90% confidence)
- Generation hints: Use browser transport first for Hyatt hotel metadata and rate-calendar endpoints., Add a hand-authored parser that extracts the JavaScript assignment window.STORE = {...}; from HTML and emits normalized availability rows., Treat direct HTTP 403/429 as expected; HYATT_TRANSPORT=http is for debugging, not the default agent path.
- Candidate command ideas: calendar — Fetch and parse a Hyatt Points Calendar page for one hotel spirit code.; scan — Repeat calendar fetches across multiple spirit codes and date windows to find points availability.
- Caveats: html-state-not-standard-json-script: The calendar payload is a JavaScript assignment, not script#__NEXT_DATA__; built-in embedded-json extraction may not parse it without hand code.

## Command Reference

**calendars** — Fetch Hyatt Points Calendar HTML pages

- `hyatt-cli calendars` — Fetch a Hyatt Points Calendar page for a hotel spirit code

**hotels** — Fetch Hyatt property metadata used to resolve city searches into hotel spirit codes

- `hyatt-cli hotels` — Fetch Hyatt hotel metadata, including names, locations, categories, and spirit codes


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
hyatt-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query.

## Recipes

### Check one Points Calendar page

```bash
hyatt-cli calendars --spirit-code KULAL --start-date 2026-09-01 --end-date 2026-09-02 --room-category STANDARD_ROOM --json --no-input --no-color --yes --select spiritCode,nights,roomCategory,days
```

Fetch one Hyatt standard-room calendar page and narrow the nested output to a specific award row.

### Resolve a city into Hyatt hotel codes

```bash
hyatt-cli hotels resolve-city --city "New York City" --agent --select name,spiritCode,city,state,category
```

Find the hotels and spirit codes to use when the user only knows the city.

### Scan a city by room category

```bash
hyatt-cli scan city --city "New York City" --start 2026-09-01 --end 2026-09-07 --nights 2 --room-categories STANDARD_ROOM,SUITE --agent --select hotelName,spiritCode,date,nights,roomCategory,isStandardRoom,pointsValue
```

Search every matching city hotel and keep room type plus length of stay explicit in the output.

### Find certificate-eligible awards

```bash
hyatt-cli awards certificate-fit --cert cat1-4 --expires 2026-12-31 --start 2026-09-01 --end 2026-12-31 --agent
```

Combine local hotel categories and synced calendar rows to find certificate-fit nights.

### Compare flexible weekly availability

```bash
hyatt-cli awards density --hotels CHIRH,NYCUA,PARPH --start 2026-07-01 --end 2026-09-30 --nights 2 --bucket week --agent
```

See which weeks have the most two-night award options across a hotel shortlist.

### Build a split-stay option

```bash
hyatt-cli awards split-stay --hotels CHIRH,CHIJD --start 2026-08-01 --end 2026-08-15 --nights 5 --max-switches 1 --agent
```

Find a viable award itinerary across hotels when one property does not have all nights.

## Auth Setup

The normal live path uses public Hyatt Points Calendar pages through `browser-use`. Logged-in Hyatt sessions are short-lived, so account-specific perks or durable authenticated availability are intentionally out of scope. Do not require `HYATT_COOKIES` unless the user is explicitly debugging direct HTTP.

Run `hyatt-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color --yes`.

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  hyatt-cli scan hotel --hotels KULAL --start 2026-09-01 --end 2026-09-05 --nights 1 --agent --select spiritCode,date,nights,roomCategory,isStandardRoom,available,pointsValue
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Offline-friendly** — sync/search commands can use the local SQLite store when available
- **Non-interactive** — never prompts, every input is a flag
- **Read-only** — do not use this CLI for create, update, delete, publish, comment, upvote, invite, order, send, or other mutating requests

### Response envelope

Commands that read from the local store or the API wrap output in a provenance envelope:

```json
{
  "meta": {"source": "browser" | "live" | "local", "synced_at": "...", "reason": "..."},
  "results": <data>
}
```

Parse `.results` for data and `.meta.source` to know whether data came from the browser, direct HTTP, or local store. A human-readable `N results (...)` summary is printed to stderr only when stdout is a terminal AND no machine-format flag (`--json`, `--csv`, `--compact`, `--quiet`, `--plain`, `--select`) is set — piped/agent consumers and explicit-format runs get pure JSON on stdout.

## Agent Feedback

When you (or the agent) notice something off about this CLI, record it:

```
hyatt-cli feedback "the --since flag is inclusive but docs say exclusive"
hyatt-cli feedback --stdin < notes.txt
hyatt-cli feedback list --json --limit 10
```

Entries are stored locally at `~/.local/share/hyatt-cli/feedback.jsonl`. They are never POSTed unless `HYATT_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `HYATT_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

Write what *surprised* you, not a bug report. Short, specific, one line: that is the part that compounds.

## Output Delivery

Every command accepts `--deliver <sink>`. The output goes to the named sink in addition to (or instead of) stdout, so agents can route command results without hand-piping. Three sinks are supported:

| Sink | Effect |
|------|--------|
| `stdout` | Default; write to stdout only |
| `file:<path>` | Atomically write output to `<path>` (tmp + rename) |
| `webhook:<url>` | POST the output body to the URL (`application/json` or `application/x-ndjson` when `--compact`) |

Unknown schemes are refused with a structured error naming the supported set. Webhook failures return non-zero and log the URL + HTTP status on stderr.

## Named Profiles

A profile is a saved set of flag values, reused across invocations. Use it when a scheduled agent calls the same command every run with the same configuration - HeyGen's "Beacon" pattern.

```
hyatt-cli profile save briefing --json
hyatt-cli --profile briefing calendars --spirit-code example-value
hyatt-cli profile list --json
hyatt-cli profile show briefing
hyatt-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 4 | Authentication required |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `hyatt-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → see Prerequisites above
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/jiahongc/hyatt-cli/cmd/hyatt-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add hyatt-mcp -- hyatt-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which hyatt-cli`
   If not found, offer to install (see Prerequisites at the top of this skill).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   hyatt-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `hyatt-cli <command> --help`.
