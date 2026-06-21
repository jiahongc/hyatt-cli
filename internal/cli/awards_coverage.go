// Copyright 2026 Jiahong Chen and contributors. Licensed under Apache-2.0. See LICENSE.
// Maintained in the World of Hyatt CLI repository.

package cli

import (
	"sort"

	"github.com/spf13/cobra"
)

type hyattCoverageRow struct {
	SpiritCode     string   `json:"spiritCode"`
	Rows           int      `json:"rows"`
	AvailableRows  int      `json:"availableRows"`
	NightsObserved []int    `json:"nightsObserved,omitempty"`
	RoomCategories []string `json:"roomCategories,omitempty"`
	FirstDate      string   `json:"firstDate,omitempty"`
	LastDate       string   `json:"lastDate,omitempty"`
}

// hyatt:data-source local
func newNovelAwardsCoverageCmd(flags *rootFlags) *cobra.Command {
	var flagHotels string
	var flagStart string
	var flagEnd string
	var dbPath string

	cmd := &cobra.Command{
		Use:         "coverage",
		Short:       "Show stale, missing, or uneven local Hyatt calendar coverage before trusting a scan.",
		Example:     "  hyatt-cli awards coverage --hotels CHIRH,NYCUA --start 2026-07-01 --end 2026-12-31 --agent",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return nil
			}
			rows, err := localHyattRows(cmd, flags, dbPath)
			if err != nil {
				return err
			}
			rows = filterAwardRows(rows, flagHotels, "", flagStart, flagEnd, 0, "")
			type agg struct {
				total int
				open  int
				first string
				last  string
				n     map[int]bool
				cats  map[string]bool
			}
			byHotel := map[string]*agg{}
			for _, row := range rows {
				a := byHotel[row.SpiritCode]
				if a == nil {
					a = &agg{n: map[int]bool{}, cats: map[string]bool{}}
					byHotel[row.SpiritCode] = a
				}
				a.total++
				if row.Available {
					a.open++
				}
				if a.first == "" || row.Date < a.first {
					a.first = row.Date
				}
				if row.Date > a.last {
					a.last = row.Date
				}
				if row.Nights > 0 {
					a.n[row.Nights] = true
				}
				if row.RoomCategory != "" {
					a.cats[row.RoomCategory] = true
				}
			}
			keys := make([]string, 0, len(byHotel))
			for key := range byHotel {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			out := make([]hyattCoverageRow, 0, len(keys))
			for _, key := range keys {
				a := byHotel[key]
				out = append(out, hyattCoverageRow{
					SpiritCode:     key,
					Rows:           a.total,
					AvailableRows:  a.open,
					NightsObserved: intSetKeys(a.n),
					RoomCategories: stringSetKeys(a.cats),
					FirstDate:      a.first,
					LastDate:       a.last,
				})
			}
			return printJSONFiltered(cmd.OutOrStdout(), out, flags)
		},
	}
	cmd.Flags().StringVar(&flagHotels, "hotels", "", "Optional comma-separated Hyatt spirit codes")
	cmd.Flags().StringVar(&flagStart, "start", "", "Start date YYYY-MM-DD")
	cmd.Flags().StringVar(&flagEnd, "end", "", "End date YYYY-MM-DD")
	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: ~/.local/share/hyatt-cli/data.db)")
	return cmd
}

func intSetKeys(m map[int]bool) []int {
	out := make([]int, 0, len(m))
	for key := range m {
		out = append(out, key)
	}
	sort.Ints(out)
	return out
}

func stringSetKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for key := range m {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}
