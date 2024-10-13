// plugins/rust/rust_plugin.go
package rust

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/GH-Solutions-Consultants/Paxly/core"
	"github.com/sirupsen/logrus"
)

// Ensure the RustPlugin implements the PackageManagerPlugin interface.
var _ core.PackageManagerPlugin = (*RustPlugin)(nil)

// RustPlugin is the plugin for managing Rust dependencies.
type RustPlugin struct {
	executor core.Executor
}

// Register the RustPlugin with the PluginRegistry during initialization.
func init() {
	plugin := NewRustPlugin(nil)
	err := core.GetPluginRegistry().RegisterPlugin(plugin.Language(), plugin)
	if err != nil {
		logrus.Fatalf("Failed to register Rust plugin: %v", err)
	}
}

// NewRustPlugin creates a new instance of RustPlugin with the given executor.
func NewRustPlugin(executor core.Executor) *RustPlugin {
	if executor == nil {
		executor = &core.RealExecutor{}
	}
	return &RustPlugin{
		executor: executor,
	}
}

// APIVersion returns the plugin API version.
func (p *RustPlugin) APIVersion() string {
	return core.PluginAPIVersion
}

// Language returns the name of the language this plugin manages.
func (p *RustPlugin) Language() string {
	return "rust"
}

// Initialize sets up the Rust plugin with necessary configurations.
func (p *RustPlugin) Initialize(config core.Config) error {
	logrus.Info("Initializing Rust plugin...")
	// Validate cargo installation
	err := p.executor.Run(&core.Command{Name: "cargo", Args: []string{"--version"}})
	if err != nil {
		return fmt.Errorf("cargo is not installed or not in PATH: %v", err)
	}
	// Ensure cargo-edit is installed
	cmd := &core.Command{
		Name: "cargo",
		Args: []string{"install", "cargo-edit"},
	}
	if err := p.executor.Run(cmd); err != nil {
		return fmt.Errorf("failed to install cargo-edit: %v", err)
	}
	// Ensure Cargo.toml exists; if not, initialize it
	_, err = p.executor.Output(&core.Command{
		Name: "cargo",
		Args: []string{"metadata", "--format-version", "1"},
	})
	if err != nil {
		logrus.Info("Cargo.toml not found. Initializing Cargo project...")
		initCmd := &core.Command{
			Name: "cargo",
			Args: []string{"init"},
		}
		if err := p.executor.Run(initCmd); err != nil {
			return fmt.Errorf("failed to initialize Cargo project: %v", err)
		}
	}
	return nil
}

// Install installs the specified Rust dependencies.
func (p *RustPlugin) Install(deps []core.Dependency) error {
	for _, dep := range deps {
		logrus.WithFields(logrus.Fields{
			"dependency": dep.Name,
			"version":    dep.Version,
		}).Info("Installing Rust package")

		// Add dependency to Cargo.toml
		if err := p.addDependencyToCargoToml(dep); err != nil {
			logrus.Errorf("Failed to add Rust package '%s' to Cargo.toml: %v", dep.Name, err)
			return err
		}

		// Run cargo build to fetch dependencies
		cmd := &core.Command{
			Name: "cargo",
			Args: []string{"build"},
		}
		if err := p.executor.Run(cmd); err != nil {
			logrus.Errorf("Failed to build Rust project after adding package '%s': %v", dep.Name, err)
			return err
		}
		logrus.Infof("Successfully installed Rust package: %s", dep.Name)
	}

	// Run security scans after installation.
	if err := p.RunSecurityScan(); err != nil {
		logrus.Warnf("Security scan encountered issues: %v", err)
	}

	return nil
}

// Update updates the specified Rust dependencies.
func (p *RustPlugin) Update(deps []core.Dependency) error {
	for _, dep := range deps {
		logrus.WithFields(logrus.Fields{
			"dependency": dep.Name,
			"version":    dep.Version,
		}).Info("Updating Rust package")

		// Update dependency in Cargo.toml
		if err := p.updateDependencyInCargoToml(dep); err != nil {
			logrus.Errorf("Failed to update Rust package '%s' in Cargo.toml: %v", dep.Name, err)
			return err
		}

		// Run cargo update
		cmd := &core.Command{
			Name: "cargo",
			Args: []string{"update", "-p", dep.Name},
		}
		if err := p.executor.Run(cmd); err != nil {
			logrus.Errorf("Failed to update Rust package '%s': %v", dep.Name, err)
			return err
		}
		logrus.Infof("Successfully updated Rust package: %s", dep.Name)
	}

	// Run security scans after update.
	if err := p.RunSecurityScan(); err != nil {
		logrus.Warnf("Security scan encountered issues: %v", err)
	}

	return nil
}

// Remove removes the specified Rust dependency.
func (p *RustPlugin) Remove(dep core.Dependency) error {
	logrus.WithFields(logrus.Fields{
		"dependency": dep.Name,
	}).Info("Removing Rust package")

	// Remove dependency from Cargo.toml
	if err := p.removeDependencyFromCargoToml(dep); err != nil {
		logrus.Errorf("Failed to remove Rust package '%s' from Cargo.toml: %v", dep.Name, err)
		return err
	}

	// Run cargo build to update dependencies
	cmd := &core.Command{
		Name: "cargo",
		Args: []string{"build"},
	}
	if err := p.executor.Run(cmd); err != nil {
		logrus.Errorf("Failed to build Rust project after removing package '%s': %v", dep.Name, err)
		return err
	}
	logrus.Infof("Successfully removed Rust package: %s", dep.Name)

	return nil
}

