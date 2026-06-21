# World of Hyatt CLI

**Hyatt award availability as a scriptable, local, agent-readable CLI.**

Hyatt's calendar is useful but browser-bound and property-by-property. This CLI resolves cities into Hyatt hotels and spirit codes, turns points-calendar pages into structured local data, and separates standard-room availability from other room categories. It also treats length of stay as an availability-changing input, so one-night results are never reused as proof of multi-night award space.

Learn more at [World of Hyatt](https://www.hyatt.com).

Created by [@jiahongc](https://github.com/jiahongc) (Jiahong Chen).

## Install

Build from this repository:

```bash
git clone https://github.com/jiahongc/hyatt-cli.git
cd hyatt-cli
go build -o bin/hyatt-cli ./cmd/hyatt-cli
go build -o bin/hyatt-mcp ./cmd/hyatt-mcp
```

Or install directly with Go:

```bash
go install github.com/jiahongc/hyatt-cli/cmd/hyatt-cli@latest
go install github.com/jiahongc/hyatt-cli/cmd/hyatt-mcp@latest
```

This project also includes [SKILL.md](./SKILL.md) for agents that support local skill files. Install or reference that file from this repository alongside the `hyatt-cli` binary.

## Use with Claude Desktop

This repository includes an MCP server binary for Claude Desktop and other MCP clients.

The MCP server reuses your local browser session. Set it up first if you haven't:

```bash
hyatt-cli auth login --chrome
```

Install the MCP binary:

```bash
go install github.com/jiahongc/hyatt-cli/cmd/hyatt-mcp@latest
```

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "hyatt": {
      "command": "hyatt-mcp"
    }
  }
}
```

## Authentication

The first shippable flow uses public Hyatt Points Calendar pages and browser-compatible replay. Logged-in Hyatt sessions are short-lived, so account-specific perks or durable authenticated availability are intentionally out of initial scope.

## Quick Start

```bash
# Check local configuration and transport assumptions without touching Hyatt.
hyatt-cli doctor --dry-run

# Fetch one browser-sniffed Points Calendar surface and keep only fields an agent needs.
hyatt-cli calendars get --spirit-code kulal --start-date 2026-09-01 --end-date 2026-09-02 --json --select spiritCode,days

# Preview Hyatt property metadata hydration before building the local hotel index.
hyatt-cli hotels sync --dry-run

# Turn a city into the matching Hyatt hotels and the spirit codes used by availability scans.
hyatt-cli hotels resolve-city --city "New York City" --json --select name,spiritCode,city,state,category

# Scan every matching city hotel and label standard-room versus other room-category award space.
hyatt-cli scan city --city "New York City" --start 2026-09-01 --end 2026-09-07 --nights 2 --room-categories STANDARD_ROOM,SUITE --agent

# Find flexible-date award clusters across a shortlist.
hyatt-cli awards density --hotels CHIRH,NYCUA,PARPH --start 2026-07-01 --end 2026-09-30 --nights 2 --bucket week --agent

```

## Unique Features

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

## Recipes


### Check one Points Calendar page

```bash
hyatt-cli calendars get --spirit-code kulal --start-date 2026-09-01 --end-date 2026-09-02 --room-category STANDARD_ROOM --agent --select spiritCode,roomCategory,days.2026-09-08.STANDARD_ROOM.pointsValue
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

## Usage

Run `hyatt-cli --help` for the full command reference and flag list.

## Commands

### calendars

Fetch Hyatt Points Calendar HTML pages

- **`hyatt-cli calendars`** - Fetch a Hyatt Points Calendar page for a hotel spirit code

### hotels

Fetch Hyatt property metadata used to resolve city searches into hotel spirit codes

- **`hyatt-cli hotels`** - Fetch Hyatt hotel metadata, including names, locations, categories, and spirit codes


## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
hyatt-cli calendars --spirit-code example-value

# JSON for scripting and agents
hyatt-cli calendars --spirit-code example-value --json

# Filter to specific fields
hyatt-cli calendars --spirit-code example-value --json --select id,name,status

# Dry run — show the request without sending
hyatt-cli calendars --spirit-code example-value --dry-run

# Agent mode — JSON + compact + no prompts in one flag
hyatt-cli calendars --spirit-code example-value --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select id,name` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Read-only by default** - this CLI does not create, update, delete, publish, send, or mutate remote resources
- **Offline-friendly** - sync/search commands can use the local SQLite store when available
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Health Check

```bash
hyatt-cli doctor
```

Verifies configuration, credentials, and connectivity to the API.

## Configuration

Config file: ``

Static request headers can be configured under `headers`; per-command header overrides take precedence.

Environment variables:

| Name | Kind | Required | Description |
| --- | --- | --- | --- |
| `HYATT_COOKIES` | per_call | Yes | Set to your API credential. |

### agentcookie (optional)

If you use agentcookie to sync secrets across machines, this CLI auto-adopts agentcookie-managed credentials with no extra setup. When the daemon writes to this CLI's config, `hyatt-cli doctor` reports `agentcookie: detected` and `auth-status` labels the source as `agentcookie`. Skip this section if you don't use agentcookie - the CLI works the same as any other.

## Troubleshooting
**Authentication errors (exit code 4)**
- Run `hyatt-cli doctor` to check credentials
- Verify the environment variable is set: `echo $HYATT_COOKIES`
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

### API-specific
- **Hyatt returns E6020, 403, or 429.** — Run `hyatt-cli doctor hyatt` and refresh the browser-cookie path before scanning again.
- **Scan results look sparse or uneven.** — Run `hyatt-cli awards coverage --hotels <codes> --start <date> --end <date>` to identify missing cached months or room categories.
- **Logged-in Hyatt data disappears quickly.** — Treat logged-in state as temporary; rerun the public calendar flow or use a fresh browser profile capture.

## HTTP Transport

This CLI uses Chrome-compatible HTTP transport for browser-facing endpoints. It does not require a resident browser process for normal API calls.

## Discovery Signals

This CLI was generated with browser-captured traffic analysis.
- Target observed: https://www.hyatt.com/explore-hotels/rate-calendar
- Capture coverage: 1 API entries from 2 total network entries
- Reachability: browser_clearance_http (80% confidence)
- Protocols: html-embedded-state (90% confidence)
- Protection signals: hyatt-browser-clearance (80% confidence)
- Discovery hints: Preserve browser-compatible transport, parse the JavaScript assignment window.STORE = {...}; from HTML into normalized availability rows, and treat direct HTTP 403/429 as expected unless Chrome-cookie replay also fails.
- Candidate command ideas: calendar — Fetch and parse a Hyatt Points Calendar page for one hotel spirit code.; scan — Repeat calendar fetches across multiple spirit codes and date windows to find points availability.

Warnings from discovery:
- html-state-not-standard-json-script: The calendar payload is a JavaScript assignment, not script#__NEXT_DATA__.

---

## Sources & Inspiration

This CLI was built by studying these projects and resources:

- [**Hyatt-award-search**](https://github.com/dewdream/Hyatt-award-search) — Python (8 stars)
- [**stayexpert-hyatt**](https://github.com/StayExpert/hyatt) — JavaScript (1 stars)
- [**hyattvalue**](https://github.com/sottenad/hyattvalue) — JavaScript
