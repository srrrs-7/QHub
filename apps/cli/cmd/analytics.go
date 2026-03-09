package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var analyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "View analytics, trends, and usage statistics",
	Long:  "View aggregated analytics for prompts, versions, and projects including execution counts, token usage, latency, and daily trends.",
	Example: `  # View analytics for all versions of a prompt
  qhub analytics prompt <prompt-id>

  # View analytics for a specific version
  qhub analytics version <prompt-id> 3

  # View project-level analytics
  qhub analytics project <project-id>

  # View daily execution trend
  qhub analytics trend <prompt-id> --days 30`,
}

var analyticsPromptCmd = &cobra.Command{
	Use:     "prompt <prompt-id>",
	Short:   "View aggregated analytics for all versions of a prompt",
	Example: "  qhub analytics prompt <prompt-id>",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var result any
		if err := apiGet("/api/v1/prompts/"+args[0]+"/analytics", &result); err != nil {
			return err
		}
		if outputFmt == "table" {
			printAnalyticsTable(result)
		} else {
			printJSON(result)
		}
		return nil
	},
}

var analyticsVersionCmd = &cobra.Command{
	Use:     "version <prompt-id> <version-number>",
	Short:   "View analytics for a specific prompt version",
	Example: "  qhub analytics version <prompt-id> 3",
	Args:    cobra.ExactArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		var result any
		path := fmt.Sprintf("/api/v1/prompts/%s/versions/%s/analytics", args[0], args[1])
		if err := apiGet(path, &result); err != nil {
			return err
		}
		if outputFmt == "table" {
			printAnalyticsTable(result)
		} else {
			printJSON(result)
		}
		return nil
	},
}

var analyticsProjectCmd = &cobra.Command{
	Use:     "project <project-id>",
	Short:   "View aggregated analytics for an entire project",
	Example: "  qhub analytics project <project-id>",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var result any
		if err := apiGet("/api/v1/projects/"+args[0]+"/analytics", &result); err != nil {
			return err
		}
		if outputFmt == "table" {
			printAnalyticsTable(result)
		} else {
			printJSON(result)
		}
		return nil
	},
}

var analyticsTrendCmd = &cobra.Command{
	Use:   "trend <prompt-id>",
	Short: "View daily execution trend for a prompt over time",
	Example: `  qhub analytics trend <prompt-id>
  qhub analytics trend <prompt-id> --days 7`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		days, _ := cmd.Flags().GetInt("days")
		path := fmt.Sprintf("/api/v1/prompts/%s/trend?days=%d", args[0], days)

		var result any
		if err := apiGet(path, &result); err != nil {
			return err
		}
		if outputFmt == "table" {
			printAnalyticsTable(result)
		} else {
			printJSON(result)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(analyticsCmd)
	analyticsCmd.AddCommand(analyticsPromptCmd, analyticsVersionCmd, analyticsProjectCmd, analyticsTrendCmd)

	analyticsTrendCmd.Flags().Int("days", 30, "Number of days for trend data")
}
