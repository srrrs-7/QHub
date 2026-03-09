package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var apikeyOrgID string

var apikeyCmd = &cobra.Command{
	Use:     "apikey",
	Aliases: []string{"api-key"},
	Short:   "Manage API keys (list, create, revoke)",
	Long:    "Manage API keys for an organization. API keys provide programmatic access to the QHub API.",
	Example: `  # List API keys
  qhub apikey --org <org-id> list

  # Create a new API key
  qhub apikey --org <org-id> create --name "CI Pipeline"

  # Revoke an API key
  qhub apikey --org <org-id> delete <key-id>`,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		if apikeyOrgID == "" {
			return fmt.Errorf("--org is required")
		}
		return nil
	},
}

func apikeyPath(parts ...string) string {
	path := "/api/v1/organizations/" + apikeyOrgID + "/api-keys"
	if len(parts) > 0 {
		path += "/" + parts[0]
	}
	return path
}

var apikeyListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all API keys for an organization",
	Example: "  qhub apikey --org <org-id> list",
	RunE: func(_ *cobra.Command, _ []string) error {
		var result any
		if err := apiGet(apikeyPath(), &result); err != nil {
			return err
		}
		if outputFmt == "table" {
			printAPIKeyTable(result)
		} else {
			printJSON(result)
		}
		return nil
	},
}

var apikeyCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new API key for an organization",
	Example: "  qhub apikey --org <org-id> create --name \"CI Pipeline\"",
	RunE: func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}

		body := map[string]string{
			"organization_id": apikeyOrgID,
			"name":            name,
		}
		var result any
		if err := apiPost(apikeyPath(), body, &result); err != nil {
			return err
		}
		printSuccess("Created API key '" + name + "'")
		if outputFmt == "table" {
			printAPIKeyTable(result)
		} else {
			printJSON(result)
		}
		return nil
	},
}

var apikeyDeleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Short:   "Revoke an API key by ID",
	Example: "  qhub apikey --org <org-id> delete <key-id>",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		if err := apiDelete(apikeyPath(args[0])); err != nil {
			return err
		}
		printSuccess("Revoked API key")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(apikeyCmd)
	apikeyCmd.PersistentFlags().StringVar(&apikeyOrgID, "org", "", "Organization ID (required)")

	apikeyCmd.AddCommand(apikeyListCmd, apikeyCreateCmd, apikeyDeleteCmd)

	apikeyCreateCmd.Flags().String("name", "", "API key name (required)")
}
