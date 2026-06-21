# Hyatt Browser-Sniff Report

## User Goal Flow
- Goal: check World of Hyatt points availability across different Hyatt hotels.
- Steps completed:
  1. Probed Hyatt homepage and rate-calendar URL during discovery.
  2. Opened Park Hyatt Chicago rate-calendar URL with `agent-browser`; it hit Hyatt E6020 and produced no useful network capture.
  3. Opened the same URL with `browser-use --headed --profile Default`; it loaded the real Points Calendar page.
  4. Installed fetch/XHR interceptors and clicked month/option controls; only telemetry fired after interceptor install.
  5. Inspected page scripts and extracted `window.STORE`.
  6. Opened a positive sample calendar for `kulal` and extracted `window.STORE.days` with 37 date entries.
- Steps skipped:
  - Authenticated account flow. User reported Hyatt login sessions expire quickly; public/logged-out calendar data loaded for the sample hotel, so the first shippable surface is public calendar HTML plus possible Chrome-cookie replay for clearance.
- Coverage: 1 of 1 primary read-only calendar flow completed.

## Pages & Interactions
1. `https://www.hyatt.com/explore-hotels/rate-calendar?spiritCode=chihr&vrcEnabled=true`
   - Purpose: initial target property calendar.
   - Result: direct/agent-browser route hit E6020, browser-use route loaded calendar shell but `responseInfo` was `soldOutHotel`.
2. `https://www.hyatt.com/explore-hotels/rate-calendar?spiritCode=kulal&startDate=2026-09-01&endDate=2026-09-02&rooms=1&adults=1&kids=0&rate=Standard&vrcEnabled=true`
   - Purpose: positive availability sample.
   - Result: loaded `window.STORE` with 37 day entries. Example date rows contained `STANDARD_ROOM.pointsValue` and `STANDARD_ROOM.pointsLevel`.
3. Calendar controls clicked:
   - June, July, August month buttons.
   - Stay length, guests, and room type controls.
   - These updated UI state but did not fire first-party XHR after interceptors were installed.

## Browser-Sniff Configuration
- Backend used: `browser-use --headed --profile Default` for successful capture; `agent-browser --headed` failed on E6020.
- Fallbacks available in this session: browser-use, agent-browser, manual HAR. Chrome-MCP and computer-use were not exposed.
- Pacing: low-volume manual clicks, roughly one interaction every 1-3 seconds.
- Proxy pattern detection: not detected. No proxy-envelope XHR surface observed.

## Endpoints Discovered
| Method | Path | Status Code | Content-Type | Auth |
|---|---:|---:|---|---|
| GET | `/explore-hotels/rate-calendar` | 200 | `text/html` | public page, browser-clearance required from non-browser clients |

## Embedded State
- Global: `window.STORE`
- Key request fields: `spiritCode`, `startDate`, `endDate`, `numRooms`, `numAdults`, `numChildren`, `rate`, `roomCategory`.
- Key response fields: `responseInfo`, `days`.
- Availability shape: `days[date][roomCategory].pointsValue[]`, `days[date][roomCategory].pointsLevel`, optional `rate`.

## Replayability Verdict
- Replayable surface found: same-site HTML document with embedded JavaScript state.
- Runtime caveat: direct HTTP and Surf probes were blocked. Generated CLI should use browser-compatible HTTP plus Chrome cookie import where supported, then verify replay without keeping a browser sidecar alive.
- If Chrome-cookie replay cannot fetch this HTML outside browser-use, the run should hold or require manual HAR/browser-captured refresh rather than shipping resident-browser transport.
