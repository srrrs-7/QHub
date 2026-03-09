package cmd

import (
	"github.com/spf13/cobra"
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Administrative commands (batch operations, maintenance)",
	Long:  "Administrative commands for system maintenance and batch operations.",
	Example: `  # Trigger batch aggregation of analytics data
  qhub admin aggregate`,
}

var adminAggregateCmd = &cobra.Command{
	Use:     "aggregate",
	Short:   "Trigger batch aggregation of analytics data",
	Example: "  qhub admin aggregate",
	RunE: func(_ *cobra.Command, _ []string) error {
		var result any
		if err := apiPost("/api/v1/admin/batch/aggregate", nil, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(adminAggregateCmd)
}
