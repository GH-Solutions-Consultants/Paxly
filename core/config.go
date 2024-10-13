// core/config.go

package core

import (
    "fmt"
    "github.com/go-playground/validator/v10"
    "github.com/sirupsen/logrus"
)

// Config represents the overall configuration.
type Config struct {
    Project        ProjectConfig                `yaml:"project" validate:"required,dive"`
    Environments   map[string]EnvironmentConfig `yaml:"environments" validate:"required,dive"`
    TrustedSources map[string][]string          `yaml:"trusted_sources" validate:"required,dive,dive,uri"`
}

// EnvironmentConfig represents the configuration for a specific environment.
type EnvironmentConfig struct {
    Dependencies map[string][]Dependency `yaml:"dependencies" validate:"required,dive,dive"`
}

// ProjectConfig represents the project-specific configuration.
type ProjectConfig struct {
    Name        string   `yaml:"name" validate:"required"`
    Version     string   `yaml:"version" validate:"required,semver"`
    Description string   `yaml:"description"`
    Authors     []Author `yaml:"authors" validate:"dive"`
}

// Author represents an author of the project.
type Author struct {
    Name  string `yaml:"name" validate:"required"`
    Email string `yaml:"email" validate:"required,email"`
}

// Validate validates the entire configuration.
func (c *Config) Validate() error {
    validate := validator.New()
    validate.RegisterValidation("semver", validateSemVer)

    // Validate Project Config
    if err := validate.Struct(c.Project); err != nil {
        return err
    }

    // Validate Environments
    for envName, envConfig := range c.Environments {
        if err := validate.Struct(envConfig); err != nil {
            return fmt.Errorf("validation failed for environment '%s': %v", envName, err)
        }

        // Validate Dependencies
        for lang, deps := range envConfig.Dependencies {
            for i := range deps { // Use index to get pointer
                dep := &deps[i] // Get pointer to the actual slice element
                if err := dep.Validate(); err != nil {
                    return fmt.Errorf("invalid dependency '%s' in environment '%s', language '%s': %v", dep.Name, envName, lang, err)
                }
                // Debug log to confirm Constraint is set
                logrus.Debugf("Dependency '%s' in language '%s' has constraint '%s'", dep.Name, lang, dep.Constraint.String())
            }
        }
    }

    return nil
}
