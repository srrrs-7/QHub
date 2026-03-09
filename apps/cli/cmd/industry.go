package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var industryCmd = &cobra.Command{
	Use:   "industry",
	Short: "Manage industry configurations and compliance",
	Long:  "Manage industry-specific configurations including compliance rules and benchmarks for prompt quality.",
	Example: `  # List available industries
  qhub industry list

  # Get industry details
  qhub industry get healthcare

  # Create an industry configuration
  qhub industry create --name "Healthcare" --slug healthcare

  # Run compliance check
  qhub industry compliance-check healthcare --content "Your prompt text"

  # View benchmarks
  qhub industry benchmarks healthcare`,
}

var industryListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all available industry configurations",
	Example: "  qhub industry list",
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
	Use:     "get <slug>",
	Short:   "Get industry configuration details by slug",
	Example: "  qhub industry get healthcare",
	Args:    cobra.ExactArgs(1),
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
	Short: "Create a new industry configuration",
	Example: `  qhub industry create --name "Healthcare" --slug healthcare
  qhub industry create --name "Finance" --slug finance --description "Financial services compliance"`,
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
	Use:     "update <slug>",
	Short:   "Update an existing industry configuration",
	Example: "  qhub industry update healthcare --name \"Healthcare & Life Sciences\"",
	Args:    cobra.ExactArgs(1),
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
	Use:     "benchmarks <slug>",
	Short:   "List quality benchmarks for an industry",
	Example: "  qhub industry benchmarks healthcare",
	Args:    cobra.ExactArgs(1),
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
	Short: "Check prompt content against industry compliance rules",
	Long:  "Run a compliance check on prompt content against industry-specific rules. Provide content via --content or --file (use - for stdin).",
	Example: `  # Check inline content
  qhub industry compliance-check healthcare --content "Your prompt text here"

  # Check content from a file
  qhub industry compliance-check finance --file prompt.txt

  # Check content from stdin
  cat prompt.txt | qhub industry compliance-check healthcare --file -`,
	Args: cobra.ExactArgs(1),
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
