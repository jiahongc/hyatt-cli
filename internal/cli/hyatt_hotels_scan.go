package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func addHyattHotelSubcommands(parent *cobra.Command, flags *rootFlags) {
	parent.AddCommand(newHyattResolveCityCmd(flags))
}

// hyatt:data-source local
func newHyattResolveCityCmd(flags *rootFlags) *cobra.Command {
	var city string
	var dbPath string
	cmd := &cobra.Command{
		Use:         "resolve-city",
		Short:       "Resolve a city into matching Hyatt hotels and spirit codes",
		Example:     "  hyatt-cli hotels resolve-city --city \"New York City\" --json --select name,spiritCode,city,state,category",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return nil
			}
			if strings.TrimSpace(city) == "" {
				return usageErr(fmt.Errorf("--city is required"))
			}
			hotels, err := localHyattHotels(cmd, flags, dbPath)
			if err != nil {
				if flags.dataSource == "local" {
					return err
				}
				hotels, err = liveHyattHotels(cmd, flags)
				if err != nil {
					return err
				}
			}
			return printJSONFiltered(cmd.OutOrStdout(), resolveHyattCity(hotels, city), flags)
		},
	}
	cmd.Flags().StringVar(&city, "city", "", "City to resolve, such as \"New York City\"")
	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: ~/.local/share/hyatt-cli/data.db)")
	return cmd
}

func newScanCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan Hyatt award availability by hotel or city",
		RunE:  parentNoSubcommandRunE(flags),
	}
	cmd.AddCommand(newScanHotelCmd(flags))
	cmd.AddCommand(newScanCityCmd(flags))
	return cmd
}

// hyatt:data-source local
func newScanHotelCmd(flags *rootFlags) *cobra.Command {
	var hotels string
	var start string
	var end string
	var nights int
	var roomCategories string
	var dbPath string
	cmd := &cobra.Command{
		Use:         "hotel",
		Short:       "Scan one or more Hyatt spirit codes for points availability",
		Example:     "  hyatt-cli scan hotel --hotels kulal --start 2026-09-01 --end 2026-09-30 --nights 1 --room-categories STANDARD_ROOM --agent",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return nil
			}
			if strings.TrimSpace(hotels) == "" {
				return usageErr(fmt.Errorf("--hotels is required"))
			}
			rows, err := localHyattRows(cmd, flags, dbPath)
			if err != nil {
				if flags.dataSource == "local" {
					return err
				}
				rows, err = liveHyattCalendarRows(cmd, flags, hotels, start, end, nights, roomCategories)
				if err != nil {
					return err
				}
			}
			rows = filterAwardRows(rows, hotels, "", start, end, nights, roomCategories)
			return printJSONFiltered(cmd.OutOrStdout(), rows, flags)
		},
	}
	cmd.Flags().StringVar(&hotels, "hotels", "", "Comma-separated Hyatt spirit codes to scan")
	cmd.Flags().StringVar(&start, "start", "", "Start date YYYY-MM-DD")
	cmd.Flags().StringVar(&end, "end", "", "End date YYYY-MM-DD")
	cmd.Flags().IntVar(&nights, "nights", 1, "Length of stay; Hyatt can show different availability for 1 night vs multiple nights")
	cmd.Flags().StringVar(&roomCategories, "room-categories", "STANDARD_ROOM", "Comma-separated Hyatt room categories, e.g. STANDARD_ROOM,SUITE")
	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: ~/.local/share/hyatt-cli/data.db)")
	return cmd
}

// hyatt:data-source local
func newScanCityCmd(flags *rootFlags) *cobra.Command {
	var city string
	var start string
	var end string
	var nights int
	var roomCategories string
	var dbPath string
	cmd := &cobra.Command{
		Use:         "city",
		Short:       "Scan all Hyatt hotels in a city for points availability",
		Example:     "  hyatt-cli scan city --city \"New York City\" --start 2026-09-01 --end 2026-09-07 --nights 2 --room-categories STANDARD_ROOM,SUITE --agent",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return nil
			}
			if strings.TrimSpace(city) == "" {
				return usageErr(fmt.Errorf("--city is required"))
			}
			hotels, err := localHyattHotels(cmd, flags, dbPath)
			if err != nil {
				if flags.dataSource == "local" {
					return err
				}
				hotels, err = liveHyattHotels(cmd, flags)
				if err != nil {
					return err
				}
			}
			matches := resolveHyattCity(hotels, city)
			if len(matches) == 0 {
				return printHyattEmpty(cmd, flags)
			}
			codes := make([]string, 0, len(matches))
			for _, hotel := range matches {
				codes = append(codes, hotel.SpiritCode)
			}
			rows, err := localHyattRows(cmd, flags, dbPath)
			if err != nil {
				if flags.dataSource == "local" {
					return err
				}
				rows, err = liveHyattCalendarRows(cmd, flags, strings.Join(codes, ","), start, end, nights, roomCategories)
				if err != nil {
					return err
				}
			}
			rows = filterAwardRows(rows, strings.Join(codes, ","), "", start, end, nights, roomCategories)
			return printJSONFiltered(cmd.OutOrStdout(), rows, flags)
		},
	}
	cmd.Flags().StringVar(&city, "city", "", "City to scan, such as \"New York City\"")
	cmd.Flags().StringVar(&start, "start", "", "Start date YYYY-MM-DD")
	cmd.Flags().StringVar(&end, "end", "", "End date YYYY-MM-DD")
	cmd.Flags().IntVar(&nights, "nights", 1, "Length of stay; Hyatt can show different availability for 1 night vs multiple nights")
	cmd.Flags().StringVar(&roomCategories, "room-categories", "STANDARD_ROOM", "Comma-separated Hyatt room categories, e.g. STANDARD_ROOM,SUITE")
	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: ~/.local/share/hyatt-cli/data.db)")
	return cmd
}

