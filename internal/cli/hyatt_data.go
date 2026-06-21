package cli

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jiahongc/hyatt-cli/internal/store"
	"github.com/spf13/cobra"
)

const defaultHyattHotelsCacheMaxAge = 24 * time.Hour

type hyattHotel struct {
	Name       string `json:"name,omitempty"`
	SpiritCode string `json:"spiritCode,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	Country    string `json:"country,omitempty"`
	Category   int    `json:"category,omitempty"`
	Brand      string `json:"brand,omitempty"`
}

type hyattAwardRow struct {
	HotelName      string `json:"hotelName,omitempty"`
	SpiritCode     string `json:"spiritCode,omitempty"`
	City           string `json:"city,omitempty"`
	State          string `json:"state,omitempty"`
	Country        string `json:"country,omitempty"`
	Category       int    `json:"category,omitempty"`
	CheckinDate    string `json:"checkinDate,omitempty"`
	CheckoutDate   string `json:"checkoutDate,omitempty"`
	Date           string `json:"date,omitempty"`
	Nights         int    `json:"nights,omitempty"`
	RoomCategory   string `json:"roomCategory,omitempty"`
	IsStandardRoom bool   `json:"isStandardRoom"`
	Available      bool   `json:"available"`
	PointsValue    int    `json:"pointsValue,omitempty"`
	PointsLevel    string `json:"pointsLevel,omitempty"`
	CashRate       string `json:"cashRate,omitempty"`
	Source         string `json:"source,omitempty"`
}

type hyattCalendarOutput struct {
	SpiritCode   string          `json:"spiritCode,omitempty"`
	CheckinDate  string          `json:"checkinDate,omitempty"`
	CheckoutDate string          `json:"checkoutDate,omitempty"`
	Nights       int             `json:"nights,omitempty"`
	RoomCategory string          `json:"roomCategory,omitempty"`
	Days         []hyattAwardRow `json:"days"`
}

var hyattStoreAssignRE = regexp.MustCompile(`window\.STORE\s*=`)

func enhanceHyattCalendarData(raw []byte, params map[string]string) (json.RawMessage, bool) {
	rows, meta, ok := hyattRowsFromPayload(raw, params)
	if !ok {
		return nil, false
	}
	if meta["roomCategory"] != "" {
		rows = filterAwardRows(rows, meta["spiritCode"], "", "", "", atoiDefault(meta["nights"], 0), meta["roomCategory"])
	}
	out := hyattCalendarOutput{
		SpiritCode:   meta["spiritCode"],
		CheckinDate:  meta["checkinDate"],
		CheckoutDate: meta["checkoutDate"],
		RoomCategory: meta["roomCategory"],
		Nights:       atoiDefault(meta["nights"], 0),
		Days:         rows,
	}
	data, err := json.Marshal(out)
	return data, err == nil
}

func hyattRowsFromPayload(raw []byte, params map[string]string) ([]hyattAwardRow, map[string]string, bool) {
	var value any
	if json.Unmarshal(raw, &value) != nil {
		storeObj, ok := extractHyattStoreFromHTML(raw)
		if !ok {
			return nil, nil, false
		}
		value = storeObj
	}
	obj, ok := value.(map[string]any)
	if !ok {
		return nil, nil, false
	}
	if wrapped, ok := obj["store"].(map[string]any); ok {
		obj = wrapped
	}
	if _, hasDays := obj["days"]; hasDays {
		rows := rowsFromHyattStore(obj, params)
		return rows, calendarMeta(obj, params), true
	}
	if row, ok := rowFromFlatCalendarObject(obj); ok {
		return []hyattAwardRow{row}, map[string]string{
			"spiritCode":   row.SpiritCode,
			"checkinDate":  row.CheckinDate,
			"checkoutDate": row.CheckoutDate,
			"roomCategory": row.RoomCategory,
			"nights":       strconv.Itoa(row.Nights),
		}, true
	}
	return nil, nil, false
}

func extractHyattStoreFromHTML(raw []byte) (map[string]any, bool) {
	loc := hyattStoreAssignRE.FindIndex(raw)
	if loc == nil {
		return nil, false
	}
	start := bytes.IndexByte(raw[loc[1]:], '{')
	if start < 0 {
		return nil, false
	}
	start += loc[1]
	end := matchingJSONObjectEnd(raw, start)
	if end <= start {
		return nil, false
	}
	var obj map[string]any
	if err := json.Unmarshal(raw[start:end], &obj); err != nil {
		return nil, false
	}
	return obj, true
}

func matchingJSONObjectEnd(raw []byte, start int) int {
	depth := 0
	inString := false
	escape := false
	for i := start; i < len(raw); i++ {
		c := raw[i]
		if inString {
			if escape {
				escape = false
				continue
			}
			if c == '\\' {
				escape = true
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}
	return -1
}

func rowsFromHyattStore(obj map[string]any, params map[string]string) []hyattAwardRow {
	meta := calendarMeta(obj, params)
	days, _ := obj["days"].(map[string]any)
	dates := make([]string, 0, len(days))
	for date := range days {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	rows := make([]hyattAwardRow, 0, len(dates))
	for _, date := range dates {
		cats, _ := days[date].(map[string]any)
		if len(cats) == 0 {
			continue
		}
		names := make([]string, 0, len(cats))
		for name := range cats {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, roomCategory := range names {
			entry, _ := cats[roomCategory].(map[string]any)
			points := firstInt(entry["pointsValue"])
			nights := atoiDefault(meta["nights"], 0)
			checkout := meta["checkoutDate"]
			if nights > 0 {
				if rowCheckout, err := addHyattNights(date, nights); err == nil {
					checkout = rowCheckout
				}
			}
			rows = append(rows, hyattAwardRow{
				SpiritCode:     meta["spiritCode"],
				CheckinDate:    date,
				CheckoutDate:   checkout,
				Date:           date,
				Nights:         nights,
				RoomCategory:   roomCategory,
				IsStandardRoom: strings.EqualFold(roomCategory, "STANDARD_ROOM"),
				Available:      points > 0,
				PointsValue:    points,
				PointsLevel:    stringValue(entry["pointsLevel"]),
				CashRate:       stringValue(entry["rate"]),
				Source:         "hyatt-calendar",
			})
		}
	}
	return rows
}

func calendarMeta(obj map[string]any, params map[string]string) map[string]string {
	checkin := firstNonEmpty(stringValue(obj["stayStartDay"]), stringValue(obj["startDate"]), params["startDate"])
	checkout := firstNonEmpty(stringValue(obj["stayEndDay"]), stringValue(obj["endDate"]), params["endDate"])
	roomCategory := firstNonEmpty(stringValue(obj["roomCategory"]), params["roomCategory"], "STANDARD_ROOM")
	nights := nightsBetween(checkin, checkout)
	return map[string]string{
		"spiritCode":   firstNonEmpty(stringValue(obj["spiritCode"]), params["spiritCode"]),
		"checkinDate":  checkin,
		"checkoutDate": checkout,
		"roomCategory": roomCategory,
		"nights":       strconv.Itoa(nights),
	}
}

func rowFromFlatCalendarObject(obj map[string]any) (hyattAwardRow, bool) {
	date := firstNonEmpty(stringValue(obj["date"]), stringValue(obj["checkinDate"]), stringValue(obj["startDate"]))
	roomCategory := firstNonEmpty(stringValue(obj["roomCategory"]), "STANDARD_ROOM")
	if date == "" && stringValue(obj["spiritCode"]) == "" {
		return hyattAwardRow{}, false
	}
	checkin := firstNonEmpty(stringValue(obj["checkinDate"]), stringValue(obj["startDate"]), date)
	checkout := firstNonEmpty(stringValue(obj["checkoutDate"]), stringValue(obj["endDate"]))
	nights := intValue(obj["nights"])
	if nights == 0 {
		nights = nightsBetween(checkin, checkout)
	}
	points := firstNonZero(intValue(obj["pointsValue"]), intValue(obj["points"]))
	return hyattAwardRow{
		HotelName:      stringValue(obj["hotelName"]),
		SpiritCode:     stringValue(obj["spiritCode"]),
		City:           stringValue(obj["city"]),
		State:          stringValue(obj["state"]),
		Country:        stringValue(obj["country"]),
		Category:       intValue(obj["category"]),
		CheckinDate:    checkin,
		CheckoutDate:   checkout,
		Date:           date,
		Nights:         nights,
		RoomCategory:   roomCategory,
		IsStandardRoom: boolValue(obj["isStandardRoom"]) || strings.EqualFold(roomCategory, "STANDARD_ROOM"),
		Available:      boolValue(obj["available"]) || points > 0,
		PointsValue:    points,
		PointsLevel:    stringValue(obj["pointsLevel"]),
		CashRate:       stringValue(obj["cashRate"]),
		Source:         firstNonEmpty(stringValue(obj["source"]), "local"),
	}, true
}

func localHyattRows(cmd *cobra.Command, flags *rootFlags, dbPath string) ([]hyattAwardRow, error) {
	db, err := openHyattStore(cmd, flags, dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if err := requireHyattSnapshotSynced(db, "calendars"); err != nil {
		return nil, err
	}
	hintIfStale(cmd, db, "calendars", flags.maxAge)
	raw, err := db.List("calendars", 0)
	if err != nil {
		return nil, err
	}
	hotels, _ := localHyattHotelsFromStore(db)
	byCode := map[string]hyattHotel{}
	for _, h := range hotels {
		byCode[strings.ToUpper(h.SpiritCode)] = h
	}
	var rows []hyattAwardRow
	for _, item := range raw {
		parsed, _, ok := hyattRowsFromPayload(item, nil)
		if !ok {
			continue
		}
		for _, row := range parsed {
			if h, ok := byCode[strings.ToUpper(row.SpiritCode)]; ok {
				row.HotelName = firstNonEmpty(row.HotelName, h.Name)
				row.City = firstNonEmpty(row.City, h.City)
				row.State = firstNonEmpty(row.State, h.State)
				row.Country = firstNonEmpty(row.Country, h.Country)
				if row.Category == 0 {
					row.Category = h.Category
				}
			}
			rows = append(rows, row)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Date == rows[j].Date {
			return rows[i].SpiritCode < rows[j].SpiritCode
		}
		return rows[i].Date < rows[j].Date
	})
	return rows, nil
}

func localHyattHotels(cmd *cobra.Command, flags *rootFlags, dbPath string) ([]hyattHotel, error) {
	db, err := openHyattStore(cmd, flags, dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if err := requireHyattSnapshotSynced(db, "hotels"); err != nil {
		return nil, err
	}
	hintIfStale(cmd, db, "hotels", flags.maxAge)
	return localHyattHotelsFromStore(db)
}

func openHyattStore(cmd *cobra.Command, flags *rootFlags, dbPath string) (*store.Store, error) {
	if flags != nil && flags.dataSource == "live" {
		return nil, usageErr(fmt.Errorf("this command reads local Hyatt snapshots; --data-source=live has no local equivalent"))
	}
	if dbPath == "" {
		dbPath = defaultDBPath("hyatt-cli")
	}
	if _, err := os.Stat(dbPath); err != nil {
		if os.IsNotExist(err) {
			return nil, configErr(fmt.Errorf("local Hyatt snapshot DB does not exist at %s; run 'hyatt-cli sync --resources hotels,calendars' after configuring browser cookies", dbPath))
		}
		return nil, err
	}
	_, cancel := boundCtx(cmd.Context(), flags)
	defer cancel()
	return store.OpenReadOnly(dbPath)
}

func requireHyattSnapshotSynced(db *store.Store, resourceType string) error {
	state, err := readSyncHintState(db, resourceType)
	if err != nil {
		return err
	}
	if !state.hasState {
		return configErr(fmt.Errorf("local Hyatt %s snapshots have not been synced; run 'hyatt-cli sync --resources %s' after configuring browser cookies", resourceType, resourceType))
	}
	return nil
}

func localHyattHotelsFromStore(db *store.Store) ([]hyattHotel, error) {
	raw, err := db.List("hotels", 0)
	if err != nil {
		return nil, err
	}
	hotels := make([]hyattHotel, 0, len(raw))
	for _, item := range raw {
		if hotel, ok := hotelFromJSON(item); ok {
			hotels = append(hotels, hotel)
		}
	}
	sort.Slice(hotels, func(i, j int) bool {
		if strings.EqualFold(hotels[i].City, hotels[j].City) {
			return hotels[i].Name < hotels[j].Name
		}
		return hotels[i].City < hotels[j].City
	})
	return hotels, nil
}

func freshCachedHyattHotels(ctx context.Context, flags *rootFlags, dbPath string) ([]hyattHotel, DataProvenance, bool) {
	if flags != nil && (flags.noCache || flags.dataSource != "auto") {
		return nil, DataProvenance{}, false
	}
	maxAge := hyattHotelsCacheMaxAge()
	if maxAge <= 0 {
		return nil, DataProvenance{}, false
	}
	dbPath = hyattDBPath(dbPath)
	if _, err := os.Stat(dbPath); err != nil {
		return nil, DataProvenance{}, false
	}
	db, err := store.OpenReadOnly(dbPath)
	if err != nil {
		return nil, DataProvenance{}, false
	}
	defer db.Close()
	state, err := readSyncHintState(db, "hotels")
	if err != nil || !state.hasState {
		return nil, DataProvenance{}, false
	}
	if age := time.Since(state.lastSynced); age > maxAge {
		return nil, DataProvenance{}, false
	}
	hotels, err := localHyattHotelsFromStore(db)
	if err != nil || len(hotels) == 0 {
		return nil, DataProvenance{}, false
	}
	prov := DataProvenance{
		Source:       "local",
		Reason:       "hyatt_hotels_cache",
		ResourceType: "hotels",
		SyncedAt:     &state.lastSynced,
	}
	return hotels, attachFreshness(prov, flags), true
}

func writeHyattHotelsCache(ctx context.Context, flags *rootFlags, dbPath string, hotels []hyattHotel) {
	if flags != nil && flags.noCache {
		return
	}
	items := hyattHotelCacheItems(hotels)
	if len(items) == 0 {
		return
	}
	db, err := store.OpenWithContext(ctx, hyattDBPath(dbPath))
	if err != nil {
		return
	}
	defer db.Close()
	stored, _, err := db.UpsertBatch("hotels", items)
	if err != nil || stored == 0 {
		return
	}
	_ = db.SaveSyncState("hotels", "", stored)
}

func hyattDBPath(dbPath string) string {
	if strings.TrimSpace(dbPath) == "" {
		return defaultDBPath("hyatt-cli")
	}
	return dbPath
}

func hyattHotelsCacheMaxAge() time.Duration {
	raw := strings.TrimSpace(os.Getenv("HYATT_HOTELS_CACHE_MAX_AGE"))
	if raw == "" {
		return defaultHyattHotelsCacheMaxAge
	}
	maxAge, err := time.ParseDuration(raw)
	if err != nil {
		return defaultHyattHotelsCacheMaxAge
	}
	return maxAge
}

func hyattHotelCacheItems(hotels []hyattHotel) []json.RawMessage {
	items := make([]json.RawMessage, 0, len(hotels))
	for _, hotel := range hotels {
		if strings.TrimSpace(hotel.SpiritCode) == "" {
			continue
		}
		data, err := json.Marshal(hotel)
		if err != nil {
			continue
		}
		var obj map[string]any
		if err := json.Unmarshal(data, &obj); err != nil {
			continue
		}
		obj["id"] = hotel.SpiritCode
		item, err := json.Marshal(obj)
		if err != nil {
			continue
		}
		items = append(items, item)
	}
	return items
}

func normalizeHyattHotelsData(raw json.RawMessage) (json.RawMessage, bool) {
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return nil, false
	}
	hotels := hotelsFromHyattValue(value)
	if len(hotels) == 0 {
		return nil, false
	}
	data, err := json.Marshal(hotels)
	return data, err == nil
}

func hotelsFromHyattValue(value any) []hyattHotel {
	switch v := value.(type) {
	case []any:
		hotels := make([]hyattHotel, 0, len(v))
		for _, item := range v {
			if obj, ok := item.(map[string]any); ok {
				if hotel, ok := hotelFromObject(obj); ok {
					hotels = append(hotels, hotel)
				}
			}
		}
		sortHyattHotels(hotels)
		return hotels
	case map[string]any:
		hotels := make([]hyattHotel, 0, len(v))
		for code, item := range v {
			obj, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if obj["spiritCode"] == nil {
				obj["spiritCode"] = code
			}
			if hotel, ok := hotelFromObject(obj); ok {
				hotels = append(hotels, hotel)
			}
		}
		sortHyattHotels(hotels)
		return hotels
	default:
		return nil
	}
}

func sortHyattHotels(hotels []hyattHotel) {
	sort.Slice(hotels, func(i, j int) bool {
		if strings.EqualFold(hotels[i].City, hotels[j].City) {
			return hotels[i].Name < hotels[j].Name
		}
		return hotels[i].City < hotels[j].City
	})
}

func hotelFromJSON(raw json.RawMessage) (hyattHotel, bool) {
	var obj map[string]any
	if json.Unmarshal(raw, &obj) != nil {
		return hyattHotel{}, false
	}
	return hotelFromObject(obj)
}

func hotelFromObject(obj map[string]any) (hyattHotel, bool) {
	location, _ := obj["location"].(map[string]any)
	stateProvince, _ := location["stateProvince"].(map[string]any)
	country, _ := location["country"].(map[string]any)
	awardCategory, _ := obj["awardCategory"].(map[string]any)
	brand, _ := obj["brand"].(map[string]any)
	h := hyattHotel{
		Name:       firstNonEmpty(stringValue(obj["name"]), stringValue(obj["hotelName"])),
		SpiritCode: strings.ToUpper(firstNonEmpty(stringValue(obj["spiritCode"]), stringValue(obj["spirit_code"]), stringValue(obj["code"]))),
		City:       firstNonEmpty(stringValue(obj["city"]), stringValue(obj["destination"]), stringValue(location["city"])),
		State:      firstNonEmpty(stringValue(obj["state"]), stringValue(obj["province"]), stringValue(stateProvince["key"]), stringValue(stateProvince["label"])),
		Country:    firstNonEmpty(stringValue(obj["country"]), stringValue(country["key"]), stringValue(country["displayName"]), stringValue(country["label"])),
		Category:   firstNonZero(intValue(obj["category"]), intValue(awardCategory["key"]), intValue(awardCategory["label"])),
		Brand:      firstNonEmpty(stringValue(brand["label"]), stringValue(brand["key"]), stringValue(obj["brand"])),
	}
	return h, h.SpiritCode != "" || h.Name != ""
}

func resolveHyattCity(hotels []hyattHotel, city string) []hyattHotel {
	want := normalizeText(city)
	aliases := map[string]bool{want: true}
	if strings.HasSuffix(want, " city") {
		aliases[strings.TrimSpace(strings.TrimSuffix(want, " city"))] = true
	}
	var out []hyattHotel
	for _, h := range hotels {
		hCity := normalizeText(h.City)
		hName := normalizeText(h.Name)
		for alias := range aliases {
			if alias == "" {
				continue
			}
			if hCity == alias || strings.Contains(hCity, alias) || strings.Contains(hName, alias) {
				out = append(out, h)
				break
			}
		}
	}
	return out
}

func filterAwardRows(rows []hyattAwardRow, hotelsCSV, city, start, end string, nights int, roomCategoriesCSV string) []hyattAwardRow {
	hotelSet := csvSetUpper(hotelsCSV)
	roomSet := csvSetUpper(roomCategoriesCSV)
	out := []hyattAwardRow{}
	for _, row := range rows {
		if len(hotelSet) > 0 && !hotelSet[strings.ToUpper(row.SpiritCode)] {
			continue
		}
		if city != "" && normalizeText(row.City) != normalizeText(city) {
			continue
		}
		if start != "" && row.Date < start {
			continue
		}
		if end != "" && row.Date > end {
			continue
		}
		if nights > 0 && row.Nights != nights {
			continue
		}
		if len(roomSet) > 0 && !roomSet[strings.ToUpper(row.RoomCategory)] {
			continue
		}
		out = append(out, row)
	}
	return out
}

func printHyattEmpty(cmd *cobra.Command, flags *rootFlags) error {
	return printJSONFiltered(cmd.OutOrStdout(), []any{}, flags)
}

func csvSetUpper(s string) map[string]bool {
	out := map[string]bool{}
	for _, part := range strings.Split(s, ",") {
		part = strings.ToUpper(strings.TrimSpace(part))
		if part != "" {
			out[part] = true
		}
	}
	return out
}

func normalizeText(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, "-", " ")
	return strings.Join(strings.Fields(s), " ")
}

func nightsBetween(start, end string) int {
	if start == "" || end == "" {
		return 0
	}
	a, errA := time.Parse("2006-01-02", start)
	b, errB := time.Parse("2006-01-02", end)
	if errA != nil || errB != nil || !b.After(a) {
		return 0
	}
	return int(b.Sub(a).Hours() / 24)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func stringValue(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		if t == float64(int(t)) {
			return strconv.Itoa(int(t))
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case nil:
		return ""
	default:
		return ""
	}
}

func intValue(v any) int {
	switch t := v.(type) {
	case int:
		return t
	case float64:
		return int(t)
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(strings.ReplaceAll(t, ",", "")))
		return i
	case []any:
		return firstInt(t)
	default:
		return 0
	}
}

func firstInt(v any) int {
	switch t := v.(type) {
	case []any:
		for _, item := range t {
			if i := intValue(item); i != 0 {
				return i
			}
		}
	}
	return intValue(v)
}

func boolValue(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return strings.EqualFold(t, "true") || strings.EqualFold(t, "yes")
	default:
		return false
	}
}

func atoiDefault(s string, fallback int) int {
	i, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return fallback
	}
	return i
}

func categoryMax(cert string) int {
	switch strings.ToLower(strings.TrimSpace(cert)) {
	case "cat1-4", "cat 1-4", "1-4":
		return 4
	case "cat1-7", "cat 1-7", "1-7":
		return 7
	default:
		return 0
	}
}

func dateBeforeOrEqual(a, b string) bool {
	if a == "" || b == "" {
		return true
	}
	return a <= b
}

func ignoreSQLNoRows(err error) error {
	if err == nil || err == sql.ErrNoRows {
		return nil
	}
	return err
}
