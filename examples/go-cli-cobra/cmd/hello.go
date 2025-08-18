package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// helloCmd represents the hello command
var helloCmd = &cobra.Command{
	Use:   "hello [name]",
	Short: "Say hello to someone",
	Long: `A simple hello command that demonstrates how to create
subcommands with Cobra.

You can pass a name as an argument, or it will default to "World".`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := "World"
		if len(args) > 0 {
			name = args[0]
		}
		
		uppercase, _ := cmd.Flags().GetBool("uppercase")
		if uppercase {
			name = fmt.Sprintf("%s!", fmt.Sprintf("%s", name))
		}
		
		fmt.Printf("Hello, %s! Welcome to {{.project_name}}.\n", name)
	},
}

func init() {
	rootCmd.AddCommand(helloCmd)
	
	// Add flags for the hello command
	helloCmd.Flags().BoolP("uppercase", "u", false, "Make the greeting uppercase")
}