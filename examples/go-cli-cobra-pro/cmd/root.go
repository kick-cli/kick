package cmd

import (
    "os"

    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
    "github.com/spf13/cobra"
)

func Root() *cobra.Command {
	var verbose bool

	root := &cobra.Command{
		Use:   "{{.project_name}}",
		Short: "{{.description}}",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			} else {
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			}
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
			return nil
		},
	}

	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")

	root.AddCommand(newHelloCmd())
	{{- if .include_version}}
	root.AddCommand(newVersionCmd())
	{{- end}}

	return root
}

// Execute runs the root command.
func Execute() {
    _ = Root().Execute()
}
