package cmd

import (
	"fmt"

	"github.com/dsaiztc/dev/internal/shell"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Print shell wrapper function",
	Long:  `Prints a shell function to stdout. Add eval "$(dev init)" to your ~/.zshrc or ~/.bashrc.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(shell.WrapperFunc())
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
