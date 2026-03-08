package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var versionPromptID string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage prompt versions",
}

var versionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List versions of a prompt",
	RunE: func(_ *cobra.Command, _ []string) error {
		if versionPromptID == "" {
			return fmt.Errorf("--prompt is required")
		}
		var versions any
		if err := apiGet("/api/v1/prompts/"+versionPromptID+"/versions", &versions); err != nil {
			return err
		}
		printJSON(versions)
		return nil
	},
}

var versionGetCmd = &cobra.Command{
	Use:   "get <number|latest|production>",
	Short: "Get a specific version",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		if versionPromptID == "" {
			return fmt.Errorf("--prompt is required")
		}
		var version any
		if err := apiGet("/api/v1/prompts/"+versionPromptID+"/versions/"+args[0], &version); err != nil {
			return err
		}
		printJSON(version)
		return nil
	},
}

var versionCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new version",
	Long:  `Create a new version. Provide content via --content flag or --content-file to read from file (use - for stdin).`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if versionPromptID == "" {
			return fmt.Errorf("--prompt is required")
		}

		content, _ := cmd.Flags().GetString("content")
		contentFile, _ := cmd.Flags().GetString("content-file")
		changeDesc, _ := cmd.Flags().GetString("change-description")
		variables, _ := cmd.Flags().GetStringSlice("variables")

		// Read content from file or stdin
		if contentFile != "" {
			var data []byte
			var err error
			if contentFile == "-" {
				data, err = os.ReadFile("/dev/stdin")
			} else {
				data, err = os.ReadFile(contentFile)
			}
			if err != nil {
				return fmt.Errorf("reading content file: %w", err)
			}
			content = string(data)
		}

		if content == "" {
			return fmt.Errorf("--content or --content-file is required")
		}

		contentJSON, _ := json.Marshal(content)
		body := map[string]any{
			"content":            json.RawMessage(contentJSON),
			"change_description": changeDesc,
		}
		if len(variables) > 0 {
			varsJSON, _ := json.Marshal(variables)
			body["variables"] = json.RawMessage(varsJSON)
		}

		var result any
		if err := apiPost("/api/v1/prompts/"+versionPromptID+"/versions", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var versionPromoteCmd = &cobra.Command{
	Use:   "promote <version-number>",
	Short: "Promote a version (draft→review→production)",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		if versionPromptID == "" {
			return fmt.Errorf("--prompt is required")
		}

		// Get current version to determine next status
		var current map[string]any
		if err := apiGet("/api/v1/prompts/"+versionPromptID+"/versions/"+args[0], &current); err != nil {
			return err
		}

		status, _ := current["status"].(string)
		var nextStatus string
		switch status {
		case "draft":
			nextStatus = "review"
		case "review":
			nextStatus = "production"
		default:
			return fmt.Errorf("cannot promote version in '%s' status", status)
		}

		var result any
		if err := apiPut("/api/v1/prompts/"+versionPromptID+"/versions/"+args[0]+"/status",
			map[string]string{"status": nextStatus}, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var versionArchiveCmd = &cobra.Command{
	Use:   "archive <version-number>",
	Short: "Archive a version",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		if versionPromptID == "" {
			return fmt.Errorf("--prompt is required")
		}

		var result any
		if err := apiPut("/api/v1/prompts/"+versionPromptID+"/versions/"+args[0]+"/status",
			map[string]string{"status": "archived"}, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var versionStatusCmd = &cobra.Command{
	Use:   "status <version-number> <status>",
	Short: "Set version status (draft, review, production, archived)",
	Args:  cobra.ExactArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		if versionPromptID == "" {
			return fmt.Errorf("--prompt is required")
		}

		var result any
		if err := apiPut("/api/v1/prompts/"+versionPromptID+"/versions/"+args[0]+"/status",
			map[string]string{"status": args[1]}, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.PersistentFlags().StringVar(&versionPromptID, "prompt", "", "Prompt ID (required)")

	versionCmd.AddCommand(versionListCmd)
	versionCmd.AddCommand(versionGetCmd)
	versionCmd.AddCommand(versionCreateCmd)
	versionCmd.AddCommand(versionPromoteCmd)
	versionCmd.AddCommand(versionArchiveCmd)
	versionCmd.AddCommand(versionStatusCmd)

	versionCreateCmd.Flags().String("content", "", "Prompt content")
	versionCreateCmd.Flags().String("content-file", "", "Read content from file (use - for stdin)")
	versionCreateCmd.Flags().String("change-description", "", "Description of changes")
	versionCreateCmd.Flags().StringSlice("variables", nil, "Template variables (comma-separated)")
}
