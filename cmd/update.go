// cmd/update.go
package cmd

import (
	"os"

	"github.com/GH-Solutions-Consultants/Paxly/core"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	updateLanguage string
	updateName     string
	updateVersion  string
	updateCmd      = &cobra.Command{
		Use:   "update",
		Short: "Update a dependency in the project",
		Run: func(cmd *cobra.Command, args []string) {
			// Read config
			data, err := os.ReadFile("paxly.yaml") // Changed from "pkgmgr.yaml" to "paxly.yaml"
			if err != nil {
				core.LogFatal(errors.Wrap(err, "failed to read paxly.yaml"))
			}

			var config core.Config
			err = yaml.Unmarshal(data, &config)
			if err != nil {
				core.LogFatal(errors.Wrap(err, "failed to parse paxly.yaml"))
			}

			// Validate configuration
			if err := config.Validate(); err != nil {
				core.LogFatal(errors.Wrap(err, "configuration validation failed"))
			}

			// Update dependency in the 'development' environment
			envConfig, exists := config.Environments["development"]
			if !exists {
				core.LogFatal(errors.Errorf("environment 'development' does not exist"))
			}

			deps, exists := envConfig.Dependencies[updateLanguage]
			if !exists {
				logrus.Fatal(errors.Errorf("no dependencies found for language '%s'", updateLanguage))
			}

			found := false
			for i, dep := range deps {
				if dep.Name == updateName {
					deps[i].Version = updateVersion
					if err := deps[i].Validate(); err != nil {
						core.LogFatal(errors.Wrap(err, "invalid version constraint"))
					}
					found = true
					break
				}
			}

			if !found {
				core.LogFatal(errors.Errorf("dependency '%s' not found in language '%s'", updateName, updateLanguage))
			}

			// Update the dependencies in the environment config
			config.Environments["development"].Dependencies[updateLanguage] = deps

			// Marshal back to YAML
			updatedData, err := yaml.Marshal(&config)
			if err != nil {
				core.LogFatal(errors.Wrap(err, "failed to marshal updated configuration"))
			}

			// Write back to paxly.yaml
			err = os.WriteFile("paxly.yaml", updatedData, 0644) // Changed from "pkgmgr.yaml" to "paxly.yaml"
			if err != nil {
				core.LogFatal(errors.Wrap(err, "failed to write updated paxly.yaml"))
			}

			// Log the successful update
			logrus.Infof("Successfully updated dependency '%s' to version '%s' in language '%s'", updateName, updateVersion, updateLanguage)
		},
	}
)

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringVarP(&updateLanguage, "language", "l", "", "Programming language of the dependency")
	updateCmd.Flags().StringVarP(&updateName, "name", "n", "", "Name of the dependency")
	updateCmd.Flags().StringVarP(&updateVersion, "version", "r", "", "New version constraint of the dependency") // Changed shorthand from 'v' to 'r'
	updateCmd.MarkFlagRequired("language")
	updateCmd.MarkFlagRequired("name")
	updateCmd.MarkFlagRequired("version")
}