package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var versionPromptID string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage prompt versions",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		if versionPromptID == "" {
			return fmt.Errorf("--prompt is required")
		}
		return nil
	},
}

func versionPath(parts ...string) string {
	path := "/api/v1/prompts/" + versionPromptID + "/versions"
	for _, p := range parts {
		path += "/" + p
	}
	return path
}

var versionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List versions of a prompt",
	RunE: func(_ *cobra.Command, _ []string) error {
		var versions any
		if err := apiGet(versionPath(), &versions); err != nil {
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
		var version any
		if err := apiGet(versionPath(args[0]), &version); err != nil {
			return err
		}
		printJSON(version)
		return nil
	},
}

var versionCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new version",
	Long:  "Create a new version. Provide content via --content or --content-file (use - for stdin).",
	RunE: func(cmd *cobra.Command, _ []string) error {
		content, _ := cmd.Flags().GetString("content")
		contentFile, _ := cmd.Flags().GetString("content-file")
		changeDesc, _ := cmd.Flags().GetString("change-description")
		variables, _ := cmd.Flags().GetStringSlice("variables")

		if contentFile != "" {
			var r io.Reader = os.Stdin
			if contentFile != "-" {
				f, err := os.Open(contentFile)
				if err != nil {
					return fmt.Errorf("opening content file: %w", err)
				}
				defer f.Close()
				r = f
			}
			data, err := io.ReadAll(r)
			if err != nil {
				return fmt.Errorf("reading content: %w", err)
			}
			content = string(data)
		}

		if content == "" {
			return fmt.Errorf("--content or --content-file is required")
		}

		contentJSON, err := json.Marshal(content)
		if err != nil {
			return fmt.Errorf("encoding content: %w", err)
		}

		body := map[string]any{
			"content":            json.RawMessage(contentJSON),
			"change_description": changeDesc,
		}
		if len(variables) > 0 {
			varsJSON, err := json.Marshal(variables)
			if err != nil {
				return fmt.Errorf("encoding variables: %w", err)
			}
			body["variables"] = json.RawMessage(varsJSON)
		}

		var result any
		if err := apiPost(versionPath(), body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var versionPromoteCmd = &cobra.Command{
	Use:   "promote <version-number>",
	Short: "Promote a version (draft->review->production)",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var current map[string]any
		if err := apiGet(versionPath(args[0]), &current); err != nil {
			return err
		}

		status, _ := current["status"].(string)
		nextStatus := map[string]string{"draft": "review", "review": "production"}[status]
		if nextStatus == "" {
			return fmt.Errorf("cannot promote version in '%s' status", status)
		}

		var result any
		if err := apiPut(versionPath(args[0], "status"), map[string]string{"status": nextStatus}, &result); err != nil {
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
		var result any
		if err := apiPut(versionPath(args[0], "status"), map[string]string{"status": args[1]}, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.PersistentFlags().StringVar(&versionPromptID, "prompt", "", "Prompt ID (required)")

	versionCmd.AddCommand(versionListCmd, versionGetCmd, versionCreateCmd, versionPromoteCmd, versionStatusCmd)

	versionCreateCmd.Flags().String("content", "", "Prompt content")
	versionCreateCmd.Flags().String("content-file", "", "Read content from file (use - for stdin)")
	versionCreateCmd.Flags().String("change-description", "", "Description of changes")
	versionCreateCmd.Flags().StringSlice("variables", nil, "Template variables (comma-separated)")
}
