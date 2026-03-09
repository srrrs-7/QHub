package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var consultingCmd = &cobra.Command{
	Use:   "consulting",
	Short: "Manage consulting sessions and messages",
}

// --- Sessions ---

var sessionListCmd = &cobra.Command{
	Use:   "sessions",
	Short: "List consulting sessions",
	RunE: func(cmd *cobra.Command, _ []string) error {
		orgID, _ := cmd.Flags().GetString("org")
		path := "/api/v1/consulting/sessions"
		if orgID != "" {
			path += "?org_id=" + orgID
		}
		var result any
		if err := apiGet(path, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var sessionGetCmd = &cobra.Command{
	Use:   "session <id>",
	Short: "Get consulting session details",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var result any
		if err := apiGet("/api/v1/consulting/sessions/"+args[0], &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var sessionCreateCmd = &cobra.Command{
	Use:   "create-session",
	Short: "Create a consulting session",
	RunE: func(cmd *cobra.Command, _ []string) error {
		orgID, _ := cmd.Flags().GetString("org")
		title, _ := cmd.Flags().GetString("title")
		industryID, _ := cmd.Flags().GetString("industry")

		if orgID == "" || title == "" {
			return fmt.Errorf("--org and --title are required")
		}

		body := map[string]any{
			"org_id": orgID,
			"title":  title,
		}
		if industryID != "" {
			body["industry_config_id"] = industryID
		}

		var result any
		if err := apiPost("/api/v1/consulting/sessions", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

// --- Messages ---

var messageListCmd = &cobra.Command{
	Use:   "messages <session-id>",
	Short: "List messages in a consulting session",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var result any
		if err := apiGet("/api/v1/consulting/sessions/"+args[0]+"/messages", &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var messageSendCmd = &cobra.Command{
	Use:   "send <session-id> <message>",
	Short: "Send a message to a consulting session",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		role, _ := cmd.Flags().GetString("role")

		body := map[string]string{
			"role":    role,
			"content": args[1],
		}
		var result any
		if err := apiPost("/api/v1/consulting/sessions/"+args[0]+"/messages", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var sessionCloseCmd = &cobra.Command{
	Use:   "close <session-id>",
	Short: "Close a consulting session",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		body := map[string]string{"status": "closed"}
		var result any
		if err := apiPut("/api/v1/consulting/sessions/"+args[0], body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(consultingCmd)
	consultingCmd.AddCommand(sessionListCmd, sessionGetCmd, sessionCreateCmd, messageListCmd, messageSendCmd, sessionCloseCmd)

	sessionListCmd.Flags().String("org", "", "Filter by organization ID")

	sessionCreateCmd.Flags().String("org", "", "Organization ID (required)")
	sessionCreateCmd.Flags().String("title", "", "Session title (required)")
	sessionCreateCmd.Flags().String("industry", "", "Industry config ID")

	messageSendCmd.Flags().String("role", "user", "Message role: user, assistant, system")
}
