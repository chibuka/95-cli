package cmd

import (
	"fmt"

	"github.com/chibuka/95-cli/client"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with GitHub OAuth",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := client.Login()
		if err != nil {
			fmt.Println("Failed to login:", err)
			return err
		}

		fmt.Println("✓ Logged in successfully!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
