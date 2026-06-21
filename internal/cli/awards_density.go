// Copyright 2026 Jiahong Chen and contributors. Licensed under Apache-2.0. See LICENSE.
// Maintained in the World of Hyatt CLI repository.

package cli

import (
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

type hyattDensityRow struct {
	Bucket        string `json:"bucket"`
	HotelsChecked int    `json:"hotelsChecked"`
	OpenOptions   int    `json:"openOptions"`
	MinPoints     int    `json:"minPoints,omitempty"`
	Nights        int    `json:"nights,omitempty"`
}

// hyatt:data-source local
func newNovelAwardsDensityCmd(flags *rootFlags) *cobra.Command {
	var flagHotels string
	var flagStart string
	var flagEnd string
	var flagBucket string
	var flagNights int
	var flagRoomCategories string
	var dbPath string

	cmd := &cobra.Command{
		Use:         "density",
		Short:       "Show which dates or weeks have the most Hyatt award options across a hotel shortlist.",
		Example:     "  hyatt-cli awards density --hotels CHIRH,NYCUA,PARPH --start 2026-07-01 --end 2026-09-30 --nights 2 --bucket week --agent",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return nil
			}
			rows, err := localHyattRows(cmd, flags, dbPath)
			if err != nil {
				return err
			}
			rows = filterAwardRows(rows, flagHotels, "", flagStart, flagEnd, flagNights, flagRoomCategories)
			type agg struct {
				hotels map[string]bool
				open   int
				min    int
			}
			byBucket := map[string]*agg{}
			for _, row := range rows {
				bucket := densityBucket(row.Date, flagBucket)
				a := byBucket[bucket]
				if a == nil {
					a = &agg{hotels: map[string]bool{}}
					byBucket[bucket] = a
				}
				a.hotels[row.SpiritCode] = true
				if row.Available {
					a.open++
					if row.PointsValue > 0 && (a.min == 0 || row.PointsValue < a.min) {
						a.min = row.PointsValue
					}
				}
			}
			keys := make([]string, 0, len(byBucket))
			for key := range byBucket {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			out := make([]hyattDensityRow, 0, len(keys))
			for _, key := range keys {
				a := byBucket[key]
				out = append(out, hyattDensityRow{Bucket: key, HotelsChecked: len(a.hotels), OpenOptions: a.open, MinPoints: a.min, Nights: flagNights})
			}
			return printJSONFiltered(cmd.OutOrStdout(), out, flags)
		},
	}
	cmd.Flags().StringVar(&flagHotels, "hotels", "", "Comma-separated Hyatt spirit codes")
	cmd.Flags().StringVar(&flagStart, "start", "", "Start date YYYY-MM-DD")
	cmd.Flags().StringVar(&flagEnd, "end", "", "End date YYYY-MM-DD")
	cmd.Flags().StringVar(&flagBucket, "bucket", "date", "Bucket size: date or week")
	cmd.Flags().IntVar(&flagNights, "nights", 1, "Length of stay to match")
	cmd.Flags().StringVar(&flagRoomCategories, "room-categories", "STANDARD_ROOM", "Room categories to include")
	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: ~/.local/share/hyatt-cli/data.db)")
	return cmd
}

func densityBucket(date, bucket string) string {
	if bucket != "week" {
		return date
	}
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	year, week := t.ISOWeek()
	return fmt.Sprintf("%04d-W%02d", year, week)
}
