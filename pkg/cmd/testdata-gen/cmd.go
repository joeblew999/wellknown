package testdatagen

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the test data generation command
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "gen-testdata",
		Short: "Generate test data for development",
		Long: `Generate sample test data for development and testing.

Populates the database with realistic test data for:
  • User accounts
  • OAuth tokens
  • Sample calendar events
  • Test collections

Use this to quickly set up a development environment with data to work with.`,
		Run: func(cmd *cobra.Command, args []string) {
			Main(args)
		},
	}
}
