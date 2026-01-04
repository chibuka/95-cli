package cmd

import (
	"fmt"

	"github.com/chibuka/95-cli/client"
	"github.com/chibuka/95-cli/internal/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear local credentials and logout",
	Long: `Clear your local authentication credentials.

This removes your stored access token and refresh token from the local config.
You'll need to run '95 login' again to authenticate.

Example:
  95 logout`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("could not load config file: %w", err)
		}

		if cfg.AccessToken == "" {
			fmt.Println("Already logged out")
			return nil
		}

		apiURL := cfg.GetAPIURL()
		err = client.Logout(cfg.AccessToken, apiURL)
		if err != nil {
			fmt.Println("⚠ Could not notify server, but clearing local credentials...")
		}

		err = config.Clear()
		if err != nil {
			return fmt.Errorf("failed to clear credentials: %w", err)
		}

		fmt.Println("✓ Logged out successfully!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
