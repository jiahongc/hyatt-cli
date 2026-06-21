# Hyatt Endpoint Smoke Test - 2026-06-21

## Summary

Result: working with browser fallback.

Hyatt rejects the raw HTTP calls with HTTP 403 browser protection, but the CLI now falls back to a headed `browser-use` session for Hyatt hotel metadata and rate-calendar pages. With that fallback installed and on `PATH`, live hotel metadata, city resolution, hotel scans, city scans, one-night searches, multi-night searches, standard rooms, and explicit suite categories all returned real Hyatt data.

## Environment

- Repo: `/Users/jiahongchen/Desktop/Coding/hyatt-cli`
- Binary: `./build/stage/bin/hyatt-cli`
- Base URL: `https://www.hyatt.com`
- Browser fallback: `browser-use`
- Default browser fallback session: `hyatt-cli`
- Disable fallback: `HYATT_BROWSER_FALLBACK=0`
- Tests: `go test ./...` passed
- Builds: `go build -o build/stage/bin/hyatt-cli ./cmd/hyatt-cli` passed; `go build -o build/stage/bin/hyatt-mcp ./cmd/hyatt-mcp` passed

## What Changed During This Run

- Fixed Hyatt cookie auth so saved cookies are sent as the raw `Cookie` header, not as `Authorization: Bearer ...`.
- Fixed `auth refresh` so failed validation clears invalid saved auth instead of leaving stale credentials.
- Fixed `doctor` so it detects `pycookiecheat` correctly.
- Fixed unsynced local Hyatt snapshots so local scans fail clearly instead of returning misleading empty arrays.
- Added browser fallback for Hyatt hotel metadata.
- Added browser fallback for Hyatt rate-calendar pages.
- Added live fallback to `hotels resolve-city`, `scan hotel`, and `scan city`.
- Normalized hotel metadata from Hyatt's keyed hotel object into rows with `name`, `spiritCode`, `city`, `state`, `country`, `category`, and `brand`.
- Fixed city matching so `New York City` also matches Hyatt rows whose city is `New York`.
- Fixed award rows so each row's `checkinDate` and `checkoutDate` correspond to that row date and the requested stay length.
- Fixed room-category filtering so `STANDARD_ROOM` calls do not leak suite/club rows, and suite calls do not duplicate standard-room rows.

## Live Results

### Hotel Metadata

Command:

```bash
PATH="$HOME/.local/bin:$PATH" ./build/stage/bin/hyatt-cli hotels \
  --json --no-input --no-color --yes --timeout 90s \
  --select name,spiritCode,city,state,country,category,brand
```

Observed:

- Exit: 0
- Rows: 2,979
- Metadata source: `browser`
- Metadata reason: `hyatt_browser_fallback`
- Sample rows included `ABDCC` Grand Hyatt Abu Dhabi, `AUHGH` Grand Hyatt Abu Dhabi, and multiple Mr & Mrs Smith rows.

### City Resolution - United States

Command:

```bash
PATH="$HOME/.local/bin:$PATH" ./build/stage/bin/hyatt-cli hotels resolve-city \
  --city "New York City" \
  --json --no-input --no-color --yes --timeout 90s \
  --select name,spiritCode,city,state,country,category,brand
```

Observed:

- Exit: 0
- Rows: 30
- Sample spirit codes: `NYCAM`, `NYCDD`, `NYCDM`, `LGATG`, `NYCUB`, `NYCUD`, `NYCCT`, `NYCTS`, `NYCAW`, `NYCGH`, `NYCHH`, `NYCXC`, `LGAZC`, `NYCZT`, `NYCZM`, `NYCRT`, `NYCUS`
- `New York City` now expands to rows whose Hyatt city is `New York`.

### City Resolution - Asia

Command:

```bash
PATH="$HOME/.local/bin:$PATH" ./build/stage/bin/hyatt-cli hotels resolve-city \
  --city "Kuala Lumpur" \
  --json --no-input --no-color --yes --timeout 90s \
  --select name,spiritCode,city,state,country,category,brand
```

Observed:

- Exit: 0
- Rows: 9
- Spirit codes: `KULAL`, `KUAGH`, `KULCT`, `KULXK`, `KULZK`, `KULRK`, `KULPH`, `M1721`, `M1783`

### Hotel Scan - One Night

Command:

```bash
PATH="$HOME/.local/bin:$PATH" ./build/stage/bin/hyatt-cli scan hotel \
  --hotels KULAL \
  --start 2026-09-01 --end 2026-09-05 \
  --nights 1 \
  --room-categories STANDARD_ROOM \
  --json --no-input --no-color --yes --timeout 120s \
  --select spiritCode,date,checkinDate,checkoutDate,nights,roomCategory,isStandardRoom,available,pointsValue,pointsLevel,source
```

Observed:

- Exit: 0
- Rows: 5
- Example row: `2026-09-01`, checkout `2026-09-02`, `STANDARD_ROOM`, `isStandardRoom: true`, `pointsValue: 7500`, `pointsLevel: OFF_PEAK`
- Later rows included `2026-09-04` and `2026-09-05` at `6000` points, `SUPER_OFF_PEAK`.

