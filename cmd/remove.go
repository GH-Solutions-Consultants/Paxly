// cmd/remove.go
package cmd

import (
	"os"

	"github.com/GH-Solutions-Consultants/Paxly/core"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	removeLanguage string
	removeName     string
	removeCmd      = &cobra.Command{
		Use:   "remove",
		Short: "Remove a dependency from the project",
		Run: func(cmd *cobra.Command, args []string) {
			// Read config
			data, err := os.ReadFile("pkgmgr.yaml")
			if err != nil {
				logrus.Fatal(errors.Wrap(err, "failed to read pkgmgr.yaml"))
			}

			var config core.Config
			err = yaml.Unmarshal(data, &config)
			if err != nil {
				logrus.Fatal(errors.Wrap(err, "failed to parse pkgmgr.yaml"))
			}

			// Validate configuration
			if err := config.Validate(); err != nil {
				logrus.Fatal(errors.Wrap(err, "configuration validation failed"))
			}

			// Remove dependency
			envConfig, exists := config.Environments["development"]
			if !exists {
				logrus.Fatal(errors.Errorf("environment 'development' does not exist"))
			}

			deps, exists := envConfig.Dependencies[removeLanguage]
			if !exists {
				logrus.Fatal(errors.Errorf("no dependencies found for language '%s'", removeLanguage))
			}

			index := -1
			for i, dep := range deps {
				if dep.Name == removeName {
					index = i
					break
				}
			}

			if index == -1 {
				logrus.Fatal(errors.Errorf("dependency '%s' not found in language '%s'", removeName, removeLanguage))
			}

			// Remove the dependency
			deps = append(deps[:index], deps[index+1:]...)
			config.Environments["development"].Dependencies[removeLanguage] = deps

			// Marshal back to YAML
			updatedData, err := yaml.Marshal(&config)
			if err != nil {
				logrus.Fatal(errors.Wrap(err, "failed to marshal updated configuration"))
			}

			// Write back to pkgmgr.yaml
			err = os.WriteFile("pkgmgr.yaml", updatedData, 0644)
			if err != nil {
				logrus.Fatal(errors.Wrap(err, "failed to write updated pkgmgr.yaml"))
			}

			logrus.Infof("Successfully removed dependency '%s' from language '%s'", removeName, removeLanguage)
		},
	}
)

func init() {
	rootCmd.AddCommand(removeCmd)

	removeCmd.Flags().StringVarP(&removeLanguage, "language", "l", "", "Programming language of the dependency")
	removeCmd.Flags().StringVarP(&removeName, "name", "n", "", "Name of the dependency")
	removeCmd.MarkFlagRequired("language")
	removeCmd.MarkFlagRequired("name")
}
