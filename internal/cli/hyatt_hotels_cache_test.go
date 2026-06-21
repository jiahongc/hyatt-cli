package cli

import (
	"context"
	"path/filepath"
	"testing"
)

func TestHyattHotelsCacheRoundTrip(t *testing.T) {
	t.Setenv("HYATT_HOTELS_CACHE_MAX_AGE", "24h")
	dbPath := filepath.Join(t.TempDir(), "data.db")
	flags := &rootFlags{dataSource: "auto"}
	hotels := []hyattHotel{
		{Name: "Park Hyatt Test", SpiritCode: "TEST1", City: "New York", State: "NY", Country: "US", Category: 7, Brand: "Park Hyatt"},
		{Name: "Hyatt Test", SpiritCode: "TEST2", City: "Tokyo", Country: "JP", Category: 4, Brand: "Hyatt"},
	}

	writeHyattHotelsCache(context.Background(), flags, dbPath, hotels)

	cached, prov, ok := freshCachedHyattHotels(context.Background(), flags, dbPath)
	if !ok {
		t.Fatal("freshCachedHyattHotels missed cache after write")
	}
	if prov.Source != "local" || prov.Reason != "hyatt_hotels_cache" || prov.ResourceType != "hotels" {
		t.Fatalf("unexpected provenance: %+v", prov)
	}
	if len(cached) != len(hotels) {
		t.Fatalf("cached hotel count = %d, want %d", len(cached), len(hotels))
	}
	if cached[0].SpiritCode != "TEST1" || cached[1].SpiritCode != "TEST2" {
		t.Fatalf("cached hotels = %+v", cached)
	}
}

func TestHyattHotelsCacheBypass(t *testing.T) {
	t.Setenv("HYATT_HOTELS_CACHE_MAX_AGE", "24h")
	dbPath := filepath.Join(t.TempDir(), "data.db")
	flags := &rootFlags{dataSource: "auto"}
	writeHyattHotelsCache(context.Background(), flags, dbPath, []hyattHotel{{Name: "Hyatt Test", SpiritCode: "TEST"}})

	for _, tt := range []struct {
		name  string
		flags *rootFlags
	}{
		{name: "no cache", flags: &rootFlags{dataSource: "auto", noCache: true}},
		{name: "live source", flags: &rootFlags{dataSource: "live"}},
		{name: "disabled ttl", flags: &rootFlags{dataSource: "auto"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "disabled ttl" {
				t.Setenv("HYATT_HOTELS_CACHE_MAX_AGE", "0")
			} else {
				t.Setenv("HYATT_HOTELS_CACHE_MAX_AGE", "24h")
			}
			if _, _, ok := freshCachedHyattHotels(context.Background(), tt.flags, dbPath); ok {
				t.Fatalf("freshCachedHyattHotels returned cache for %s", tt.name)
			}
		})
	}
}
