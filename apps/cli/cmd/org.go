package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var orgCmd = &cobra.Command{
	Use:     "org",
	Aliases: []string{"organization"},
	Short:   "Manage organizations",
}

var orgGetCmd = &cobra.Command{
	Use:   "get <slug>",
	Short: "Get organization details",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var org any
		if err := apiGet("/api/v1/organizations/"+args[0], &org); err != nil {
			return err
		}
		printJSON(org)
		return nil
	},
}

var orgCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an organization",
	RunE: func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")
		plan, _ := cmd.Flags().GetString("plan")

		if name == "" || slug == "" {
			return fmt.Errorf("--name and --slug are required")
		}

		body := map[string]string{"name": name, "slug": slug, "plan": plan}
		var result any
		if err := apiPost("/api/v1/organizations", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(orgCmd)
	orgCmd.AddCommand(orgGetCmd)
	orgCmd.AddCommand(orgCreateCmd)

	orgCreateCmd.Flags().String("name", "", "Organization name (required)")
	orgCreateCmd.Flags().String("slug", "", "Organization slug (required)")
	orgCreateCmd.Flags().String("plan", "free", "Plan: free, pro, team, enterprise")
}
