package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage tags (create, delete, assign to prompts)",
	Long:  "Create, delete, and assign tags to prompts for organization and filtering.",
	Example: `  # List all tags
  qhub tag list --org <org-id>

  # Create a tag
  qhub tag create --org <org-id> --name "production" --color green

  # Add a tag to a prompt
  qhub tag add <prompt-id> <tag-id>

  # List tags on a prompt
  qhub tag list-by-prompt <prompt-id>`,
}

var tagListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tags, optionally filtered by organization",
	Example: `  qhub tag list
  qhub tag list --org <org-id>`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		orgID, _ := cmd.Flags().GetString("org")
		path := "/api/v1/tags"
		if orgID != "" {
			path += "?org_id=" + orgID
		}
		var result any
		if err := apiGet(path, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var tagCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new tag in an organization",
	Example: `  qhub tag create --org <org-id> --name "production" --color green`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		orgID, _ := cmd.Flags().GetString("org")
		name, _ := cmd.Flags().GetString("name")
		color, _ := cmd.Flags().GetString("color")

		if orgID == "" || name == "" {
			return fmt.Errorf("--org and --name are required")
		}

		body := map[string]string{
			"org_id": orgID,
			"name":   name,
			"color":  color,
		}
		var result any
		if err := apiPost("/api/v1/tags", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var tagDeleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Short:   "Delete a tag by ID",
	Example: "  qhub tag delete <tag-id>",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		if err := apiDelete("/api/v1/tags/" + args[0]); err != nil {
			return err
		}
		fmt.Println("Tag deleted.")
		return nil
	},
}

var tagAddCmd = &cobra.Command{
	Use:     "add <prompt-id> <tag-id>",
	Short:   "Assign a tag to a prompt",
	Example: "  qhub tag add <prompt-id> <tag-id>",
	Args:    cobra.ExactArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		body := map[string]string{"tag_id": args[1]}
		var result any
		if err := apiPost("/api/v1/prompts/"+args[0]+"/tags", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var tagRemoveCmd = &cobra.Command{
	Use:     "remove <prompt-id> <tag-id>",
	Short:   "Remove a tag from a prompt",
	Example: "  qhub tag remove <prompt-id> <tag-id>",
	Args:    cobra.ExactArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		if err := apiDelete("/api/v1/prompts/" + args[0] + "/tags/" + args[1]); err != nil {
			return err
		}
		fmt.Println("Tag removed.")
		return nil
	},
}

var tagListByPromptCmd = &cobra.Command{
	Use:     "list-by-prompt <prompt-id>",
	Short:   "List all tags assigned to a specific prompt",
	Example: "  qhub tag list-by-prompt <prompt-id>",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var result any
		if err := apiGet("/api/v1/prompts/"+args[0]+"/tags", &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)
	tagCmd.AddCommand(tagListCmd, tagCreateCmd, tagDeleteCmd, tagAddCmd, tagRemoveCmd, tagListByPromptCmd)

	tagListCmd.Flags().String("org", "", "Filter by organization ID")

	tagCreateCmd.Flags().String("org", "", "Organization ID (required)")
	tagCreateCmd.Flags().String("name", "", "Tag name (required)")
	tagCreateCmd.Flags().String("color", "blue", "Tag color")
}