func liveHyattHotels(cmd *cobra.Command, flags *rootFlags) ([]hyattHotel, error) {
	c, err := flags.newClient()
	if err != nil {
		return nil, err
	}
	path := "/explore-hotels/service/hotels"
	params := map[string]string{}
	var data json.RawMessage
	if shouldUseHyattBrowserFirst() {
		data, err = hyattBrowserJSON(cmd.Context(), c.BaseURL, path, params)
	} else {
		data, err = c.GetWithHeaders(cmd.Context(), path, params, nil)
	}
	if err != nil && !shouldUseHyattBrowserFirst() {
		fallbackData, attempted, fallbackErr := hyattBrowserJSONFallback(cmd.Context(), c.BaseURL, path, params, err)
		if !attempted {
			return nil, classifyAPIError(err, flags)
		}
		if fallbackErr != nil {
			return nil, apiErr(fallbackErr)
		}
		data = fallbackData
	}
	if err != nil {
		return nil, apiErr(err)
	}
	normalized, ok := normalizeHyattHotelsData(data)
	if !ok {
		return nil, apiErr(fmt.Errorf("Hyatt hotel metadata response did not contain hotel rows"))
	}
	var hotels []hyattHotel
	if err := json.Unmarshal(normalized, &hotels); err != nil {
		return nil, apiErr(fmt.Errorf("parsing Hyatt hotel metadata: %w", err))
	}
	return hotels, nil
}

func liveHyattCalendarRows(cmd *cobra.Command, flags *rootFlags, hotelsCSV, start, end string, nights int, roomCategoriesCSV string) ([]hyattAwardRow, error) {
	if strings.TrimSpace(start) == "" || strings.TrimSpace(end) == "" {
		return nil, usageErr(fmt.Errorf("--start and --end are required for live Hyatt scans"))
	}
	if nights < 1 {
		return nil, usageErr(fmt.Errorf("--nights must be >= 1"))
	}
	c, err := flags.newClient()
	if err != nil {
		return nil, err
	}
	checkout, err := addHyattNights(start, nights)
	if err != nil {
		return nil, usageErr(err)
	}
	path := "/explore-hotels/rate-calendar"
	codes := csvValues(hotelsCSV)
	categories := csvValues(roomCategoriesCSV)
	if len(categories) == 0 {
		categories = []string{"STANDARD_ROOM"}
	}
	var out []hyattAwardRow
	for _, code := range codes {
		for _, category := range categories {
			params := map[string]string{
				"spiritCode":   code,
				"startDate":    start,
				"endDate":      checkout,
				"rooms":        "1",
				"adults":       "1",
				"kids":         "0",
				"rate":         "Standard",
				"roomCategory": category,
				"vrcEnabled":   "true",
			}
			var data json.RawMessage
			if shouldUseHyattBrowserFirst() {
				data, err = hyattBrowserCalendar(cmd.Context(), c.BaseURL, path, params)
			} else {
				data, err = c.GetWithHeaders(cmd.Context(), path, params, nil)
			}
			if err != nil && !shouldUseHyattBrowserFirst() {
				fallbackData, attempted, fallbackErr := hyattBrowserCalendarFallback(cmd.Context(), c.BaseURL, path, params, err)
				if !attempted {
					return nil, classifyAPIError(err, flags)
				}
				if fallbackErr != nil {
					return nil, apiErr(fmt.Errorf("Hyatt calendar fallback failed for %s/%s: %w", code, category, fallbackErr))
				}
				data = fallbackData
			}
			if err != nil {
				return nil, apiErr(fmt.Errorf("Hyatt calendar browser transport failed for %s/%s: %w", code, category, err))
			}
			rows, _, ok := hyattRowsFromPayload(data, params)
			if !ok {
				return nil, apiErr(fmt.Errorf("Hyatt calendar response for %s did not contain availability rows", code))
			}
			rows = filterAwardRows(rows, code, "", "", "", nights, category)
			out = append(out, rows...)
		}
	}
	return filterAwardRows(out, hotelsCSV, "", start, end, nights, roomCategoriesCSV), nil
}

func csvValues(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func addHyattNights(start string, nights int) (string, error) {
	d, err := time.Parse("2006-01-02", start)
	if err != nil {
		return "", fmt.Errorf("--start must be YYYY-MM-DD: %w", err)
	}
	return d.AddDate(0, 0, nights).Format("2006-01-02"), nil
}
