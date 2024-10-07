// plugins/go/go_plugin.go
package go_plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/GH-Solutions-Consultants/Paxly/core"
	"github.com/sirupsen/logrus"
)

// Ensure the GoPlugin implements the PackageManagerPlugin interface.
var _ core.PackageManagerPlugin = (*GoPlugin)(nil)

// GoPlugin is the plugin for managing Go dependencies.
type GoPlugin struct {
	executor core.Executor
}

// NewGoPlugin creates a new instance of GoPlugin with the given executor.
func NewGoPlugin(executor core.Executor) *GoPlugin {
	if executor == nil {
		executor = &core.RealExecutor{}
	}
	return &GoPlugin{
		executor: executor,
	}
}

// APIVersion returns the plugin API version.
func (p *GoPlugin) APIVersion() string {
	return core.PluginAPIVersion
}

// Language returns the name of the language this plugin manages.
func (p *GoPlugin) Language() string {
	return "go"
}

// Initialize sets up the Go plugin with necessary configurations.
func (p *GoPlugin) Initialize(config core.Config) error {
	logrus.Info("Initializing Go plugin...")
	// Validate Go installation
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go is not installed or not in PATH")
	}
	// Ensure go.mod exists; if not, initialize it
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		logrus.Info("go.mod not found. Initializing Go module...")
		cmd := exec.Command("go", "mod", "init", config.Project.Name)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to initialize Go module: %v", err)
		}
	}
	return nil
}

// Install installs the specified Go dependencies.
func (p *GoPlugin) Install(deps []core.Dependency) error {
	for _, dep := range deps {
		logrus.WithFields(logrus.Fields{
			"dependency": dep.Name,
			"version":    dep.Version,
		}).Info("Installing Go package")

		cmd := core.Command{
			Name: "go",
			Args: []string{"get", dep.Name},
		}
		err := p.executor.Run(&cmd)
		if err != nil {
			logrus.Errorf("Failed to install Go package '%s': %v", dep.Name, err)
			return err
		}
		logrus.Infof("Successfully installed Go package: %s", dep.Name)
	}

	// Run go mod tidy to clean up dependencies
	if err := p.runGoModTidy(); err != nil {
		logrus.Errorf("Failed to run 'go mod tidy': %v", err)
		return err
	}

	// Optionally, run security scans here if integrated

	return nil
}

// Update updates the specified Go dependencies.
func (p *GoPlugin) Update(deps []core.Dependency) error {
	for _, dep := range deps {
		pkgStr := fmt.Sprintf("%s@%s", dep.Name, dep.Version)
		logrus.WithFields(logrus.Fields{
			"dependency": dep.Name,
			"version":    dep.Version,
		}).Info("Updating Go package")

		cmd := core.Command{
			Name: "go",
			Args: []string{"get", "-u", pkgStr},
		}
		err := p.executor.Run(&cmd)
		if err != nil {
			logrus.Errorf("Failed to update Go package '%s': %v", dep.Name, err)
			return err
		}
		logrus.Infof("Successfully updated Go package: %s", dep.Name)
	}

	// Run go mod tidy to clean up dependencies
	if err := p.runGoModTidy(); err != nil {
		logrus.Errorf("Failed to run 'go mod tidy': %v", err)
		return err
	}

	return nil
}

// Remove removes the specified Go dependency.
func (p *GoPlugin) Remove(dep core.Dependency) error {
	pkgStr := dep.Name
	logrus.WithFields(logrus.Fields{
		"dependency": dep.Name,
	}).Info("Removing Go package")

	// To remove a dependency, we need to edit go.mod to remove the require directive
	// Since Go doesn't have a direct command to remove a module, we'll use 'go mod tidy' after manual removal

	// Remove the dependency from go.mod by running 'go mod edit -droprequire=dep.Name'
	cmd := core.Command{
		Name: "go",
		Args: []string{"mod", "edit", "-droprequire", dep.Name},
	}
	err := p.executor.Run(&cmd)
	if err != nil {
		logrus.Errorf("Failed to remove Go package '%s' from go.mod: %v", dep.Name, err)
		return err
	}

	// Run go mod tidy to clean up
	if err := p.runGoModTidy(); err != nil {
		logrus.Errorf("Failed to run 'go mod tidy' after removing package '%s': %v", dep.Name, err)
		return err
	}

	logrus.Infof("Successfully removed Go package: %s", dep.Name)
	return nil
}

