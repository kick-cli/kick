package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"{{.module_name}}/internal/version"
)

func newVersionCmd() *cobra.Command {
    return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
        RunE: func(cmd *cobra.Command, _ []string) error {
            _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\n", version.Human())
            return err
        },
	}
}
