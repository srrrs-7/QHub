package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var promptProjectID string

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Manage prompts (list, create, view, update)",
	Long:  "Manage prompts within a project. Prompts are templates that can have multiple versions.",
	Example: `  # List prompts in a project
  qhub prompt --project <proj-id> list

  # Create a new prompt
  qhub prompt --project <proj-id> create --name "Greeting" --slug greeting --type system

  # Get prompt details
  qhub prompt --project <proj-id> get greeting`,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		if promptProjectID == "" {
			return fmt.Errorf("--project is required")
		}
		return nil
	},
}

var promptListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all prompts in a project",
	Example: "  qhub prompt --project <proj-id> list",
	RunE: func(_ *cobra.Command, _ []string) error {
		var prompts any
		if err := apiGet("/api/v1/projects/"+promptProjectID+"/prompts", &prompts); err != nil {
			return err
		}
		if outputFmt == "table" {
			printPromptTable(prompts)
		} else {
			printJSON(prompts)
		}
		return nil
	},
}

var promptGetCmd = &cobra.Command{
	Use:     "get <slug>",
	Short:   "Get prompt details by slug",
	Example: "  qhub prompt --project <proj-id> get greeting",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var prompt any
		if err := apiGet("/api/v1/projects/"+promptProjectID+"/prompts/"+args[0], &prompt); err != nil {
			return err
		}
		if outputFmt == "table" {
			printPromptTable(prompt)
		} else {
			printJSON(prompt)
		}
		return nil
	},
}

var promptCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new prompt in a project",
	Example: `  qhub prompt --project <proj-id> create --name "Greeting" --slug greeting
  qhub prompt --project <proj-id> create --name "Summary" --slug summary --type user --description "Summarization prompt"`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")
		promptType, _ := cmd.Flags().GetString("type")
		desc, _ := cmd.Flags().GetString("description")

		if name == "" || slug == "" {
			return fmt.Errorf("--name and --slug are required")
		}

		body := map[string]string{
			"project_id": promptProjectID, "name": name, "slug": slug,
			"prompt_type": promptType, "description": desc,
		}
		var result any
		if err := apiPost("/api/v1/projects/"+promptProjectID+"/prompts", body, &result); err != nil {
			return err
		}
		printSuccess("Created prompt '" + slug + "'")
		if outputFmt == "table" {
			printPromptTable(result)
		} else {
			printJSON(result)
		}
		return nil
	},
}

var promptUpdateCmd = &cobra.Command{
	Use:     "update <slug>",
	Short:   "Update an existing prompt",
	Example: "  qhub prompt --project <proj-id> update greeting --name \"Welcome Greeting\"",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]string{}

		if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			body["name"] = name
		}
		if cmd.Flags().Changed("description") {
			desc, _ := cmd.Flags().GetString("description")
			body["description"] = desc
		}

		if len(body) == 0 {
			return fmt.Errorf("at least one of --name or --description must be provided")
		}

		var result any
		if err := apiPut("/api/v1/projects/"+promptProjectID+"/prompts/"+args[0], body, &result); err != nil {
			return err
		}
		if outputFmt == "table" {
			printPromptTable(result)
		} else {
			printJSON(result)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(promptCmd)
	promptCmd.PersistentFlags().StringVar(&promptProjectID, "project", "", "Project ID (required)")

	promptCmd.AddCommand(promptListCmd, promptGetCmd, promptCreateCmd, promptUpdateCmd)

	promptCreateCmd.Flags().String("name", "", "Prompt name (required)")
	promptCreateCmd.Flags().String("slug", "", "Prompt slug (required)")
	promptCreateCmd.Flags().String("type", "system", "Prompt type: system, user, combined")
	promptCreateCmd.Flags().String("description", "", "Prompt description")

	promptUpdateCmd.Flags().String("name", "", "Prompt name")
	promptUpdateCmd.Flags().String("description", "", "Prompt description")
}
