package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// whoamiCmd represents the whoami command.
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display information about the authenticated GitLab user",
	Long:  `Display information about the authenticated GitLab user including username, full name, email, and user ID.`,
	Run: func(_ *cobra.Command, _ []string) {
		// Setup environment (validates GITLAB_TOKEN and sets default GITLAB_URI)
		setupEnvironment()

		// Create GitLab client
		gitlabClient, err := gitlab.NewClient(os.Getenv("GITLAB_TOKEN"), gitlab.WithBaseURL(os.Getenv("GITLAB_URI")))
		if err != nil {
			logrus.Errorln("Failed to create GitLab client:", err.Error())
			os.Exit(1)
		}

		// Fetch current user information
		user, _, err := gitlabClient.Users.CurrentUser()
		if err != nil {
			logrus.Errorln("Failed to fetch user information:", err.Error())
			os.Exit(1)
		}

		// Display user information
		displayUserInfo(user)
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
