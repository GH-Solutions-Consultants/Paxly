// cmd/add.go
package cmd

import (
	"os"

	"github.com/GH-Solutions-Consultants/Paxly/core"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	addLanguage string
	addName     string
	addVersion  string
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new dependency to the project",
	Run: func(cmd *cobra.Command, args []string) {
		// Read existing config
		data, err := os.ReadFile("paxly.yaml")
		if err != nil {
			logrus.Fatalf("Failed to read paxly.yaml: %v", err)
		}

		var config core.Config
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			logrus.Fatalf("Failed to parse paxly.yaml: %v", err)
		}

		// Validate inputs
		if addLanguage == "" || addName == "" || addVersion == "" {
			logrus.Fatal("Language, name, and version must be specified")
		}

		// Add dependency to the specified environment (defaulting to development)
		envConfig, exists := config.Environments["development"]
		if !exists {
			logrus.Fatal("Development environment not found in configuration")
		}

		dep := core.Dependency{
			Name:    addName,
			Version: addVersion,
		}

		envConfig.Dependencies[addLanguage] = append(envConfig.Dependencies[addLanguage], dep)
		config.Environments["development"] = envConfig

		// Marshal back to YAML
		newData, err := yaml.Marshal(&config)
		if err != nil {
			logrus.Fatalf("Failed to marshal updated configuration: %v", err)
		}

		// Write back to paxly.yaml
		err = os.WriteFile("paxly.yaml", newData, 0644)
		if err != nil {
			logrus.Fatalf("Failed to write updated paxly.yaml: %v", err)
		}

		logrus.Infof("Added dependency '%s' version '%s' to language '%s' in development environment.", addName, addVersion, addLanguage)
	},
}

func init() {
	addCmd.Flags().StringVarP(&addLanguage, "language", "l", "", "Programming language of the dependency")
	addCmd.Flags().StringVarP(&addName, "name", "n", "", "Name of the dependency")
	addCmd.Flags().StringVarP(&addVersion, "version", "v", "", "Version constraint of the dependency")
	addCmd.MarkFlagRequired("language")
	addCmd.MarkFlagRequired("name")
	addCmd.MarkFlagRequired("version")

	rootCmd.AddCommand(addCmd)
}
