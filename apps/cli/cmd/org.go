package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var orgCmd = &cobra.Command{
	Use:     "org",
	Aliases: []string{"organization"},
	Short:   "Manage organizations (create, view, update)",
	Long:    "Create, view, and update organizations. Organizations are the top-level container for projects, members, and API keys.",
	Example: `  # Create a new organization
  qhub org create --name "Acme Corp" --slug acme-corp --plan pro

  # Get organization details
  qhub org get acme-corp

  # Update an organization
  qhub org update acme-corp --plan enterprise`,
}

var orgGetCmd = &cobra.Command{
	Use:     "get <slug>",
	Short:   "Get organization details by slug",
	Example: "  qhub org get acme-corp",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var org any
		if err := apiGet("/api/v1/organizations/"+args[0], &org); err != nil {
			return err
		}
		if outputFmt == "table" {
			printOrgTable(org)
		} else {
			printJSON(org)
		}
		return nil
	},
}

var orgCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new organization",
	Example: `  qhub org create --name "Acme Corp" --slug acme-corp
  qhub org create --name "Enterprise" --slug enterprise --plan enterprise`,
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
		printSuccess("Created organization '" + slug + "'")
		if outputFmt == "table" {
			printOrgTable(result)
		} else {
			printJSON(result)
		}
		return nil
	},
}

var orgUpdateCmd = &cobra.Command{
	Use:     "update <slug>",
	Short:   "Update an existing organization",
	Example: "  qhub org update acme-corp --name \"Acme Corporation\" --plan pro",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]string{}

		if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			body["name"] = name
		}
		if cmd.Flags().Changed("plan") {
			plan, _ := cmd.Flags().GetString("plan")
			body["plan"] = plan
		}

		if len(body) == 0 {
			return fmt.Errorf("at least one of --name or --plan must be provided")
		}

		var result any
		if err := apiPut("/api/v1/organizations/"+args[0], body, &result); err != nil {
			return err
		}
		if outputFmt == "table" {
			printOrgTable(result)
		} else {
			printJSON(result)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(orgCmd)
	orgCmd.AddCommand(orgGetCmd)
	orgCmd.AddCommand(orgCreateCmd)
	orgCmd.AddCommand(orgUpdateCmd)

	orgCreateCmd.Flags().String("name", "", "Organization name (required)")
	orgCreateCmd.Flags().String("slug", "", "Organization slug (required)")
	orgCreateCmd.Flags().String("plan", "free", "Plan: free, pro, team, enterprise")

	orgUpdateCmd.Flags().String("name", "", "Organization name")
	orgUpdateCmd.Flags().String("plan", "", "Plan: free, pro, team, enterprise")
}
