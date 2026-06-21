# Hyatt CLI Brief

## API Identity
- Domain: World of Hyatt award-search and points-calendar availability across Hyatt properties.
- Users: points travelers with flexible dates, Hyatt loyalists, Globalist/milestone runners, and agents comparing points-vs-cash options.
- Data profile: hotel/property metadata, spiritCode identifiers, stay dates, room type, award availability, points price, cash rate when available, points-per-cent valuation, multi-night stay windows, and alert/watch history.

## Reachability Risk
- High. Hyatt public pages and the rate-calendar URL returned HTTP 403/429 to direct fetches. `probe-reachability` classified both `https://www.hyatt.com/` and a sample `rate-calendar` URL as `browser_clearance_http`, meaning both stdlib HTTP and Surf/Chrome-like transport received bot-protection signals.
- 403 body evidence: Hyatt error page says "Your browser did something unexpected and we were unable to process your request" and exposes error code `E6020`.
- Runtime implication: a shippable CLI must find a replayable HTTP/HTML/API surface through browser capture or manual HAR. If the only viable path is a live page-context browser session, this run should hold or pivot scope.
- User auth context: user can log in, but Hyatt has a short logged-in window and may log out quickly. Do not assume durable browser session state.

## Top Workflows
1. Check a list of Hyatt hotels for points availability across a flexible date range.
2. Search a destination/region for the cheapest available award nights by points price and category.
3. Detect multi-night award windows that are hidden when searching one night at a time.
4. Compare points price to cash rate and rank by cents-per-point value.
5. Watch hard-to-book properties and alert when standard room/suite award space opens.
6. Resolve a city, such as New York City, into all matching Hyatt hotels and their `spiritCode`s before scanning availability.
7. Separate standard-room award availability from other room categories so users can tell exactly what type of award is open.
8. Treat length of stay as an availability dimension because Hyatt can show different award space for 1 night versus multiple nights.

## Table Stakes
- Hyatt first-party: per-property Points Calendar, month navigation, length-of-stay, guests, room type, no-availability days.
- Rooms.aero: program-wide Hyatt search, filters by nights, brand, category, country, max points, cpp, hotel pages, maps, and alerts.
- MaxMyPoint / BurnMyPoints: searchable Hyatt award calendars and alerts; BurnMyPoints specifically highlights last-minute Hyatt availability, high cpp stays, and hidden multi-night stays.

## Data Layer
- Primary entities: hotels, search windows, calendar days, room award entries, cash-rate snapshots, watch alerts.
- Sync cursor: per hotel spiritCode + month + room type + occupancy + length-of-stay.
- FTS/search: hotel names, destination/country/state, brand, category, known hard-to-book properties.

## User Vision
- Build a CLI to get Hyatt availability with points for different Hyatt hotels.
- The user can log in, but Hyatt sessions are short-lived. Browser discovery should be treated as temporary capture, not normal runtime.
- City input should work even when the user does not know Hyatt `spiritCode`s; the CLI should resolve the city to matching hotels, show the codes, and then scan those hotels.
- Room type must be explicit. Standard-room availability should be labeled separately from other Hyatt room categories such as club or suite awards when present.
- The 1-night / multi-night picker changes availability. Scan output and cached snapshots must include `nights`, `checkinDate`, and `checkoutDate`, and multi-night availability must be verified with the actual length-of-stay request.

## Product Thesis
- Name: Hyatt Award Scout
- Why it should exist: Hyatt's own calendar is property-by-property and browser-bound; third-party tools are broad but not agent-native. A CLI can make flexible-date, multi-property award checks scriptable, comparable, stored locally, and alertable.

## Build Priorities
1. Discover a replayable Hyatt rate-calendar / availability contract through browser capture or manual HAR.
2. Generate commands for hotel lookup by spiritCode, points calendar, date-range availability, city-wide scans, explicit room-category checks, and length-of-stay-aware multi-hotel comparisons.
3. Persist snapshots locally so agents can compare stale/fresh availability and calculate cpp against cash-rate fields when present.
4. Add novel commands for multi-night window detection, category/certificate fit, and watchlist delta alerts.

## Evidence Notes
- Hyatt official redemption page: free nights start at 3,000 points and can be redeemed at over 1,000 hotels.
- Hyatt official free nights/upgrades page: standard rooms have no blackout dates, but points/cash and suites are subject to availability; published category chart defines point bands.
- TPG July 17, 2025: Hyatt launched a points calendar showing when award nights are and are not available; users can adjust dates, length of stay, guests, and room type.
- TPG Jan. 30, 2026: calendar is reached after an award search via View Rates -> Points Calendar; it displays award rates and no-award-space days.
- Traveling for Miles July 18, 2025: new availability calendar shows a year of award availability at most properties, but not all properties and some bugs remain.
- Frequent Miler Sept. 4, 2023: Rooms.aero provided program-wide Hyatt award search with filters for nights, brand, category, country, points cost, cpp, property calendars, and alerts.
- Reddit user pain: flexible users want date-agnostic multi-hotel point availability instead of clicking hotel-by-hotel/date-by-date.
