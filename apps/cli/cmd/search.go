package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Semantic search across prompts using embeddings",
	Long:  "Search prompts by semantic similarity using vector embeddings. Also check embedding service health.",
	Example: `  # Search for prompts similar to a query
  qhub search semantic "customer support greeting" --org <org-id>

  # Search with minimum similarity threshold
  qhub search semantic "error handling" --org <org-id> --min-score 0.7 --limit 5

  # Check embedding service status
  qhub search status`,
}

var searchSemanticCmd = &cobra.Command{
	Use:   "semantic <query>",
	Short: "Search prompts by semantic similarity to a query",
	Example: `  qhub search semantic "customer support greeting" --org <org-id>
  qhub search semantic "error handling" --org <org-id> --min-score 0.7 --limit 5`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		orgID, _ := cmd.Flags().GetString("org")
		limit, _ := cmd.Flags().GetInt("limit")
		minScore, _ := cmd.Flags().GetFloat64("min-score")

		if orgID == "" {
			return fmt.Errorf("--org is required")
		}

		body := map[string]any{
			"query":     args[0],
			"org_id":    orgID,
			"limit":     limit,
			"min_score": minScore,
		}

		var result any
		if err := apiPost("/api/v1/search/semantic", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var searchStatusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Check embedding service health and availability",
	Example: "  qhub search status",
	RunE: func(_ *cobra.Command, _ []string) error {
		var result any
		if err := apiGet("/api/v1/search/embedding-status", &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.AddCommand(searchSemanticCmd, searchStatusCmd)

	searchSemanticCmd.Flags().String("org", "", "Organization ID (required)")
	searchSemanticCmd.Flags().Int("limit", 10, "Maximum results")
	searchSemanticCmd.Flags().Float64("min-score", 0.0, "Minimum similarity score (0.0-1.0)")
}
