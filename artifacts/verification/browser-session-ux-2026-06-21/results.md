# Browser Session UX Verification - 2026-06-21

## Question

Can Hyatt availability searches run without repeatedly opening and closing visible browser windows?

## Findings

- Raw HTTP still receives Hyatt 403 protection, so direct fetches are not enough.
- Basic headless `browser-use` navigation loads Hyatt's "browser did something unexpected" page for the calendar surface and does not expose `window.STORE`.
- Reusing one headed `browser-use` session works.
- Navigating the existing tab with `window.location.href = <url>` works and exposes `window.STORE` for the next hotel.
- On macOS, minimizing matching Hyatt Chrome windows with AppleScript still allows `browser-use eval` to read `window.STORE`.
- `browser-use python` can navigate and read `browser.html` in one process. The CLI uses that faster path before falling back to repeated `eval` polling.
- Hyatt hotel metadata from `/explore-hotels/service/hotels` changes slowly and is now cached locally, so default `auto` mode does not need to open that page on every city resolution or `hotels` call.

## Change

The CLI now:

- starts a headed browser only when the named session is not already running;
- reuses the existing `HYATT_BROWSER_SESSION` session, default `hyatt-cli`;
- navigates the existing tab between Hyatt URLs with JavaScript;
- does not close/reopen the session for normal multi-hotel scans;
- minimizes Hyatt Chrome windows on macOS after navigation by default;
- allows `HYATT_BROWSER_BACKGROUND=0` to leave the browser visible;
- extracts page HTML through one `browser-use python` call where possible;
- caches normalized hotel metadata for `HYATT_HOTELS_CACHE_MAX_AGE`, default `24h`;
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
- With backgrounding enabled, the scan still returned rows after the CLI minimized the Hyatt Chrome window.
- Cold two-hotel scan timing after Python extraction path: about 5.3 seconds.
- Warm-session two-hotel scan timing: about 1.2 seconds.

## UX Caveat

The browser is still a real headed browser because Hyatt blocked the basic headless path. The UX improvement is one persistent browser session/tab plus macOS minimization after navigation, instead of many windows opening and closing.
