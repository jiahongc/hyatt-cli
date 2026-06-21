# Browser Session UX Verification - 2026-06-21

## Question

Can Hyatt availability searches run without repeatedly opening and closing visible browser windows?

## Findings

- Raw HTTP still receives Hyatt 403 protection, so direct fetches are not enough.
- Basic headless `browser-use` navigation loads Hyatt's "browser did something unexpected" page for the calendar surface and does not expose `window.STORE`.
- Reusing one headed `browser-use` session works.
- Navigating the existing tab with `window.location.href = <url>` works and exposes `window.STORE` for the next hotel.

## Change

The CLI now:

- starts a headed browser only when the named session is not already running;
- reuses the existing `HYATT_BROWSER_SESSION` session, default `hyatt-cli`;
- navigates the existing tab between Hyatt URLs with JavaScript;
- does not close/reopen the session for normal multi-hotel scans;
- supports `HYATT_BROWSER_HEADLESS=true` as an opt-in experiment, but not as the default.

## Live Check

Command:

```bash
PATH="$HOME/.local/bin:$PATH" ./build/stage/bin/hyatt-cli scan hotel \
  --hotels KULAL,KUAGH \
  --start 2026-09-01 \
  --end 2026-09-01 \
  --nights 1 \
  --room-categories STANDARD_ROOM \
  --json --no-input --no-color --yes \
  --timeout 180s \
  --select spiritCode,date,nights,roomCategory,isStandardRoom,available,pointsValue
```

Observed:

- Exit: 0
- Rows: 2
- `KULAL` returned 7,500 points for 2026-09-01.
- `KUAGH` returned 12,000 points for 2026-09-01.
- `browser-use --session hyatt-cli tab list` showed one tab, at the final `KUAGH` calendar URL.

## UX Caveat

The browser is still visible by default because Hyatt blocked the basic headless path. The improvement is that it should be one persistent browser session/tab instead of many windows opening and closing.
