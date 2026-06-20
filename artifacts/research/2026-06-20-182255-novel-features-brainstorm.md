# Hyatt Novel Features Brainstorm

## Customer model

**Flexible points traveler**

Today (without this CLI): They open Hyatt property pages one by one, click into Points Calendar, change dates, room type, guests, and length of stay, then mentally compare scattered results. They may also check Rooms.aero or BurnMyPoints, but those results are web-table oriented rather than scriptable.

Weekly ritual: They rerun broad destination searches whenever travel dates are flexible, looking for which hotel/date combination actually has standard-room award space.

Frustration: They cannot quickly answer "which dates have the most options across my shortlist?" without clicking hotel-by-hotel and date-by-date.

**Hard-to-book Hyatt watcher**

Today (without this CLI): They monitor aspirational or constrained properties manually, sometimes refreshing saved searches and hoping award space appears. Alerts exist in third-party products, but local watch history and why availability changed are opaque.

Weekly ritual: They check the same hotel/date windows repeatedly, especially for suites, club rooms, or peak properties.

Frustration: They can see a new opening, but not whether a property is generally volatile, stale, or worth watching more aggressively.

**Certificate and points optimizer**

Today (without this CLI): They compare Hyatt category charts, certificate rules, point prices, cash rates, and room availability across multiple tabs. Category filters and max-points filters help, but they still have to reason manually about free-night certificates and off-peak opportunities.

Weekly ritual: They look for trips that spend the fewest points or burn an expiring certificate without wasting it.

Frustration: Hyatt data is available, but the decision layer is missing: "is this a good use of my certificate or points right now?"

**Travel-planning agent**

Today (without this CLI): An agent can fetch structured calendar rows, but it still has to stitch together local hotel metadata, room categories, point bands, cash rates, stale snapshots, and watch deltas itself.

Weekly ritual: It prepares repeatable award-search reports for a destination, shortlist, or travel window.

Frustration: Raw availability rows are too granular; the agent needs higher-level, locally verifiable commands that explain coverage, tradeoffs, and next actions.

## Candidates (pre-cut)

| # | Name | Command | Description | Persona served | Source | Pre-cut disposition | Long Description |
|---|------|---------|-------------|----------------|--------|---------------------|------------------|
| 1 | Certificate fit finder | `awards certificate-fit` | Find award nights that fit a Cat 1-4 or Cat 1-7 certificate before an expiration date. | Certificate and points optimizer | user briefing + service-specific | keep | Use this command for free-night certificate fit. Do NOT use it for generic points-price filtering; use `scan region` instead. |
| 2 | Flexible density matrix | `awards density` | Summarize which weeks or dates have the most award options across a hotel shortlist. | Flexible points traveler | persona-driven | keep | Use this command for aggregate availability density across hotels. Do NOT use it for exact bookable stay windows; use `awards windows` or `awards split-stay` instead. |
| 3 | Split-stay builder | `awards split-stay` | Build viable multi-hotel itineraries when no single hotel has the full consecutive stay. | Flexible points traveler | cross-entity local query | keep | Use this command when the user accepts switching hotels. Do NOT use it for single-property consecutive stays; use `awards windows` instead. |
| 4 | Watch volatility ranking | `watch volatility` | Rank watched hotels by how often award availability opens, closes, or changes price. | Hard-to-book Hyatt watcher | cross-entity local query | keep | Use this command for historical watch churn. Do NOT use it for the latest delta run; use `watch run` instead. |
| 5 | Off-peak opportunity finder | `awards offpeak` | Find off-peak or unusually low-point award clusters across synced calendar data. | Certificate and points optimizer | service-specific | keep | Use this command for low-points timing. Do NOT use it for cash-rate value ranking; use `awards value` instead. |
| 6 | Room-category ladder | `awards room-ladder` | Compare standard, club, and suite award availability side by side for the same hotel/date window. | Hard-to-book Hyatt watcher | service-specific | keep | Use this command to compare room award categories. Do NOT use it for searching one room category broadly; use `scan hotel` or `scan region` instead. |
| 7 | Snapshot coverage audit | `awards coverage` | Show stale, missing, or uneven local calendar coverage before relying on scan results. | Travel-planning agent | cross-entity local query | keep | none |
| 8 | Award anomaly audit | `awards anomalies` | Flag parsed calendar rows with contradictory price, level, or room-category state. | Travel-planning agent | cross-entity local query | borderline, cut | none |
| 9 | Personalized elite-benefit planner | `awards elite-plan` | Use logged-in Hyatt status and account data to rank stays by perks and milestone progress. | Hard-to-book Hyatt watcher | user briefing | cut: auth gap | none |
| 10 | Push/SMS availability alerts | `watch notify` | Send phone or email alerts when watched award space opens. | Hard-to-book Hyatt watcher | competitor pattern | cut: external service / persistent process | none |
| 11 | Natural-language trip summarizer | `awards explain` | Summarize best redemption choices in prose. | Travel-planning agent | persona-driven | cut: LLM dependency | none |
| 12 | Rooms.aero parity checker | `awards compare-third-party` | Compare local Hyatt results against third-party award-search sites. | Flexible points traveler | competitor pattern | cut: external service | none |
| 13 | Map-distance optimizer | `awards map-fit` | Rank award hotels by distance to a venue or neighborhood. | Flexible points traveler | persona-driven | cut: external geocoding / unclear metadata | none |
| 14 | Calendar screenshot archiver | `calendars archive-html` | Save visual Hyatt calendar screenshots alongside parsed data. | Travel-planning agent | service-specific | cut: scope creep, weak user value | none |

