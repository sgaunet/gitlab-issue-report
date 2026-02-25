package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// whoamiCmd represents the whoami command.
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display information about the authenticated GitLab user",
	Long:  `Display information about the authenticated GitLab user including username, full name, email, and user ID.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Setup environment (validates GITLAB_TOKEN and sets default GITLAB_URI)
		if err := setupEnvironment(); err != nil {
			return err
		}

		// Apply timeout from environment variable if flag not set
		applyTimeoutFromEnv(&opts, cmd.Flags().Changed("api-timeout"))

		// Create GitLab client
		gitlabClient, err := createGitlabClient(os.Getenv("GITLAB_TOKEN"), os.Getenv("GITLAB_URI"), opts.apiTimeout)
		if err != nil {
			return err
		}

		// Fetch current user information
		user, _, err := gitlabClient.Users.CurrentUser()
		if err != nil {
			return fmt.Errorf("failed to fetch user information: %w", err)
		}

		// Display user information
		displayUserInfo(user)
		return nil
	},
}

// displayUserInfo formats and displays user information.
func displayUserInfo(user *gitlab.User) {
	fmt.Printf("Username: %s\n", user.Username)
	fmt.Printf("Full Name: %s\n", user.Name)
	fmt.Printf("Email: %s\n", user.Email)
	fmt.Printf("User ID: %d\n", user.ID)
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
