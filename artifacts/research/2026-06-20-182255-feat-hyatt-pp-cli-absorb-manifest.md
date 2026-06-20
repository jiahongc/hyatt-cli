# Absorb Manifest

## Sources Reviewed

- Hyatt first-party Points Calendar and redemption pages.
- Hyatt browser-sniffed `window.STORE` embedded state from `/explore-hotels/rate-calendar`.
- Rooms.aero Hyatt search/calendar/alerts feature surface.
- MaxMyPoint and BurnMyPoints award-search/alert patterns.
- GitHub: `dewdream/Hyatt-award-search`, `sottenad/hyattvalue`, and `StayExpert/hyatt`.
- Community pain from Reddit/travel-hacking discussion around date-agnostic and multi-hotel Hyatt award search.

## Absorbed (match or beat everything that exists)

| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-------------|-------------------|-------------|
| 1 | Fetch one Hyatt Points Calendar page for a hotel spiritCode/date/occupancy/room category | Hyatt Points Calendar + browser sniff | (generated endpoint) calendars get | Agent-readable JSON/selection over Hyatt's embedded `window.STORE` instead of browser-only viewing |
| 2 | Parse calendar-day points availability and no-award days | Hyatt `window.STORE.days` | (behavior in hyatt-pp-cli calendars get) parse days into date, room category, points value, points level, availability, `isStandardRoom`, and `nights` | Stable structured output from embedded page state, with standard-room awards and stay length clearly separated from other room categories/searches |
| 3 | Search one hotel across a date range and consecutive-night length | `dewdream/Hyatt-award-search` + user length-of-stay requirement | hyatt-pp-cli scan hotel | Headless/scriptable, JSON/CSV/agent output, local snapshot support, and explicit `--nights` fan-out |
| 4 | Validate hotel spiritCode input and date windows before scanning | `dewdream/Hyatt-award-search` | (behavior in hyatt-pp-cli scan hotel) input validation | Clear CLI errors instead of GUI popups |
| 5 | Hydrate Hyatt property metadata from `/explore-hotels/service/hotels` | `sottenad/hyattvalue` + `StayExpert/hyatt` | hyatt-pp-cli hotels sync | SQLite/FTS property index; no Mongo/Postgres setup |
| 6 | Search Hyatt hotels by country/city/state/brand/category/name and show spirit codes | Rooms.aero Hotels search + `StayExpert/hyatt` + user requirement | hyatt-pp-cli hotels search | Offline FTS, `--json`, `--select`, repeatable scripts, and visible `spiritCode` output so users do not need to know codes upfront |
| 7 | Resolve city input into all matching Hyatt hotels and spirit codes | User requirement + Hyatt `/explore-hotels/service/hotels` metadata | hyatt-pp-cli hotels resolve-city | Turns input like `New York City` into the actual hotel list and codes the scanner will use |
| 8 | Scan all Hyatt hotels in a city without manually providing spirit codes | User requirement + Rooms.aero city/destination search pattern | hyatt-pp-cli scan city | Pulls the matching city hotel list, fans out across each `spiritCode`, and reports which hotels have points availability for the requested room categories and length of stay |
| 9 | Program-wide Hyatt award search across regions | Rooms.aero Explore | hyatt-pp-cli scan region | Agent-native filters and local snapshots instead of a web table |
| 10 | Filter by nights, brand, category, country, maximum points, and room category | Rooms.aero filters | (behavior in hyatt-pp-cli scan region) filter flags | Composable CLI filters with JSON/CSV output |
| 11 | Search standard rooms and/or other room categories explicitly | Hyatt Points Calendar room type control + user requirement | (behavior in hyatt-pp-cli scan hotel, scan city, scan region) `--room-categories` accepts `STANDARD_ROOM` and other categories | Users can tell whether availability is standard-room award space or another room type before trying to book |
| 12 | Treat length of stay as an availability-changing input | Hyatt length-of-stay picker + user requirement | (behavior in hyatt-pp-cli scan hotel, scan city, scan region, awards windows) `--nights` and derived checkout date are part of the cache key and output | Prevents 1-night availability from being misreported as multi-night availability |
| 13 | Include cash rate when present and compute cents-per-point value | Rooms.aero cpp filters + `sottenad/hyattvalue` pricing | (behavior in hyatt-pp-cli scan region) cpp calculation | Local valuation sorting and custom cpp threshold flags |
| 14 | Detect multi-night stays and hidden consecutive-night availability | Rooms.aero multi-night stays + Reddit/BurnMyPoints user pain + user length-of-stay requirement | hyatt-pp-cli awards windows | Finds viable stay windows using actual `--nights`/checkout requests rather than assuming adjacent 1-night results imply a multi-night stay |
| 15 | Watch specific hotels/dates and report availability deltas | Rooms.aero alerts + BurnMyPoints alert pattern | hyatt-pp-cli watch run | Local watchlist/delta output without requiring hosted alerts, keyed by room category and nights |
| 16 | Last-minute Hyatt availability search | BurnMyPoints last-minute Hyatt availability pattern | hyatt-pp-cli awards last-minute | Sortable local scan over upcoming windows and properties |
| 17 | Identify high-value redemptions by cpp | BurnMyPoints deals pattern + Rooms.aero cpp | hyatt-pp-cli awards value | Combines points, cash, length of stay, and cached history into ranked output |
| 18 | Explain bot-protection/session status and cookie freshness | Browser-sniff reachability evidence | hyatt-pp-cli doctor hyatt | Clear diagnostics for E6020/403/429/expired browser-cookie cases |

## Transcendence (only possible with our approach)

