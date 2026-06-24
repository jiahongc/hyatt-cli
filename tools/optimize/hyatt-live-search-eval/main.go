package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jiahongc/hyatt-cli/internal/store"
)

type metrics struct {
	TotalSeconds              float64 `json:"total_seconds"`
	BuildPassed               int     `json:"build_passed"`
	UnitTestsPassed           int     `json:"unit_tests_passed"`
	HotelsCacheSourceLocal    int     `json:"hotels_cache_source_local"`
	CityResolutionCount       int     `json:"city_resolution_count"`
	CalendarSuccess           int     `json:"calendar_success"`
	CalendarHasRequiredFields int     `json:"calendar_has_required_fields"`
	HotelsSeconds             float64 `json:"hotels_seconds"`
	ResolveCitySeconds        float64 `json:"resolve_city_seconds"`
	CalendarSeconds           float64 `json:"calendar_seconds"`
	CalendarRows              int     `json:"calendar_rows"`
	BrowserSessionsAfter      int     `json:"browser_sessions_after"`
	StdoutBytes               int     `json:"stdout_bytes"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 230*time.Second)
	defer cancel()

	var out metrics
	tmp, err := os.MkdirTemp("", "hyatt-live-search-eval-*")
	if err != nil {
		printMetrics(out)
		return
	}
	defer os.RemoveAll(tmp)

	bin := filepath.Join(tmp, "hyatt-cli")
	if runQuiet(ctx, "go", "build", "-o", bin, "./cmd/hyatt-cli") == nil {
		out.BuildPassed = 1
	}
	if runQuiet(ctx, "go", "test", "./...") == nil {
		out.UnitTestsPassed = 1
	}
	if out.BuildPassed != 1 {
		printMetrics(out)
		return
	}

	home := filepath.Join(tmp, "home")
	if err := seedHotelCache(ctx, home); err != nil {
		fmt.Fprintf(os.Stderr, "seed hotel cache: %v\n", err)
		printMetrics(out)
		return
	}

	session := "hyatt-optimize-" + strconv.Itoa(os.Getpid())
	cacheEnv := append(os.Environ(),
		"HOME="+home,
		"HYATT_BROWSER_SESSION="+session,
		"HYATT_HOTELS_CACHE_MAX_AGE=24h",
	)
	liveEnv := append(os.Environ(),
		"HYATT_BROWSER_SESSION="+session,
	)
	defer closeBrowserSession(session)

	hotelsStart := time.Now()
	hotelsOut, err := runCapture(ctx, cacheEnv, bin, "hotels", "--json", "--no-input", "--no-color", "--yes", "--timeout", "60s", "--select", "name,spiritCode,city,state,country,category,brand")
	out.HotelsSeconds = secondsSince(hotelsStart)
	out.StdoutBytes += len(hotelsOut)
	if err == nil && hotelsSourceLocal(hotelsOut) {
		out.HotelsCacheSourceLocal = 1
	}

	cityStart := time.Now()
	cityOut, err := runCapture(ctx, cacheEnv, bin, "hotels", "resolve-city", "--city", "New York City", "--json", "--no-input", "--no-color", "--yes", "--select", "name,spiritCode,city,state,country,category,brand")
	out.ResolveCitySeconds = secondsSince(cityStart)
	out.StdoutBytes += len(cityOut)
	if err == nil {
		out.CityResolutionCount = cityResolutionCount(cityOut)
	}

	calendarStart := time.Now()
	calendarOut, err := runCapture(ctx, liveEnv, bin, "scan", "hotel", "--hotels", "KULAL", "--start", "2026-09-01", "--end", "2026-09-01", "--nights", "1", "--room-categories", "STANDARD_ROOM", "--data-source", "live", "--json", "--no-input", "--no-color", "--yes", "--timeout", "180s", "--select", "spiritCode,date,nights,roomCategory,isStandardRoom,available,pointsValue")
	out.CalendarSeconds = secondsSince(calendarStart)
	out.StdoutBytes += len(calendarOut)
	if err == nil {
		rows, required := calendarRows(calendarOut)
		out.CalendarRows = rows
		if rows > 0 {
			out.CalendarSuccess = 1
		}
		if required {
			out.CalendarHasRequiredFields = 1
		}
	} else {
		fmt.Fprintf(os.Stderr, "calendar command failed: %v\n", err)
	}

	out.TotalSeconds = out.HotelsSeconds + out.ResolveCitySeconds + out.CalendarSeconds
	closeBrowserSession(session)
	out.BrowserSessionsAfter = browserSessionCount()
	printMetrics(out)
}

func seedHotelCache(ctx context.Context, home string) error {
	dbPath := filepath.Join(home, ".local", "share", "hyatt-cli", "data.db")
	db, err := store.OpenWithContext(ctx, dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	items := []json.RawMessage{
		json.RawMessage(`{"id":"NYCAM","name":"Andaz 5th Avenue","spiritCode":"NYCAM","city":"New York","state":"NY","country":"US","category":8,"brand":"Andaz"}`),
		json.RawMessage(`{"id":"NYCUA","name":"The Beekman","spiritCode":"NYCUA","city":"New York","state":"NY","country":"US","category":7,"brand":"The Unbound Collection by Hyatt"}`),
		json.RawMessage(`{"id":"LGAZL","name":"Hyatt Place Long Island City / New York City","spiritCode":"LGAZL","city":"Long Island City","state":"NY","country":"US","category":4,"brand":"Hyatt Place"}`),
		json.RawMessage(`{"id":"KULAL","name":"Alila Bangsar Kuala Lumpur","spiritCode":"KULAL","city":"Kuala Lumpur","country":"MY","category":1,"brand":"Alila"}`),
	}
	stored, _, err := db.UpsertBatch("hotels", items)
	if err != nil {
		return err
	}
	return db.SaveSyncState("hotels", "", stored)
}

func runQuiet(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCapture(ctx context.Context, env []string, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = env
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return out, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func hotelsSourceLocal(raw []byte) bool {
	var envelope struct {
		Meta struct {
			Source string `json:"source"`
			Reason string `json:"reason"`
		} `json:"meta"`
	}
	return json.Unmarshal(raw, &envelope) == nil && envelope.Meta.Source == "local" && envelope.Meta.Reason == "hyatt_hotels_cache"
}

func cityResolutionCount(raw []byte) int {
	var rows []map[string]any
	if json.Unmarshal(raw, &rows) != nil {
		return 0
	}
	return len(rows)
}

func calendarRows(raw []byte) (int, bool) {
	var rows []map[string]any
	if json.Unmarshal(raw, &rows) != nil {
		var envelope struct {
			Results []map[string]any `json:"results"`
		}
		if json.Unmarshal(raw, &envelope) != nil {
			return 0, false
		}
		rows = envelope.Results
	}
	if len(rows) == 0 {
		return 0, false
	}
	required := true
	for _, key := range []string{"spiritCode", "date", "nights", "roomCategory", "isStandardRoom", "available"} {
		if _, ok := rows[0][key]; !ok {
			required = false
		}
	}
	return len(rows), required
}

func browserSessionCount() int {
	out, err := exec.Command("browser-use", "sessions").CombinedOutput()
	if err != nil {
		return 0
	}
	count := 0
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == "running" {
			count++
		}
	}
	return count
}

func closeBrowserSession(session string) {
	if session == "" {
		return
	}
	_ = exec.Command("browser-use", "--session", session, "close").Run()
}

func secondsSince(start time.Time) float64 {
	return float64(time.Since(start).Milliseconds()) / 1000
}

func printMetrics(out metrics) {
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fmt.Println(`{"build_passed":0}`)
		return
	}
	fmt.Println(string(data))
}
