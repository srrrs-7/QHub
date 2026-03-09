package cmd

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Show current CLI configuration (API URL, token, output format)",
	Example: "  qhub config",
	RunE: func(_ *cobra.Command, _ []string) error {
		masked := authToken
		if len(masked) > 8 {
			masked = masked[:4] + "****" + masked[len(masked)-4:]
		} else if len(masked) > 0 {
			masked = "****"
		}

		config := map[string]string{
			"api_url": apiURL,
			"token":   masked,
			"output":  outputFmt,
		}
		printResult(config)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
