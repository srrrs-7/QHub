package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var evalCmd = &cobra.Command{
	Use:     "eval",
	Aliases: []string{"evaluation"},
	Short:   "Manage evaluations",
}

var evalGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get evaluation details",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var result any
		if err := apiGet("/api/v1/evaluations/"+args[0], &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var evalListCmd = &cobra.Command{
	Use:   "list --log <log-id>",
	Short: "List evaluations for an execution log",
	RunE: func(cmd *cobra.Command, _ []string) error {
		logID, _ := cmd.Flags().GetString("log")
		if logID == "" {
			return fmt.Errorf("--log is required")
		}
		var result any
		if err := apiGet("/api/v1/logs/"+logID+"/evaluations", &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var evalCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an evaluation",
	RunE: func(cmd *cobra.Command, _ []string) error {
		logID, _ := cmd.Flags().GetString("log")
		overall, _ := cmd.Flags().GetString("overall-score")
		accuracy, _ := cmd.Flags().GetString("accuracy-score")
		relevance, _ := cmd.Flags().GetString("relevance-score")
		fluency, _ := cmd.Flags().GetString("fluency-score")
		safety, _ := cmd.Flags().GetString("safety-score")
		feedback, _ := cmd.Flags().GetString("feedback")
		evalType, _ := cmd.Flags().GetString("evaluator-type")

		if logID == "" {
			return fmt.Errorf("--log is required")
		}

		body := map[string]any{
			"execution_log_id": logID,
			"evaluator_type":   evalType,
		}
		if overall != "" {
			body["overall_score"] = overall
		}
		if accuracy != "" {
			body["accuracy_score"] = accuracy
		}
		if relevance != "" {
			body["relevance_score"] = relevance
		}
		if fluency != "" {
			body["fluency_score"] = fluency
		}
		if safety != "" {
			body["safety_score"] = safety
		}
		if feedback != "" {
			body["feedback"] = feedback
		}

		var result any
		if err := apiPost("/api/v1/evaluations", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(evalCmd)
	evalCmd.AddCommand(evalGetCmd, evalListCmd, evalCreateCmd)

	evalListCmd.Flags().String("log", "", "Execution log ID (required)")

	evalCreateCmd.Flags().String("log", "", "Execution log ID (required)")
	evalCreateCmd.Flags().String("overall-score", "", "Overall score")
	evalCreateCmd.Flags().String("accuracy-score", "", "Accuracy score")
	evalCreateCmd.Flags().String("relevance-score", "", "Relevance score")
	evalCreateCmd.Flags().String("fluency-score", "", "Fluency score")
	evalCreateCmd.Flags().String("safety-score", "", "Safety score")
	evalCreateCmd.Flags().String("feedback", "", "Feedback text")
	evalCreateCmd.Flags().String("evaluator-type", "human", "Evaluator type: human, auto")
}
