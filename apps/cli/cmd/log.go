package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Manage execution logs",
}

var logListCmd = &cobra.Command{
	Use:   "list",
	Short: "List execution logs",
	RunE: func(cmd *cobra.Command, _ []string) error {
		orgID, _ := cmd.Flags().GetString("org")
		promptID, _ := cmd.Flags().GetString("prompt")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		path := fmt.Sprintf("/api/v1/logs?limit=%d&offset=%d", limit, offset)
		if orgID != "" {
			path += "&org_id=" + orgID
		}
		if promptID != "" {
			path += "&prompt_id=" + promptID
		}

		var result any
		if err := apiGet(path, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var logGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get execution log details",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var result any
		if err := apiGet("/api/v1/logs/"+args[0], &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var logCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an execution log",
	Long:  "Create an execution log entry. Provide request/response bodies via flags or --file for full JSON.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		file, _ := cmd.Flags().GetString("file")

		var body map[string]any
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
			if err := json.NewDecoder(r).Decode(&body); err != nil {
				return fmt.Errorf("decoding JSON: %w", err)
			}
		} else {
			orgID, _ := cmd.Flags().GetString("org")
			promptID, _ := cmd.Flags().GetString("prompt")
			versionNum, _ := cmd.Flags().GetInt("version")
			model, _ := cmd.Flags().GetString("model")
			provider, _ := cmd.Flags().GetString("provider")
			inputTokens, _ := cmd.Flags().GetInt("input-tokens")
			outputTokens, _ := cmd.Flags().GetInt("output-tokens")
			latency, _ := cmd.Flags().GetInt("latency-ms")
			cost, _ := cmd.Flags().GetString("estimated-cost")
			status, _ := cmd.Flags().GetString("status")
			env, _ := cmd.Flags().GetString("env")
			executedAt, _ := cmd.Flags().GetString("executed-at")

			if orgID == "" || promptID == "" {
				return fmt.Errorf("--org and --prompt are required")
			}

			body = map[string]any{
				"org_id":         orgID,
				"prompt_id":      promptID,
				"version_number": versionNum,
				"model":          model,
				"provider":       provider,
				"input_tokens":   inputTokens,
				"output_tokens":  outputTokens,
				"total_tokens":   inputTokens + outputTokens,
				"latency_ms":     latency,
				"estimated_cost": cost,
				"status":         status,
				"environment":    env,
				"request_body":   json.RawMessage(`{}`),
				"executed_at":    executedAt,
			}
		}

		var result any
		if err := apiPost("/api/v1/logs", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
	logCmd.AddCommand(logListCmd, logGetCmd, logCreateCmd)

	logListCmd.Flags().String("org", "", "Filter by organization ID")
	logListCmd.Flags().String("prompt", "", "Filter by prompt ID")
	logListCmd.Flags().Int("limit", 20, "Number of results")
	logListCmd.Flags().Int("offset", 0, "Offset for pagination")

	logCreateCmd.Flags().String("file", "", "JSON file with log data (use - for stdin)")
	logCreateCmd.Flags().String("org", "", "Organization ID")
	logCreateCmd.Flags().String("prompt", "", "Prompt ID")
	logCreateCmd.Flags().Int("version", 1, "Version number")
	logCreateCmd.Flags().String("model", "", "LLM model name")
	logCreateCmd.Flags().String("provider", "", "LLM provider")
	logCreateCmd.Flags().Int("input-tokens", 0, "Input token count")
	logCreateCmd.Flags().Int("output-tokens", 0, "Output token count")
	logCreateCmd.Flags().Int("latency-ms", 0, "Latency in milliseconds")
	logCreateCmd.Flags().String("estimated-cost", "0", "Estimated cost")
	logCreateCmd.Flags().String("status", "success", "Status: success, error")
	logCreateCmd.Flags().String("env", "development", "Environment: development, staging, production")
	logCreateCmd.Flags().String("executed-at", "", "Execution timestamp (RFC3339)")
}
