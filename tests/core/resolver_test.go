// tests/core/resolver_test.go
package core_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestResolver_ResolveDependencies_NoConflicts(t *testing.T) {
	config := core.Config{
		Project: core.ProjectConfig{
			Name:    "TestProject",
			Version: "1.0.0",
			Authors: []core.Author{
				{
					Name:  "Tester",
					Email: "tester@example.com",
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
		},
		TrustedSources: map[string][]string{
			"python":      {"https://pypi.org/simple"},
			"javascript":  {"https://registry.npmjs.org/"},
			"go":          {"https://proxy.golang.org/"},
			"rust":        {"https://crates.io/"},
		},
	}

	// Validate config
	validate := validator.New()
	err := validate.Struct(config)
	assert.NoError(t, err)

	// Initialize Plugin Registry with mock plugins
	pr := core.NewPluginRegistry()
	pr.RegisterPlugin(&MockPythonPlugin{})
	pr.RegisterPlugin(&MockJavaScriptPlugin{})

	// Initialize Resolver
	resolver := core.NewResolver(config, pr)
	resolvedVersions, err := resolver.ResolveDependencies("development")
	assert.NoError(t, err, "ResolveDependencies should not return an error")

	// Verify resolved dependencies
	expected := map[string]string{
		"requests": "2.28.1",
		"express":  "4.17.1",
	}
	for name, version := range expected {
		resolvedVersion, exists := resolvedVersions[name]
		assert.True(t, exists, "Dependency '%s' should be resolved", name)
		assert.Equal(t, version, resolvedVersion, "Version mismatch for '%s'", name)
	}
}

// Mock Plugins for Testing
type MockPythonPlugin struct{}

func (p *MockPythonPlugin) APIVersion() string        { return core.PluginAPIVersion }
func (p *MockPythonPlugin) Language() string          { return "python" }
func (p *MockPythonPlugin) Initialize(config core.Config) error { return nil }
func (p *MockPythonPlugin) Install(deps []core.Dependency) error { return nil }
func (p *MockPythonPlugin) Update(deps []core.Dependency) error { return nil }
func (p *MockPythonPlugin) Remove(dep core.Dependency) error { return nil }
func (p *MockPythonPlugin) List() ([]core.Dependency, error) { return nil, nil }
func (p *MockPythonPlugin) ListVersions(depName string) ([]string, error) {
	return []string{"2.28.0", "2.28.1"}, nil
}
func (p *MockPythonPlugin) GetTransitiveDependencies(depName, version string) ([]core.Dependency, error) {
	return nil, nil
}
func (p *MockPythonPlugin) Cleanup() error { return nil }

type MockJavaScriptPlugin struct{}

func (p *MockJavaScriptPlugin) APIVersion() string        { return core.PluginAPIVersion }
func (p *MockJavaScriptPlugin) Language() string          { return "javascript" }
func (p *MockJavaScriptPlugin) Initialize(config core.Config) error { return nil }
func (p *MockJavaScriptPlugin) Install(deps []core.Dependency) error { return nil }
func (p *MockJavaScriptPlugin) Update(deps []core.Dependency) error { return nil }
func (p *MockJavaScriptPlugin) Remove(dep core.Dependency) error { return nil }
func (p *MockJavaScriptPlugin) List() ([]core.Dependency, error) { return nil, nil }
func (p *MockJavaScriptPlugin) ListVersions(depName string) ([]string, error) {
	return []string{"4.17.0", "4.17.1"}, nil
}
func (p *MockJavaScriptPlugin) GetTransitiveDependencies(depName, version string) ([]core.Dependency, error) {
	return nil, nil
}
func (p *MockJavaScriptPlugin) Cleanup() error { return nil }
