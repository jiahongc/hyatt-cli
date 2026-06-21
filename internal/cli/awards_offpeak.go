// Copyright 2026 Jiahong Chen and contributors. Licensed under Apache-2.0. See LICENSE.
// Maintained in the World of Hyatt CLI repository.

package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

// hyatt:data-source local
func newNovelAwardsOffpeakCmd(flags *rootFlags) *cobra.Command {
	var flagCountry string
	var flagStart string
	var flagEnd string
	var flagMinNights string
	var dbPath string

	cmd := &cobra.Command{
		Use:         "offpeak",
		Short:       "Find off-peak or unusually low-point Hyatt award clusters across synced calendar data.",
		Example:     "  hyatt-cli awards offpeak --country US --start 2026-06-01 --end 2027-05-31 --min-nights 2 --agent",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return nil
			}
			rows, err := localHyattRows(cmd, flags, dbPath)
			if err != nil {
				return err
			}
			minNights := atoiDefault(flagMinNights, 1)
			rows = filterAwardRows(rows, "", "", flagStart, flagEnd, 0, "STANDARD_ROOM")
			out := make([]hyattAwardRow, 0, len(rows))
			for _, row := range rows {
				if !row.Available || row.Nights < minNights {
					continue
				}
				if flagCountry != "" && !strings.EqualFold(row.Country, flagCountry) {
					continue
				}
				if strings.Contains(strings.ToUpper(row.PointsLevel), "OFF_PEAK") || strings.Contains(strings.ToUpper(row.PointsLevel), "LOW") {
					out = append(out, row)
				}
			}
			return printJSONFiltered(cmd.OutOrStdout(), out, flags)
		},
	}
	cmd.Flags().StringVar(&flagCountry, "country", "", "Optional country filter")
	cmd.Flags().StringVar(&flagStart, "start", "", "Start date YYYY-MM-DD")
	cmd.Flags().StringVar(&flagEnd, "end", "", "End date YYYY-MM-DD")
	cmd.Flags().StringVar(&flagMinNights, "min-nights", "1", "Minimum length of stay")
	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: ~/.local/share/hyatt-cli/data.db)")
	return cmd
}
