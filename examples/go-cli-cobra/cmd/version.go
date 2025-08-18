{{- if .include_version}}
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is set during build time
	Version = "dev"
	// Commit is set during build time
	Commit = "unknown"
	// Date is set during build time
	Date = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of {{.project_name}}",
	Long:  `All software has versions. This is {{.project_name}}'s`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("{{.project_name}} %s (%s) built on %s\n", Version, Commit, Date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
{{- end}}