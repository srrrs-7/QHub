package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var memberOrgID string

var memberCmd = &cobra.Command{
	Use:   "member",
	Short: "Manage organization members (list, add, update, remove)",
	Long:  "Manage members of an organization including role assignments (owner, admin, member, viewer).",
	Example: `  # List members
  qhub member --org <org-id> list

  # Add a member
  qhub member --org <org-id> add --user <user-id> --role admin

  # Update a member's role
  qhub member --org <org-id> update <user-id> --role viewer

  # Remove a member
  qhub member --org <org-id> remove <user-id>`,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		if memberOrgID == "" {
			return fmt.Errorf("--org is required")
		}
		return nil
	},
}

func memberPath(parts ...string) string {
	path := "/api/v1/organizations/" + memberOrgID + "/members"
	if len(parts) > 0 {
		path += "/" + parts[0]
	}
	return path
}

var memberListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all members of an organization",
	Example: "  qhub member --org <org-id> list",
	RunE: func(_ *cobra.Command, _ []string) error {
		var result any
		if err := apiGet(memberPath(), &result); err != nil {
			return err
		}
		if outputFmt == "table" {
			printMemberTable(result)
		} else {
			printJSON(result)
		}
		return nil
	},
}

var memberAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new member to the organization",
	Example: `  qhub member --org <org-id> add --user <user-id>
  qhub member --org <org-id> add --user <user-id> --role admin`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user")
		role, _ := cmd.Flags().GetString("role")

		if userID == "" {
			return fmt.Errorf("--user is required")
		}

		body := map[string]string{"user_id": userID, "role": role}
		var result any
		if err := apiPost(memberPath(), body, &result); err != nil {
			return err
		}
		printSuccess("Added member to organization")
		if outputFmt == "table" {
			printMemberTable(result)
		} else {
			printJSON(result)
		}
		return nil
	},
}

var memberUpdateCmd = &cobra.Command{
	Use:     "update <user-id>",
	Short:   "Update a member's role in the organization",
	Example: "  qhub member --org <org-id> update <user-id> --role admin",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		role, _ := cmd.Flags().GetString("role")
		if role == "" {
			return fmt.Errorf("--role is required")
		}

		body := map[string]string{"role": role}
		var result any
		if err := apiPut(memberPath(args[0]), body, &result); err != nil {
			return err
		}
		if outputFmt == "table" {
			printMemberTable(result)
		} else {
			printJSON(result)
		}
		return nil
	},
}

var memberRemoveCmd = &cobra.Command{
	Use:     "remove <user-id>",
	Short:   "Remove a member from the organization",
	Example: "  qhub member --org <org-id> remove <user-id>",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		if err := apiDelete(memberPath(args[0])); err != nil {
			return err
		}
		printSuccess("Removed member from organization")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(memberCmd)
	memberCmd.PersistentFlags().StringVar(&memberOrgID, "org", "", "Organization ID (required)")

	memberCmd.AddCommand(memberListCmd, memberAddCmd, memberUpdateCmd, memberRemoveCmd)

	memberAddCmd.Flags().String("user", "", "User ID (required)")
	memberAddCmd.Flags().String("role", "member", "Role: owner, admin, member, viewer")

	memberUpdateCmd.Flags().String("role", "", "New role (required)")
}
