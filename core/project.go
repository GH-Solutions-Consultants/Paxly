// core/project.go
package core

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// InitializeProject initializes a new pkgmgr project by creating a pkgmgr.yaml file.
func InitializeProject(name, version, description string, authors []string) error {
	if name == "" {
		return fmt.Errorf("project name is required")
	}

	authorStructs := []Author{}
	for _, authorStr := range authors {
		author, err := parseAuthor(authorStr)
		if err != nil {
			return errors.Wrapf(err, "invalid author format '%s'", authorStr)
		}
		authorStructs = append(authorStructs, author)
	}

	config := Config{
		Project: ProjectConfig{
			Name:        name,
			Version:     version,
			Description: description,
			Authors:     authorStructs,
		},
		Environments: map[string]EnvironmentConfig{
			"development": {
				Dependencies: map[string][]Dependency{},
			},
			"production": {
				Dependencies: map[string][]Dependency{},
			},
		},
		TrustedSources: map[string][]string{
			"python":      {"https://pypi.org/simple"},
			"javascript":  {"https://registry.npmjs.org/"},
			"go":          {"https://proxy.golang.org/"},
			"rust":        {"https://crates.io/"},
		},
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return errors.Wrap(err, "failed to marshal configuration")
	}

	if _, err := os.Stat("pkgmgr.yaml"); err == nil {
		return fmt.Errorf("pkgmgr.yaml already exists")
	}

	err = os.WriteFile("pkgmgr.yaml", data, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write pkgmgr.yaml")
	}

	logrus.Info("Initialized pkgmgr project with pkgmgr.yaml")

	// Validate the config
	if err := config.Validate(); err != nil {
		return errors.Wrap(err, "configuration validation failed")
	}

	logrus.Info("Configuration validated successfully.")

	return nil
}

// parseAuthor parses a string in the format "Name <email>"
func parseAuthor(authorStr string) (Author, error) {
	re := regexp.MustCompile(`^(.*?)\s*<(.+)>$`)
	matches := re.FindStringSubmatch(authorStr)
	if len(matches) != 3 {
		return Author{}, fmt.Errorf("invalid author format")
	}
	return Author{
		Name:  matches[1],
		Email: matches[2],
	}, nil
}
