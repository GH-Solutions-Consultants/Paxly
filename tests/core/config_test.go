// tests/core/config_test.go
package core_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate_Success(t *testing.T) {
	config := core.Config{
		Project: core.ProjectConfig{
			Name:        "TestProject",
			Version:     "1.0.0",
			Description: "A test project",
			Authors: []core.Author{
				{
					Name:  "John Doe",
					Email: "john.doe@example.com",
				},
			},
		},
		Environments: map[string]core.EnvironmentConfig{
			"development": {
				Dependencies: map[string][]core.Dependency{
					"python": {
						{
							Name:    "requests",
							Version: "^2.28",
						},
					},
					"javascript": {
						{
							Name:    "express",
							Version: "^4.17.1",
						},
					},
				},
			},
			"production": {
				Dependencies: map[string][]core.Dependency{
					"python": {
						{
							Name:    "requests",
							Version: "^2.28",
						},
					},
					"javascript": {
						{
							Name:    "express",
							Version: "^4.17.1",
						},
					},
					"go": {
						{
							Name:    "github.com/gin-gonic/gin",
							Version: "^1.7.4",
						},
					},
					"rust": {
						{
							Name:    "serde",
							Version: "^1.0",
						},
					},
				},
			},
		},
		TrustedSources: map[string][]string{
			"python":      {"https://pypi.org/simple"},
			"javascript":  {"https://registry.npmjs.org/"},
			"go":          {"https://proxy.golang.org/"},
			"rust":        {"https://crates.io/"},
		},
	}

	err := config.Validate()
	assert.NoError(t, err, "Expected configuration to be valid")
}

func TestConfig_Validate_MissingProjectName(t *testing.T) {
	config := core.Config{
		Project: core.ProjectConfig{
			// Name is missing
			Version:     "1.0.0",
			Description: "A test project",
			Authors: []core.Author{
				{
					Name:  "John Doe",
					Email: "john.doe@example.com",
				},
			},
		},
		Environments: map[string]core.EnvironmentConfig{},
		TrustedSources: map[string][]string{
			"python": {"https://pypi.org/simple"},
		},
	}

	err := config.Validate()
	assert.Error(t, err, "Expected configuration to fail validation due to missing project name")
}

func TestConfig_Validate_InvalidProjectVersion(t *testing.T) {
	config := core.Config{
		Project: core.ProjectConfig{
			Name:        "TestProject",
			Version:     "invalid_version", // Invalid semantic version
			Description: "A test project",
			Authors: []core.Author{
				{
					Name:  "John Doe",
					Email: "john.doe@example.com",
				},
			},
		},
		Environments: map[string]core.EnvironmentConfig{},
		TrustedSources: map[string][]string{
			"python": {"https://pypi.org/simple"},
		},
	}

	err := config.Validate()
	assert.Error(t, err, "Expected configuration to fail validation due to invalid project version")
}

func TestConfig_Validate_InvalidAuthorEmail(t *testing.T) {
	config := core.Config{
		Project: core.ProjectConfig{
			Name:        "TestProject",
			Version:     "1.0.0",
			Description: "A test project",
			Authors: []core.Author{
				{
					Name:  "John Doe",
					Email: "invalid-email", // Invalid email format
				},
			},
		},
		Environments: map[string]core.EnvironmentConfig{},
		TrustedSources: map[string][]string{
			"python": {"https://pypi.org/simple"},
		},
	}

	err := config.Validate()
	assert.Error(t, err, "Expected configuration to fail validation due to invalid author email")
}

func TestConfig_Validate_InvalidTrustedSourceURL(t *testing.T) {
	config := core.Config{
		Project: core.ProjectConfig{
			Name:        "TestProject",
			Version:     "1.0.0",
			Description: "A test project",
			Authors: []core.Author{
				{
					Name:  "John Doe",
					Email: "john.doe@example.com",
				},
			},
		},
		Environments: map[string]core.EnvironmentConfig{},
		TrustedSources: map[string][]string{
			"python": {"invalid-url"}, // Invalid URL
		},
	}

	err := config.Validate()
	assert.Error(t, err, "Expected configuration to fail validation due to invalid trusted source URL")
}

func TestDependency_Validate_Success(t *testing.T) {
	dep := core.Dependency{
		Name:    "requests",
		Version: "^2.28",
	}

	err := dep.Validate()
	assert.NoError(t, err, "Expected dependency validation to pass")
	assert.NotNil(t, dep.Constraint, "Expected Constraint to be parsed")
}

func TestDependency_Validate_InvalidVersion(t *testing.T) {
	dep := core.Dependency{
		Name:    "requests",
		Version: "invalid_version",
	}

	err := dep.Validate()
	assert.Error(t, err, "Expected dependency validation to fail due to invalid version")
}

func TestDependency_Validate_MissingName(t *testing.T) {
	dep := core.Dependency{
		// Name is missing
		Version: "^2.28",
	}

	err := dep.Validate()
	assert.Error(t, err, "Expected dependency validation to fail due to missing name")
}

func TestDependency_Validate_MissingVersion(t *testing.T) {
	dep := core.Dependency{
		Name:    "requests",
		Version: "", // Missing version
	}

	err := dep.Validate()
	assert.Error(t, err, "Expected dependency validation to fail due to missing version")
}
