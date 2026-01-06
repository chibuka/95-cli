package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/chibuka/95-cli/client"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with GitHub OAuth",
	Long: `Authenticate with 95 using your GitHub account.

This command opens your browser to complete GitHub OAuth authentication.
Your credentials are securely stored locally for future commands.

Example:
  95 login`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Display colorful ASCII art logo
		green := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
		orange := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true)
		gray := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

		fmt.Println()
		fmt.Println(orange.Render(" █████╗ ███████╗"))
		fmt.Println(orange.Render("██╔══██╗██╔════╝"))
		fmt.Println(orange.Render("╚██████║███████╗"))
		fmt.Println(orange.Render(" ╚═══██║╚════██║"))
		fmt.Println(orange.Render(" █████╔╝███████║"))
		fmt.Println(orange.Render(" ╚════╝ ╚══════╝"))
		fmt.Println()
		fmt.Println(gray.Render("Build your coding skills, one challenge at a time"))
		fmt.Println()

		err := client.Login()
		if err != nil {
			fmt.Println(orange.Render("✗ Failed to login: " + err.Error()))
			return err
		}

		fmt.Println(green.Render("✓ Logged in successfully!"))
		fmt.Println(gray.Render("You're ready to start coding!"))
		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
