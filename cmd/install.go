// cmd/install.go
package cmd

import (
	"os"

	"github.com/GH-Solutions-Consultants/Paxly/core"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var env string

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install all dependencies for the project",
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

		// Initialize Resolver
		resolver := core.NewResolver(config, core.GetPluginRegistry())
		if err := resolver.ResolveDependencies(env); err != nil {
			logrus.Fatal(errors.Wrap(err, "dependency resolution failed"))
		}

		// Iterate over resolved dependencies and install via plugins
		envConfig, exists := config.Environments[env]
		if !exists {
			logrus.Fatal(errors.Errorf("specified environment '%s' does not exist", env))
		}

		for lang, deps := range envConfig.Dependencies {
			plugin, exists := core.GetPluginRegistry().GetPlugin(lang)
			if !exists {
				logrus.Warnf("No plugin found for language '%s'", lang)
				continue
			}
			if err := plugin.Install(deps); err != nil {
				logrus.Errorf("Failed to install dependencies for language '%s': %v", lang, err)
				logrus.Info("Ensure that the necessary package manager is installed and configured correctly.")
			}
		}

		logrus.Info("All dependencies installed successfully.")
	},
}

func init() {
	installCmd.Flags().StringVarP(&env, "env", "e", "development", "Specify the environment to use")
	rootCmd.AddCommand(installCmd)
}
