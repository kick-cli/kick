package cmd

import (
    "fmt"
    "strings"

    "github.com/spf13/cobra"
    "{{.module_name}}/internal/greet"
)

func newHelloCmd() *cobra.Command {
	var uppercase bool

	cmd := &cobra.Command{
		Use:   "hello [name]",
		Short: "Print a friendly greeting",
		Args:  cobra.MaximumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
			name := "World"
			if len(args) > 0 {
				name = args[0]
			}
			msg := greet.Greeting(name)
			if uppercase {
				msg = strings.ToUpper(msg)
			}
            _, err := fmt.Fprintln(cmd.OutOrStdout(), msg)
            return err
		},
	}

	cmd.Flags().BoolVarP(&uppercase, "uppercase", "u", false, "print in uppercase")
	return cmd
}
