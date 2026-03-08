package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var analyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "View analytics and trends",
}

var analyticsPromptCmd = &cobra.Command{
	Use:   "prompt <prompt-id>",
	Short: "Get analytics for a prompt (all versions)",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var result any
		if err := apiGet("/api/v1/prompts/"+args[0]+"/analytics", &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var analyticsVersionCmd = &cobra.Command{
	Use:   "version <prompt-id> <version-number>",
	Short: "Get analytics for a specific version",
	Args:  cobra.ExactArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		var result any
		path := fmt.Sprintf("/api/v1/prompts/%s/versions/%s/analytics", args[0], args[1])
		if err := apiGet(path, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var analyticsProjectCmd = &cobra.Command{
	Use:   "project <project-id>",
	Short: "Get analytics for a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var result any
		if err := apiGet("/api/v1/projects/"+args[0]+"/analytics", &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var analyticsTrendCmd = &cobra.Command{
	Use:   "trend <prompt-id>",
	Short: "Get daily execution trend for a prompt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		days, _ := cmd.Flags().GetInt("days")
		path := fmt.Sprintf("/api/v1/prompts/%s/trend?days=%d", args[0], days)

		var result any
		if err := apiGet(path, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(analyticsCmd)
	analyticsCmd.AddCommand(analyticsPromptCmd, analyticsVersionCmd, analyticsProjectCmd, analyticsTrendCmd)

	analyticsTrendCmd.Flags().Int("days", 30, "Number of days for trend data")
}