## Survivors and kills

### Survivors

| # | Feature | Command | Score | Buildability | How It Works | Evidence | Long Description |
|---|---------|---------|-------|--------------|--------------|----------|------------------|
| 1 | Certificate fit finder | `awards certificate-fit --cert cat1-4 --expires 2026-12-31 --start 2026-09-01 --end 2026-12-31` | 8/10 | hand-code | This uses local hotel category metadata plus synced Hyatt `days[date][roomCategory].pointsValue/pointsLevel/rate` calendar data to return available award nights eligible for the selected certificate with no external dependencies. | Brief notes Hyatt category chart point bands, free-night redemption rules, and users comparing points-vs-cash; absorb manifest already covers category and max-points filters but not certificate expiration fit. | Use this command for free-night certificate fit. Do NOT use it for generic points-price filtering; use `scan region` instead. |
| 2 | Flexible density matrix | `awards density --hotels CHIRH,NYCUA,PARPH --start 2026-07-01 --end 2026-09-30 --bucket week` | 8/10 | hand-code | This uses local calendar-day availability rows grouped by hotel and date bucket to compute option counts, minimum points, and no-award density with no external dependencies. | Brief cites Reddit pain around date-agnostic multi-hotel point availability; top workflow is checking a list of Hyatt hotels across flexible date ranges. | Use this command for aggregate availability density across hotels. Do NOT use it for exact bookable stay windows; use `awards windows` or `awards split-stay` instead. |
| 3 | Split-stay builder | `awards split-stay --hotels CHIRH,CHIJD --start 2026-08-01 --end 2026-08-15 --nights 5 --max-switches 1` | 8/10 | hand-code | This uses local per-night availability across multiple hotels to compute date-contiguous itineraries, allowing bounded hotel switches, with no external dependencies. | Brief lists hidden multi-night availability as a core workflow and Reddit pain around flexible multi-hotel searching; absorb manifest covers single-hotel consecutive windows, not cross-hotel stay construction. | Use this command when the user accepts switching hotels. Do NOT use it for single-property consecutive stays; use `awards windows` instead. |
| 4 | Watch volatility ranking | `watch volatility --since 30d --limit 20` | 7/10 | hand-code | This uses local watch snapshots and delta history to count openings, closures, point changes, and room-category changes per hotel/date window with no external dependencies. | Absorb manifest covers `watch run` deltas; brief identifies watch alerts for hard-to-book properties as a top workflow and local watch history as a primary data entity. | Use this command for historical watch churn. Do NOT use it for the latest delta run; use `watch run` instead. |
| 5 | Off-peak opportunity finder | `awards offpeak --country US --start 2026-06-01 --end 2027-05-31 --min-nights 2` | 7/10 | hand-code | This uses local hotel category metadata and synced calendar `pointsLevel`/`pointsValue` fields to find off-peak or lowest-band available clusters with no external dependencies. | Brief notes Hyatt category chart point bands and Points Calendar day-level award rates; optimizer users compare point prices and want high-value redemption timing. | Use this command for low-points timing. Do NOT use it for cash-rate value ranking; use `awards value` instead. |
| 6 | Room-category ladder | `awards room-ladder --hotel CHIRH --start 2026-10-01 --end 2026-10-07` | 7/10 | hand-code | This uses locally cached calendar rows for multiple `roomCategory` values at the same hotel/date range to compare standard, club, and suite award availability with no external dependencies. | API summary says `window.STORE.days[date][roomCategory]` carries points fields; brief notes Hyatt calendar supports room type changes and suites are subject to availability. | Use this command to compare room award categories. Do NOT use it for searching one room category broadly; use `scan hotel` or `scan region` instead. |
| 7 | Snapshot coverage audit | `awards coverage --hotels CHIRH,NYCUA --start 2026-07-01 --end 2026-12-31` | 6/10 | hand-code | This uses local snapshot metadata, hotel metadata, and requested date windows to report missing months, stale captures, and room-category coverage gaps with no external dependencies. | Brief says direct HTTP is bot-protected and runtime should rely on public calendar data plus optional Chrome-cookie diagnostics; local snapshots are a build priority and agents need reliable cached data. | none |

### Killed candidates

| # | Candidate | Reason killed |
|---|-----------|---------------|
| 8 | Award anomaly audit | Score fell below survivor cutoff: useful for debugging parser quality, but weaker customer pain than coverage, density, or certificate fit. |
| 9 | Personalized elite-benefit planner | Requires durable logged-in Hyatt account/session data, which the API summary explicitly says is not durable enough for initial runtime. |
| 10 | Push/SMS availability alerts | Requires external notification services or persistent background infrastructure; descope remains covered by local `watch run` and `watch volatility`. |
| 11 | Natural-language trip summarizer | Requires LLM summarization; mechanical alternatives are `awards density`, `awards split-stay`, and `awards value`. |
| 12 | Rooms.aero parity checker | Requires scraping or calling a third-party service outside the Hyatt spec and brief. |
| 13 | Map-distance optimizer | Requires reliable geocoding or venue-distance data not established in the API spec summary. |
| 14 | Calendar screenshot archiver | More of an archival/browser artifact than a customer workflow; higher complexity without enough incremental value over structured calendar snapshots. |