// List lists all installed Rust dependencies.
func (p *RustPlugin) List() ([]core.Dependency, error) {
	cmd := &core.Command{
		Name: "cargo",
		Args: []string{"tree", "--no-dev"},
	}
	output, err := p.executor.Output(cmd)
	if err != nil {
		return nil, err
	}

	// Parse the cargo tree output to list dependencies
	deps := []core.Dependency{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "├─") || strings.HasPrefix(line, "└─") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				pkgInfo := parts[1]
				pkgParts := strings.Split(pkgInfo, " ")
				if len(pkgParts) == 2 {
					deps = append(deps, core.Dependency{
						Name:    pkgParts[0],
						Version: "=" + pkgParts[1],
					})
				}
			}
		}
	}

	return deps, nil
}

// ListVersions lists all available versions for a given Rust package.
func (p *RustPlugin) ListVersions(depName string) ([]string, error) {
	cmd := &core.Command{
		Name: "cargo",
		Args: []string{"search", depName, "--limit", "100"},
	}
	output, err := p.executor.Output(cmd)
	if err != nil {
		return nil, err
	}

	// Parse cargo search output to get versions
	// Cargo search format: name = "description" [version]
	versions := []string{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, depName) {
			// Extract version using simple string manipulation
			parts := strings.Split(line, " ")
			for _, part := range parts {
				if strings.HasPrefix(part, "v") {
					versions = append(versions, part)
				}
			}
		}
	}

	return versions, nil
}

// GetTransitiveDependencies fetches transitive dependencies for a given dependency.
// Rust handles transitive dependencies automatically, so this can be a no-op.
func (p *RustPlugin) GetTransitiveDependencies(depName, version string) ([]core.Dependency, error) {
	// Rust handles transitive dependencies automatically, so return nil
	return nil, nil
}

// RunSecurityScan runs security scans on installed Rust packages using 'cargo audit'.
func (p *RustPlugin) RunSecurityScan() error {
	logrus.Info("Running security scan on Rust dependencies using 'cargo audit'...")
	// Ensure cargo-audit is installed
	cmd := &core.Command{
		Name: "cargo",
		Args: []string{"install", "cargo-audit"},
	}
	if err := p.executor.Run(cmd); err != nil {
		return fmt.Errorf("failed to install cargo-audit: %v", err)
	}

	// Run cargo audit
	scanCmd := &core.Command{
		Name: "cargo",
		Args: []string{"audit", "--json"},
	}
	scanOutput, err := p.executor.Output(scanCmd)
	if err != nil {
		return fmt.Errorf("security scan failed: %v", err)
	}

	var auditResults []core.SecurityVulnerability
	if err := json.Unmarshal(scanOutput, &auditResults); err != nil {
		return fmt.Errorf("failed to parse cargo audit output: %v", err)
	}

	if len(auditResults) > 0 {
		logrus.Warn("Vulnerabilities detected in Rust dependencies:")
		for _, vuln := range auditResults {
			logrus.Warnf("- %s: %s", vuln.PackageName, vuln.VulnerabilityID)
		}
	} else {
		logrus.Info("No vulnerabilities found in Rust dependencies.")
	}

	return nil
}

// GetVulnerabilities retrieves security vulnerabilities.
func (p *RustPlugin) GetVulnerabilities() ([]core.SecurityVulnerability, error) {
	cmd := &core.Command{
		Name: "cargo",
		Args: []string{"audit", "--json"},
	}
	scanOutput, err := p.executor.Output(cmd)
	if err != nil {
		return nil, fmt.Errorf("security scan failed: %v", err)
	}

	var auditResults []core.SecurityVulnerability
	if err := json.Unmarshal(scanOutput, &auditResults); err != nil {
		return nil, fmt.Errorf("failed to parse cargo audit output: %v", err)
	}

	return auditResults, nil
}

// Cleanup cleans up Rust plugin resources.
func (p *RustPlugin) Cleanup() error {
	logrus.Info("Cleaning up Rust plugin resources...")
	// Implement any necessary cleanup
	return nil
}

// addDependencyToCargoToml adds a dependency to Cargo.toml
func (p *RustPlugin) addDependencyToCargoToml(dep core.Dependency) error {
	// Add dependency using cargo-edit
	cmd := &core.Command{
		Name: "cargo",
		Args: []string{"add", fmt.Sprintf("%s=%s", dep.Name, strings.TrimPrefix(dep.Version, "^"))},
	}
	return p.executor.Run(cmd)
}

// updateDependencyInCargoToml updates a dependency in Cargo.toml
func (p *RustPlugin) updateDependencyInCargoToml(dep core.Dependency) error {
	// Update dependency using cargo-edit
	cmd := &core.Command{
		Name: "cargo",
		Args: []string{"upgrade", dep.Name},
	}
	return p.executor.Run(cmd)
}

// removeDependencyFromCargoToml removes a dependency from Cargo.toml
func (p *RustPlugin) removeDependencyFromCargoToml(dep core.Dependency) error {
	// Remove dependency using cargo-edit
	cmd := &core.Command{
		Name: "cargo",
		Args: []string{"remove", dep.Name},
	}
	return p.executor.Run(cmd)
}
