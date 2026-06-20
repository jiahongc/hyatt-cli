# World of Hyatt CLI

**Hyatt award availability as a scriptable, local, agent-readable CLI.**

Hyatt's calendar is useful but browser-bound and property-by-property. This CLI resolves cities into Hyatt hotels and spirit codes, turns points-calendar pages into structured local data, and separates standard-room availability from other room categories. It also treats length of stay as an availability-changing input, so one-night results are never reused as proof of multi-night award space.

Learn more at [World of Hyatt](https://www.hyatt.com).

Created by [@jiahongc](https://github.com/jiahongc) (Jiahong Chen).

## Install

The recommended path installs both the `hyatt-pp-cli` binary and the `pp-hyatt` agent skill (Claude Code, Codex, Cursor, Gemini CLI, GitHub Copilot, and other agents supported by the upstream [`skills`](https://github.com/vercel-labs/skills) CLI) in one shot:

```bash
npx -y @mvanhorn/printing-press-library install hyatt
```

For CLI only (no skill):

```bash
npx -y @mvanhorn/printing-press-library install hyatt --cli-only
```

For skill only — installs the skill into the same agents as the default command above, but skips the CLI binary (use this to update or reinstall just the skill):

```bash
npx -y @mvanhorn/printing-press-library install hyatt --skill-only
```

To constrain the skill install to one or more specific agents (repeatable — agent names match the [`skills`](https://github.com/vercel-labs/skills) CLI):

```bash
npx -y @mvanhorn/printing-press-library install hyatt --agent claude-code
npx -y @mvanhorn/printing-press-library install hyatt --agent claude-code --agent codex
```

### Without Node (Go fallback)

If `npx` isn't available (no Node, offline), install the CLI directly via Go (requires Go 1.26.4 or newer):

```bash
go install github.com/mvanhorn/printing-press-library/library/travel/hyatt/cmd/hyatt-pp-cli@latest
```

This installs the CLI only — no skill.

### Pre-built binary

Download a pre-built binary for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/hyatt-current). On macOS, clear the Gatekeeper quarantine: `xattr -d com.apple.quarantine <binary>`. On Unix, mark it executable: `chmod +x <binary>`.

<!-- pp-hermes-install-anchor -->
## Install for Hermes

Install the CLI binary first. The installer writes binaries to a per-user managed bin directory by default: `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows.

```bash
npx -y @mvanhorn/printing-press-library install hyatt --cli-only
```

Then install the focused Hermes skill.

From the Hermes CLI:

```bash
hermes skills install mvanhorn/printing-press-library/cli-skills/pp-hyatt --force
```

Inside a Hermes chat session:

```bash
/skills install mvanhorn/printing-press-library/cli-skills/pp-hyatt --force
```

Restart the Hermes session or gateway if the newly installed skill is not visible immediately.

## Install for OpenClaw
Install both the CLI binary and the focused OpenClaw skill. The installer defaults binaries to a per-user bin directory (`$HOME/.local/bin` on macOS/Linux, `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows):

```bash
npx -y @mvanhorn/printing-press-library install hyatt --agent openclaw
```

Restart the OpenClaw session or gateway if the newly installed skill is not visible immediately.

## Use with Claude Desktop

This CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle — Claude Desktop's standard format for one-click MCP extension installs (no JSON config required).

The bundle reuses your local browser session — set it up first if you haven't:

```bash
hyatt-pp-cli auth login --chrome
```

To install:

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/hyatt-current).
2. Double-click the `.mcpb` file. Claude Desktop opens and walks you through the install.

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS Apple Silicon (`darwin-arm64`) and Windows (`amd64`, `arm64`); for other platforms, use the manual config below.

<details>
<summary>Manual JSON config (advanced)</summary>

If you can't use the MCPB bundle (older Claude Desktop, unsupported platform), install the MCP binary and configure it manually.


```bash
go install github.com/mvanhorn/printing-press-library/library/travel/hyatt/cmd/hyatt-pp-mcp@latest
```

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "hyatt": {
      "command": "hyatt-pp-mcp"
    }
  }
}
```

</details>

## Authentication

The first shippable flow uses public Hyatt Points Calendar pages and browser-compatible replay. Logged-in Hyatt sessions are short-lived, so account-specific perks or durable authenticated availability are intentionally out of initial scope.

## Quick Start

```bash
# Check local configuration and transport assumptions without touching Hyatt.
hyatt-pp-cli doctor --dry-run

# Fetch one browser-sniffed Points Calendar surface and keep only fields an agent needs.
hyatt-pp-cli calendars get --spirit-code kulal --start-date 2026-09-01 --end-date 2026-09-02 --json --select spiritCode,days

# Preview Hyatt property metadata hydration before building the local hotel index.
hyatt-pp-cli hotels sync --dry-run

# Turn a city into the matching Hyatt hotels and the spirit codes used by availability scans.
hyatt-pp-cli hotels resolve-city --city "New York City" --json --select name,spiritCode,city,state,category

# Scan every matching city hotel and label standard-room versus other room-category award space.
hyatt-pp-cli scan city --city "New York City" --start 2026-09-01 --end 2026-09-07 --nights 2 --room-categories STANDARD_ROOM,SUITE --agent

# Find flexible-date award clusters across a shortlist.
hyatt-pp-cli awards density --hotels CHIRH,NYCUA,PARPH --start 2026-07-01 --end 2026-09-30 --nights 2 --bucket week --agent

```

## Unique Features

These capabilities aren't available in any other tool for this API.

### Award decision helpers
- **`awards certificate-fit`** — Find available Hyatt award nights that fit a Cat 1-4 or Cat 1-7 certificate before it expires.

  _Use this when the user needs to burn a Hyatt certificate without manually checking category and date fit._

  ```bash
  hyatt-pp-cli awards certificate-fit --cert cat1-4 --expires 2026-12-31 --start 2026-09-01 --end 2026-12-31 --agent
  ```
- **`awards offpeak`** — Find off-peak or unusually low-point Hyatt award clusters across synced calendar data.

  _Use this when the user wants low-points timing rather than a specific hotel._

  ```bash
  hyatt-pp-cli awards offpeak --country US --start 2026-06-01 --end 2027-05-31 --min-nights 2 --agent
  ```
- **`awards room-ladder`** — Compare standard, club, and suite award availability side by side for one Hyatt stay window.

  _Use this when the user cares whether better room categories are open, not just standard rooms._

  ```bash
  hyatt-pp-cli awards room-ladder --hotel CHIRH --start 2026-10-01 --end 2026-10-07 --agent
  ```

### Flexible-date search
- **`awards density`** — Show which dates or weeks have the most Hyatt award options across a hotel shortlist.

  _Use this before picking travel dates when flexibility matters more than a specific property._

  ```bash
  hyatt-pp-cli awards density --hotels CHIRH,NYCUA,PARPH --start 2026-07-01 --end 2026-09-30 --bucket week --agent
  ```
- **`awards split-stay`** — Build viable multi-hotel Hyatt itineraries when no one property has the full stay.

  _Use this when the user will switch hotels to make an award trip work._

  ```bash
  hyatt-pp-cli awards split-stay --hotels CHIRH,CHIJD --start 2026-08-01 --end 2026-08-15 --nights 5 --max-switches 1 --agent
  ```

### Watchlist intelligence
- **`watch volatility`** — Rank watched Hyatt hotels by how often award space opens, closes, or changes price.

  _Use this to decide which hard-to-book Hyatt searches deserve closer monitoring._

  ```bash
  hyatt-pp-cli watch volatility --since 30d --limit 20 --agent
  ```

### Reachability mitigation
- **`awards coverage`** — Show stale, missing, or uneven local Hyatt calendar coverage before trusting a scan.

  _Use this when browser protection or expired cookies may have left gaps in cached availability._

  ```bash
  hyatt-pp-cli awards coverage --hotels CHIRH,NYCUA --start 2026-07-01 --end 2026-12-31 --agent
  ```

## Recipes


### Check one Points Calendar page

```bash
hyatt-pp-cli calendars get --spirit-code kulal --start-date 2026-09-01 --end-date 2026-09-02 --room-category STANDARD_ROOM --agent --select spiritCode,roomCategory,days.2026-09-08.STANDARD_ROOM.pointsValue
```

Fetch one Hyatt standard-room calendar page and narrow the nested output to a specific award row.

### Resolve a city into Hyatt hotel codes

```bash
hyatt-pp-cli hotels resolve-city --city "New York City" --agent --select name,spiritCode,city,state,category
```

Find the hotels and spirit codes to use when the user only knows the city.

### Scan a city by room category

```bash
hyatt-pp-cli scan city --city "New York City" --start 2026-09-01 --end 2026-09-07 --nights 2 --room-categories STANDARD_ROOM,SUITE --agent --select hotelName,spiritCode,date,nights,roomCategory,isStandardRoom,pointsValue
```

Search every matching city hotel and keep room type plus length of stay explicit in the output.

### Find certificate-eligible awards

```bash
hyatt-pp-cli awards certificate-fit --cert cat1-4 --expires 2026-12-31 --start 2026-09-01 --end 2026-12-31 --agent
```

Combine local hotel categories and synced calendar rows to find certificate-fit nights.

### Compare flexible weekly availability

```bash
hyatt-pp-cli awards density --hotels CHIRH,NYCUA,PARPH --start 2026-07-01 --end 2026-09-30 --nights 2 --bucket week --agent
```

See which weeks have the most two-night award options across a hotel shortlist.

### Build a split-stay option

```bash
hyatt-pp-cli awards split-stay --hotels CHIRH,CHIJD --start 2026-08-01 --end 2026-08-15 --nights 5 --max-switches 1 --agent
```

Find a viable award itinerary across hotels when one property does not have all nights.

## Usage

Run `hyatt-pp-cli --help` for the full command reference and flag list.

## Commands

### calendars

Fetch Hyatt Points Calendar HTML pages

- **`hyatt-pp-cli calendars`** - Fetch a Hyatt Points Calendar page for a hotel spirit code

### hotels

Fetch Hyatt property metadata used to resolve city searches into hotel spirit codes

- **`hyatt-pp-cli hotels`** - Fetch Hyatt hotel metadata, including names, locations, categories, and spirit codes


## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
hyatt-pp-cli calendars --spirit-code example-value

# JSON for scripting and agents
hyatt-pp-cli calendars --spirit-code example-value --json

# Filter to specific fields
hyatt-pp-cli calendars --spirit-code example-value --json --select id,name,status

# Dry run — show the request without sending
hyatt-pp-cli calendars --spirit-code example-value --dry-run

# Agent mode — JSON + compact + no prompts in one flag
hyatt-pp-cli calendars --spirit-code example-value --agent
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
hyatt-pp-cli doctor
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

If you use agentcookie to sync secrets across machines, this CLI auto-adopts agentcookie-managed credentials with no extra setup. When the daemon writes to this CLI's config, `hyatt-pp-cli doctor` reports `agentcookie: detected` and `auth-status` labels the source as `agentcookie`. Skip this section if you don't use agentcookie - the CLI works the same as any other.

## Troubleshooting
**Authentication errors (exit code 4)**
- Run `hyatt-pp-cli doctor` to check credentials
- Verify the environment variable is set: `echo $HYATT_COOKIES`
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

### API-specific
- **Hyatt returns E6020, 403, or 429.** — Run `hyatt-pp-cli doctor hyatt` and refresh the browser-cookie path before scanning again.
- **Scan results look sparse or uneven.** — Run `hyatt-pp-cli awards coverage --hotels <codes> --start <date> --end <date>` to identify missing cached months or room categories.
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
- Generation hints: Pass this traffic-analysis file during generation so browser_clearance_http is preserved., Add a hand-authored parser that extracts the JavaScript assignment window.STORE = {...}; from HTML and emits normalized availability rows., Treat direct HTTP 403/429 as expected unless Chrome-cookie replay also fails.
- Candidate command ideas: calendar — Fetch and parse a Hyatt Points Calendar page for one hotel spirit code.; scan — Repeat calendar fetches across multiple spirit codes and date windows to find points availability.

Warnings from discovery:
- html-state-not-standard-json-script: The calendar payload is a JavaScript assignment, not script#__NEXT_DATA__; built-in embedded-json extraction may not parse it without hand code.

---

## Sources & Inspiration

This CLI was built by studying these projects and resources:

- [**Hyatt-award-search**](https://github.com/dewdream/Hyatt-award-search) — Python (8 stars)
- [**stayexpert-hyatt**](https://github.com/StayExpert/hyatt) — JavaScript (1 stars)
- [**hyattvalue**](https://github.com/sottenad/hyattvalue) — JavaScript

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