// List lists all installed Go dependencies.
func (p *GoPlugin) List() ([]core.Dependency, error) {
	cmd := core.Command{
		Name: "go",
		Args: []string{"list", "-m", "all"},
	}
	output, err := p.executor.Output(&cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list Go modules: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	deps := []core.Dependency{}
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		deps = append(deps, core.Dependency{
			Name:    parts[0],
			Version: "=" + parts[1],
		})
	}

	return deps, nil
}

// ListVersions lists all available versions for a given Go dependency.
func (p *GoPlugin) ListVersions(depName string) ([]string, error) {
	cmd := core.Command{
		Name: "go",
		Args: []string{"list", "-m", "-versions", depName},
	}
	output, err := p.executor.Output(&cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions for '%s': %v", depName, err)
	}

	parts := strings.Fields(string(output))
	if len(parts) < 2 {
		return nil, fmt.Errorf("no versions found for '%s'", depName)
	}

	versions := strings.Split(parts[1], " ")
	return versions, nil
}

// GetTransitiveDependencies fetches transitive dependencies for a given dependency.
func (p *GoPlugin) GetTransitiveDependencies(depName, version string) ([]core.Dependency, error) {
	// Use 'go list -m -json all' to parse dependencies
	cmd := core.Command{
		Name: "go",
		Args: []string{"list", "-m", "-json", "all"},
	}
	output, err := p.executor.Output(&cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list modules for '%s': %v", depName, err)
	}

	var modules []map[string]interface{}
	dec := json.NewDecoder(bytes.NewReader(output))
	for {
		var m map[string]interface{}
		if err := dec.Decode(&m); err != nil {
			break
		}
		modules = append(modules, m)
	}

	transDeps := []core.Dependency{}
	for _, mod := range modules {
		if mod["Path"] == depName && mod["Version"] == version {
			if deps, ok := mod["Dependencies"].([]interface{}); ok {
				for _, d := range deps {
					if depMap, ok := d.(string); ok {
						parts := strings.Split(depMap, " ")
						if len(parts) >= 2 {
							transDeps = append(transDeps, core.Dependency{
								Name:    parts[0],
								Version: "=" + parts[1],
							})
						} else {
							transDeps = append(transDeps, core.Dependency{
								Name:    parts[0],
								Version: "",
							})
						}
					}
				}
			}
			break
		}
	}

	return transDeps, nil
}

// GetVulnerabilities retrieves security vulnerabilities using 'go list -m -json all' and an external tool.
func (p *GoPlugin) GetVulnerabilities() ([]core.SecurityVulnerability, error) {
	// Placeholder: Implement integration with a security tool like 'go-critic' or 'govulncheck'
	// For demonstration, we'll return an empty slice
	return []core.SecurityVulnerability{}, nil
}

// Cleanup performs any necessary cleanup operations.
func (p *GoPlugin) Cleanup() error {
	logrus.Info("Cleaning up Go plugin resources...")
	// Implement any necessary cleanup
	return nil
}

// runGoModTidy runs 'go mod tidy' to clean up dependencies.
func (p *GoPlugin) runGoModTidy() error {
	cmd := core.Command{
		Name: "go",
		Args: []string{"mod", "tidy"},
	}
	err := p.executor.Run(&cmd)
	if err != nil {
		return fmt.Errorf("failed to run 'go mod tidy': %v", err)
	}
	logrus.Info("Successfully ran 'go mod tidy'")
	return nil
}
