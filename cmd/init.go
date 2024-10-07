// cmd/init.go
package cmd

import (
	"github.com/GH-Solutions-Consultants/Paxly/core"
	"github.com/spf13/cobra"
)

var (
	projectName        string
	projectVersion     string
	projectDescription string
	projectAuthors     []string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new paxly project",
	Run: func(cmd *cobra.Command, args []string) {
		err := core.InitializeProject(projectName, projectVersion, projectDescription, projectAuthors)
		if err != nil {
			cmd.PrintErrf("Failed to initialize project: %v\n", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&projectName, "name", "n", "", "Name of the project")
	// Change the shorthand for the version flag from 'v' to something else, e.g., 'r'
	initCmd.Flags().StringVarP(&projectVersion, "version", "r", "1.0.0", "Version of the project")
	initCmd.Flags().StringVarP(&projectDescription, "description", "s", "", "Description of the project")
	initCmd.Flags().StringSliceVarP(&projectAuthors, "authors", "a", []string{}, "Authors of the project (format: 'Name <email>')")

	initCmd.MarkFlagRequired("name")
}
