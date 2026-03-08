package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var promptProjectID string

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Manage prompts",
}

var promptListCmd = &cobra.Command{
	Use:   "list",
	Short: "List prompts in a project",
	RunE: func(_ *cobra.Command, _ []string) error {
		if promptProjectID == "" {
			return fmt.Errorf("--project is required")
		}
		var prompts any
		if err := apiGet("/api/v1/projects/"+promptProjectID+"/prompts", &prompts); err != nil {
			return err
		}
		printJSON(prompts)
		return nil
	},
}

var promptGetCmd = &cobra.Command{
	Use:   "get <slug>",
	Short: "Get prompt details",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		if promptProjectID == "" {
			return fmt.Errorf("--project is required")
		}
		var prompt any
		if err := apiGet("/api/v1/projects/"+promptProjectID+"/prompts/"+args[0], &prompt); err != nil {
			return err
		}
		printJSON(prompt)
		return nil
	},
}

var promptCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a prompt",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if promptProjectID == "" {
			return fmt.Errorf("--project is required")
		}
		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")
		promptType, _ := cmd.Flags().GetString("type")
		desc, _ := cmd.Flags().GetString("description")

		if name == "" || slug == "" {
			return fmt.Errorf("--name and --slug are required")
		}

		body := map[string]string{
			"project_id":  promptProjectID,
			"name":        name,
			"slug":        slug,
			"prompt_type": promptType,
			"description": desc,
		}
		var result any
		if err := apiPost("/api/v1/projects/"+promptProjectID+"/prompts", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(promptCmd)
	promptCmd.PersistentFlags().StringVar(&promptProjectID, "project", "", "Project ID (required)")

	promptCmd.AddCommand(promptListCmd)
	promptCmd.AddCommand(promptGetCmd)
	promptCmd.AddCommand(promptCreateCmd)

	promptCreateCmd.Flags().String("name", "", "Prompt name (required)")
	promptCreateCmd.Flags().String("slug", "", "Prompt slug (required)")
	promptCreateCmd.Flags().String("type", "system", "Prompt type: system, user, combined")
	promptCreateCmd.Flags().String("description", "", "Prompt description")
}
