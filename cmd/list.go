// cmd/list.go
package cmd

import (
	"fmt"
	"os"

	"github.com/GH-Solutions-Consultants/Paxly/core"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all installed dependencies",
	Run: func(cmd *cobra.Command, args []string) {
		// Read config
		data, err := os.ReadFile("paxly.yaml")
		if err != nil {
			logrus.Fatal(errors.Wrap(err, "failed to read paxly.yaml"))
		}

		var config core.Config
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			logrus.Fatal(errors.Wrap(err, "failed to parse paxly.yaml"))
		}

		// Validate configuration
		if err := config.Validate(); err != nil {
			logrus.Fatal(errors.Wrap(err, "configuration validation failed"))
		}

		// Iterate over environments
		for envName, envConfig := range config.Environments {
			fmt.Printf("Environment: %s\n", envName)
			for lang, deps := range envConfig.Dependencies {
				fmt.Printf("  Language: %s\n", lang)
				fmt.Printf("  Dependencies: %v\n", deps)
				plugin, exists := core.GetPluginRegistry().GetPlugin(lang)
				if !exists {
					fmt.Printf("    No plugin found for language '%s'\n", lang)
					continue
				}
				installedDeps, err := plugin.List()
				if err != nil {
					fmt.Printf("    Failed to list dependencies for language '%s': %v\n", lang, err)
					continue
				}
				for _, dep := range installedDeps {
					fmt.Printf("    - %s: %s\n", dep.Name, dep.Version)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
