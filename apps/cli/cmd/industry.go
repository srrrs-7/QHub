package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var industryCmd = &cobra.Command{
	Use:   "industry",
	Short: "Manage industry configurations",
}

var industryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List industry configurations",
	RunE: func(_ *cobra.Command, _ []string) error {
		var result any
		if err := apiGet("/api/v1/industries", &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var industryGetCmd = &cobra.Command{
	Use:   "get <slug>",
	Short: "Get industry configuration by slug",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var result any
		if err := apiGet("/api/v1/industries/"+args[0], &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var industryCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an industry configuration",
	RunE: func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")
		desc, _ := cmd.Flags().GetString("description")

		if name == "" || slug == "" {
			return fmt.Errorf("--name and --slug are required")
		}

		body := map[string]string{
			"name":        name,
			"slug":        slug,
			"description": desc,
		}
		var result any
		if err := apiPost("/api/v1/industries", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var industryUpdateCmd = &cobra.Command{
	Use:   "update <slug>",
	Short: "Update an industry configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{}
		if name, _ := cmd.Flags().GetString("name"); name != "" {
			body["name"] = name
		}
		if desc, _ := cmd.Flags().GetString("description"); desc != "" {
			body["description"] = desc
		}

		var result any
		if err := apiPut("/api/v1/industries/"+args[0], body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var industryBenchmarksCmd = &cobra.Command{
	Use:   "benchmarks <slug>",
	Short: "List benchmarks for an industry",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var result any
		if err := apiGet("/api/v1/industries/"+args[0]+"/benchmarks", &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var industryComplianceCmd = &cobra.Command{
	Use:   "compliance-check <slug>",
	Short: "Run compliance check against industry rules",
	Long:  "Check prompt content against industry compliance rules. Provide content via --content or --file.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		content, _ := cmd.Flags().GetString("content")
		file, _ := cmd.Flags().GetString("file")

		if file != "" {
			var r io.Reader = os.Stdin
			if file != "-" {
				f, err := os.Open(file)
				if err != nil {
					return fmt.Errorf("opening file: %w", err)
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
			return fmt.Errorf("--content or --file is required")
		}

		body := map[string]string{"content": content}
		var result any
		if err := apiPost("/api/v1/industries/"+args[0]+"/compliance-check", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(industryCmd)
	industryCmd.AddCommand(industryListCmd, industryGetCmd, industryCreateCmd, industryUpdateCmd, industryBenchmarksCmd, industryComplianceCmd)

	industryCreateCmd.Flags().String("name", "", "Industry name (required)")
	industryCreateCmd.Flags().String("slug", "", "Industry slug (required)")
	industryCreateCmd.Flags().String("description", "", "Industry description")

	industryUpdateCmd.Flags().String("name", "", "Industry name")
	industryUpdateCmd.Flags().String("description", "", "Industry description")

	industryComplianceCmd.Flags().String("content", "", "Content to check")
	industryComplianceCmd.Flags().String("file", "", "Read content from file (use - for stdin)")
}