| # | Feature | Command | Score | Buildability | How It Works | Evidence | Long Description |
|---|---------|---------|-------|--------------|--------------|----------|------------------|
| 1 | Certificate fit finder | `awards certificate-fit --cert cat1-4 --expires 2026-12-31 --start 2026-09-01 --end 2026-12-31` | 8/10 | hand-code | This uses local hotel category metadata plus synced Hyatt `days[date][roomCategory].pointsValue/pointsLevel/rate` calendar data to return available award nights eligible for the selected certificate with no external dependencies. | Brief notes Hyatt category chart point bands, free-night redemption rules, and users comparing points-vs-cash; absorbed features cover category and max-points filters but not certificate expiration fit. | Use this command for free-night certificate fit. Do NOT use it for generic points-price filtering; use `scan region` instead. |
| 2 | Flexible density matrix | `awards density --hotels CHIRH,NYCUA,PARPH --start 2026-07-01 --end 2026-09-30 --nights 2 --bucket week` | 8/10 | hand-code | This uses local calendar availability rows grouped by hotel, date bucket, room category, and length of stay to compute option counts, minimum points, and no-award density with no external dependencies. | Brief cites Reddit pain around date-agnostic multi-hotel point availability; user clarified the 1-night/multi-night picker changes availability. | Use this command for aggregate availability density across hotels. Do NOT use it for exact bookable stay windows; use `awards windows` or `awards split-stay` instead. |
| 3 | Split-stay builder | `awards split-stay --hotels CHIRH,CHIJD --start 2026-08-01 --end 2026-08-15 --nights 5 --max-switches 1` | 8/10 | hand-code | This uses local length-of-stay-aware availability across multiple hotels to compute date-contiguous itineraries, allowing bounded hotel switches, with no external dependencies. | Brief lists hidden multi-night availability as a core workflow; user clarified multi-night requests must be checked as multi-night availability, not inferred from one-night results. | Use this command when the user accepts switching hotels. Do NOT use it for single-property consecutive stays; use `awards windows` instead. |
| 4 | Watch volatility ranking | `watch volatility --since 30d --limit 20` | 7/10 | hand-code | This uses local watch snapshots and delta history to count openings, closures, point changes, and room-category changes per hotel/date window with no external dependencies. | Absorbed features cover `watch run` deltas; brief identifies watch alerts for hard-to-book properties as a top workflow and local watch history as a primary data entity. | Use this command for historical watch churn. Do NOT use it for the latest delta run; use `watch run` instead. |
| 5 | Off-peak opportunity finder | `awards offpeak --country US --start 2026-06-01 --end 2027-05-31 --min-nights 2` | 7/10 | hand-code | This uses local hotel category metadata and synced calendar `pointsLevel`/`pointsValue` fields to find off-peak or lowest-band available clusters with no external dependencies. | Brief notes Hyatt category chart point bands and Points Calendar day-level award rates; optimizer users compare point prices and want high-value redemption timing. | Use this command for low-points timing. Do NOT use it for cash-rate value ranking; use `awards value` instead. |
| 6 | Room-category ladder | `awards room-ladder --hotel CHIRH --start 2026-10-01 --end 2026-10-07` | 7/10 | hand-code | This uses locally cached calendar rows for multiple `roomCategory` values at the same hotel/date range to compare standard, club, and suite award availability with no external dependencies. | API summary says `window.STORE.days[date][roomCategory]` carries points fields; brief notes Hyatt calendar supports room type changes and suites are subject to availability. | Use this command to compare room award categories. Do NOT use it for searching one room category broadly; use `scan hotel` or `scan region` instead. |
| 7 | Snapshot coverage audit | `awards coverage --hotels CHIRH,NYCUA --start 2026-07-01 --end 2026-12-31` | 6/10 | hand-code | This uses local snapshot metadata, hotel metadata, and requested date windows to report missing months, stale captures, and room-category coverage gaps with no external dependencies. | Brief says direct HTTP is bot-protected and runtime should rely on public calendar data plus optional Chrome-cookie diagnostics; local snapshots are a build priority and agents need reliable cached data. | none |

## Cut Novel Candidates

| # | Candidate | Reason killed |
|---|-----------|---------------|
| 8 | Award anomaly audit | Score fell below survivor cutoff: useful for debugging parser quality, but weaker customer pain than coverage, density, or certificate fit. |
| 9 | Personalized elite-benefit planner | Requires durable logged-in Hyatt account/session data, which the API summary explicitly says is not durable enough for initial runtime. |
| 10 | Push/SMS availability alerts | Requires external notification services or persistent background infrastructure; descope remains covered by local `watch run` and `watch volatility`. |
| 11 | Natural-language trip summarizer | Requires LLM summarization; mechanical alternatives are `awards density`, `awards split-stay`, and `awards value`. |
| 12 | Rooms.aero parity checker | Requires scraping or calling a third-party service outside the Hyatt spec and brief. |
| 13 | Map-distance optimizer | Requires reliable geocoding or venue-distance data not established in the API spec summary. |
| 14 | Calendar screenshot archiver | More of an archival/browser artifact than a customer workflow; higher complexity without enough incremental value over structured calendar snapshots. |

## Stub and Risk Notes

- No approved rows are planned as stubs.
- Primary risk: Hyatt direct HTTP currently returns 403/429/E6020 without a real browser profile. Generation must keep the browser-clearance/cookie diagnostic path explicit and avoid pretending logged-in sessions are durable.
- All seven transcendence rows are `hand-code`, so approval commits the build to seven post-generation Cobra commands plus root wiring and tests. The absorbed scope also now explicitly includes city-to-`spiritCode` resolution, `scan city`, room-category-aware scan output, and length-of-stay-aware cache/output behavior.
