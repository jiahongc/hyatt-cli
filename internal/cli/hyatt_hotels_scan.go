package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func addHyattHotelSubcommands(parent *cobra.Command, flags *rootFlags) {
	parent.AddCommand(newHyattResolveCityCmd(flags))
}

// pp:data-source local
func newHyattResolveCityCmd(flags *rootFlags) *cobra.Command {
	var city string
	var dbPath string
	cmd := &cobra.Command{
		Use:         "resolve-city",
		Short:       "Resolve a city into matching Hyatt hotels and spirit codes",
		Example:     "  hyatt-pp-cli hotels resolve-city --city \"New York City\" --json --select name,spiritCode,city,state,category",
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
				return err
			}
			return printJSONFiltered(cmd.OutOrStdout(), resolveHyattCity(hotels, city), flags)
		},
	}
	cmd.Flags().StringVar(&city, "city", "", "City to resolve, such as \"New York City\"")
	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: ~/.local/share/hyatt-pp-cli/data.db)")
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

// pp:data-source local
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
		Example:     "  hyatt-pp-cli scan hotel --hotels kulal --start 2026-09-01 --end 2026-09-30 --nights 1 --room-categories STANDARD_ROOM --agent",
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
				return err
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
	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: ~/.local/share/hyatt-pp-cli/data.db)")
	return cmd
}

// pp:data-source local
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
		Example:     "  hyatt-pp-cli scan city --city \"New York City\" --start 2026-09-01 --end 2026-09-07 --nights 2 --room-categories STANDARD_ROOM,SUITE --agent",
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
				return err
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
				return err
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
	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: ~/.local/share/hyatt-pp-cli/data.db)")
	return cmd
}
