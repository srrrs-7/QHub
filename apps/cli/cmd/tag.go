package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage tags",
}

var tagListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tags",
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
	Use:   "create",
	Short: "Create a tag",
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
	Use:   "delete <id>",
	Short: "Delete a tag",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		if err := apiDelete("/api/v1/tags/" + args[0]); err != nil {
			return err
		}
		fmt.Println("Tag deleted.")
		return nil
	},
}

var tagAddCmd = &cobra.Command{
	Use:   "add <prompt-id> <tag-id>",
	Short: "Add a tag to a prompt",
	Args:  cobra.ExactArgs(2),
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
	Use:   "remove <prompt-id> <tag-id>",
	Short: "Remove a tag from a prompt",
	Args:  cobra.ExactArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		if err := apiDelete("/api/v1/prompts/" + args[0] + "/tags/" + args[1]); err != nil {
			return err
		}
		fmt.Println("Tag removed.")
		return nil
	},
}

var tagListByPromptCmd = &cobra.Command{
	Use:   "list-by-prompt <prompt-id>",
	Short: "List tags for a prompt",
	Args:  cobra.ExactArgs(1),
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
