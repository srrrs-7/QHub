package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var projectOrgID string

var projectCmd = &cobra.Command{
	Use:     "project",
	Aliases: []string{"proj"},
	Short:   "Manage projects (list, create, view, update, delete)",
	Long:    "Manage projects within an organization. Projects group related prompts together.",
	Example: `  # List all projects in an organization
  qhub project --org <org-id> list

  # Create a new project
  qhub project --org <org-id> create --name "Chatbot" --slug chatbot

  # Get project details
  qhub project --org <org-id> get chatbot

  # Delete a project
  qhub project --org <org-id> delete chatbot`,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		if projectOrgID == "" {
			return fmt.Errorf("--org is required")
		}
		return nil
	},
}

var projectListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all projects in an organization",
	Example: "  qhub project --org <org-id> list",
	RunE: func(_ *cobra.Command, _ []string) error {
		var projects any
		if err := apiGet("/api/v1/organizations/"+projectOrgID+"/projects", &projects); err != nil {
			return err
		}
		if outputFmt == "table" {
			printProjectTable(projects)
		} else {
			printJSON(projects)
		}
		return nil
	},
}

var projectGetCmd = &cobra.Command{
	Use:     "get <slug>",
	Short:   "Get project details by slug",
	Example: "  qhub project --org <org-id> get chatbot",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var project any
		if err := apiGet("/api/v1/organizations/"+projectOrgID+"/projects/"+args[0], &project); err != nil {
			return err
		}
		if outputFmt == "table" {
			printProjectTable(project)
		} else {
			printJSON(project)
		}
		return nil
	},
}

var projectCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project in an organization",
	Example: `  qhub project --org <org-id> create --name "Chatbot" --slug chatbot
  qhub project --org <org-id> create --name "Support" --slug support --description "Customer support prompts"`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")
		desc, _ := cmd.Flags().GetString("description")

		if name == "" || slug == "" {
			return fmt.Errorf("--name and --slug are required")
		}

		body := map[string]string{
			"organization_id": projectOrgID, "name": name, "slug": slug, "description": desc,
		}
		var result any
		if err := apiPost("/api/v1/organizations/"+projectOrgID+"/projects", body, &result); err != nil {
			return err
		}
		printSuccess("Created project '" + slug + "'")
		if outputFmt == "table" {
			printProjectTable(result)
		} else {
			printJSON(result)
		}
		return nil
	},
}

var projectDeleteCmd = &cobra.Command{
	Use:     "delete <slug>",
	Short:   "Delete a project by slug",
	Example: "  qhub project --org <org-id> delete chatbot",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		if err := apiDelete("/api/v1/organizations/" + projectOrgID + "/projects/" + args[0]); err != nil {
			return err
		}
		printSuccess("Deleted project '" + args[0] + "'")
		return nil
	},
}

var projectUpdateCmd = &cobra.Command{
	Use:     "update <slug>",
	Short:   "Update an existing project",
	Example: "  qhub project --org <org-id> update chatbot --name \"AI Chatbot\" --description \"Updated description\"",
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
		if err := apiPut("/api/v1/organizations/"+projectOrgID+"/projects/"+args[0], body, &result); err != nil {
			return err
		}
		if outputFmt == "table" {
			printProjectTable(result)
		} else {
			printJSON(result)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.PersistentFlags().StringVar(&projectOrgID, "org", "", "Organization ID (required)")

	projectCmd.AddCommand(projectListCmd, projectGetCmd, projectCreateCmd, projectDeleteCmd, projectUpdateCmd)

	projectCreateCmd.Flags().String("name", "", "Project name (required)")
	projectCreateCmd.Flags().String("slug", "", "Project slug (required)")
	projectCreateCmd.Flags().String("description", "", "Project description")

	projectUpdateCmd.Flags().String("name", "", "Project name")
	projectUpdateCmd.Flags().String("description", "", "Project description")
}
