// core/config.go
package core

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Config represents the entire project configuration.
type Config struct {
	Project        ProjectConfig                `yaml:"project" validate:"required,dive"`
	Environments   map[string]EnvironmentConfig `yaml:"environments" validate:"required,dive"`
	TrustedSources map[string][]string          `yaml:"trusted_sources" validate:"required,dive,dive,uri"`
}

// EnvironmentConfig holds dependencies for a specific environment.
type EnvironmentConfig struct {
	Dependencies map[string][]Dependency `yaml:"dependencies" validate:"required,dive,dive"`
}

// ProjectConfig holds project metadata.
type ProjectConfig struct {
	Name        string   `yaml:"name" validate:"required"`
	Version     string   `yaml:"version" validate:"required,semver"`
	Description string   `yaml:"description"`
	Authors     []Author `yaml:"authors" validate:"dive"`
}

// Author represents a project author.
type Author struct {
	Name  string `yaml:"name" validate:"required"`
	Email string `yaml:"email" validate:"required,email"`
}

// Validate parses and validates the entire configuration.
func (c *Config) Validate() error {
	validate := validator.New()
	// Register custom validation
	validate.RegisterValidation("semver", validateSemVer)

	// Validate Project
	if err := validate.Struct(c.Project); err != nil {
		return err
	}

	// Validate Environments
	for envName, envConfig := range c.Environments {
		if err := validate.Struct(envConfig); err != nil {
			return fmt.Errorf("validation failed for environment '%s': %v", envName, err)
		}
		for lang, deps := range envConfig.Dependencies {
			for _, dep := range deps {
				if err := dep.Validate(); err != nil {
					return fmt.Errorf("invalid dependency '%s' in environment '%s', language '%s': %v", dep.Name, envName, lang, err)
				}
			}
		}
	}

	return nil
}