### Hotel Scan - Multi Night

Command:

```bash
PATH="$HOME/.local/bin:$PATH" ./build/stage/bin/hyatt-cli scan hotel \
  --hotels KULAL \
  --start 2026-09-01 --end 2026-09-05 \
  --nights 2 \
  --room-categories STANDARD_ROOM \
  --json --no-input --no-color --yes --timeout 120s \
  --select spiritCode,date,checkinDate,checkoutDate,nights,roomCategory,isStandardRoom,available,pointsValue,pointsLevel,source
```

Observed:

- Exit: 0
- Rows: 5
- Same date window as the one-night scan, but each row has `nights: 2`.
- Example row: `2026-09-01`, checkout `2026-09-03`, `STANDARD_ROOM`, `isStandardRoom: true`, `pointsValue: 7500`.

### Hotel Scan - Multiple Room Categories

Command:

```bash
PATH="$HOME/.local/bin:$PATH" ./build/stage/bin/hyatt-cli scan hotel \
  --hotels KULAL \
  --start 2026-09-01 --end 2026-09-03 \
  --nights 1 \
  --room-categories STANDARD_ROOM,SUITE \
  --json --no-input --no-color --yes --timeout 180s \
  --select spiritCode,date,nights,roomCategory,isStandardRoom,available,pointsValue,pointsLevel,source
```

Observed:

- Exit: 0
- Rows: 3
- Rows returned only for `STANDARD_ROOM`.
- No duplicate standard-room rows were emitted for the `SUITE` request.

### City Scan - Kuala Lumpur

Command:

```bash
PATH="$HOME/.local/bin:$PATH" ./build/stage/bin/hyatt-cli scan city \
  --city "Kuala Lumpur" \
  --start 2026-09-01 --end 2026-09-02 \
  --nights 1 \
  --room-categories STANDARD_ROOM \
  --json --no-input --no-color --yes --timeout 360s \
  --select spiritCode,date,nights,roomCategory,isStandardRoom,available,pointsValue,pointsLevel,source
```

Observed:

- Exit: 0
- Rows: 16
- Covered spirit codes included `KULAL`, `KUAGH`, `KULCT`, `KULXK`, `KULZK`, `KULRK`, `KULPH`, and `M1783`.
- Points examples: `KULCT` at 4,500, `KULAL` at 7,500, `KUAGH` at 12,000, `KULPH` at 20,000, `M1783` at 16,750.

### Calendar - United States, November 2026

Command:

```bash
PATH="$HOME/.local/bin:$PATH" ./build/stage/bin/hyatt-cli calendars \
  --spirit-code NYCAM \
  --start-date 2026-11-05 --end-date 2026-11-06 \
  --room-category STANDARD_ROOM \
  --json --no-input --no-color --yes --timeout 120s
```

Observed:

- Exit: 0
- Metadata source: `browser`
- Metadata reason: `hyatt_browser_fallback`
- `spiritCode: NYCAM`
- `nights: 1`
- `roomCategory: STANDARD_ROOM`
- Sample points: 45,000, `OFF_PEAK`

### Calendar - Asia, January 2027, Standard Room

Command:

```bash
PATH="$HOME/.local/bin:$PATH" ./build/stage/bin/hyatt-cli calendars \
  --spirit-code KUAGH \
  --start-date 2027-01-10 --end-date 2027-01-12 \
  --room-category STANDARD_ROOM \
  --json --no-input --no-color --yes --timeout 120s
```

Observed:

- Exit: 0
- Metadata source: `browser`
- `spiritCode: KUAGH`
- `nights: 2`
- Rows: 36
- Categories present: `STANDARD_ROOM` only
- Sample points: 12,000 to 20,000 depending on date and peak level.

### Calendar - Asia, January 2027, Suite

Command:

```bash
PATH="$HOME/.local/bin:$PATH" ./build/stage/bin/hyatt-cli calendars \
  --spirit-code KUAGH \
  --start-date 2027-01-10 --end-date 2027-01-12 \
  --room-category STANDARD_SUITE \
  --json --no-input --no-color --yes --timeout 120s
```

Observed:

- Exit: 0
- Metadata source: `browser`
- `spiritCode: KUAGH`
- `nights: 2`
- Rows: 10
- Categories present: `STANDARD_SUITE` only
- `isStandardRoom: false`
- Sample points: 20,000 to 25,500 depending on date and peak level.

## Remaining Caveats

- Raw HTTP still receives HTTP 403 from Hyatt. The working path is the browser fallback.
- `browser-use` must be installed and on `PATH`.
- The first fallback call opens a headed browser session.
- City scans are slower than direct API calls because each hotel/category calendar request goes through the browser fallback serially.
- For long date ranges, the current scan asks Hyatt for the requested start stay window and filters the returned calendar rows. A future improvement should page month-by-month when the requested range spans multiple Hyatt calendar windows.
- `--agent` implies compact output; for inspecting nested calendar rows, use `--json --no-input --no-color --yes` with `--select` instead of relying on compact defaults.
