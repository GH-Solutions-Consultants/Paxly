// plugins/python/python_plugin.go
package python

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GH-Solutions-Consultants/Paxly/core"
	"github.com/sirupsen/logrus"
)

// Ensure the PythonPlugin implements the PackageManagerPlugin interface.
var _ core.PackageManagerPlugin = (*PythonPlugin)(nil)

// PythonPlugin is the plugin for managing Python dependencies.
type PythonPlugin struct {
	executor core.Executor
}

// NewPythonPlugin creates a new instance of PythonPlugin with the given executor.
func NewPythonPlugin(executor core.Executor) *PythonPlugin {
	if executor == nil {
		executor = &core.RealExecutor{}
	}
	return &PythonPlugin{
		executor: executor,
	}
}

// APIVersion returns the plugin API version.
func (p *PythonPlugin) APIVersion() string {
	return core.PluginAPIVersion
}

// Language returns the name of the language this plugin manages.
func (p *PythonPlugin) Language() string {
	return "python"
}

// Initialize sets up the Python plugin with necessary configurations.
func (p *PythonPlugin) Initialize(config core.Config) error {
	logrus.Info("Initializing Python plugin...")
	// Validate Python installation
	if err := p.executor.Run(&core.Command{Name: "python3", Args: []string{"--version"}}); err != nil {
		return fmt.Errorf("python3 is not installed or not in PATH")
	}
	if err := p.executor.Run(&core.Command{Name: p.getPipPath(), Args: []string{"--version"}}); err != nil {
		return fmt.Errorf("pip is not installed or not in PATH")
	}
	// Ensure pipdeptree is installed
	if err := p.ensurePipDeptree(); err != nil {
		return err
	}
	return nil
}

// ensurePipDeptree ensures that pipdeptree is installed in the virtual environment.
func (p *PythonPlugin) ensurePipDeptree() error {
	cmd := &core.Command{
		Name: p.getPipPath(),
		Args: []string{"install", "pipdeptree"},
	}
	if err := p.executor.Run(cmd); err != nil {
		return fmt.Errorf("failed to install pipdeptree: %v", err)
	}
	return nil
}

// getPipPath returns the path to the pip executable, handling cross-platform paths.
func (p *PythonPlugin) getPipPath() string {
	if core.IsWindows() {
		return filepath.Join("venv", "Scripts", "pip.exe")
	}
	return filepath.Join("venv", "bin", "pip")
}

// Install installs the specified Python dependencies along with transitive dependencies.
func (p *PythonPlugin) Install(deps []core.Dependency) error {
	// Check if virtual environment exists.
	_, err := os.Stat("venv")
	if os.IsNotExist(err) {
		logrus.Info("Creating Python virtual environment...")
		cmd := &core.Command{
			Name: "python3",
			Args: []string{"-m", "venv", "venv"},
		}
		if err := p.executor.Run(cmd); err != nil {
			logrus.Errorf("Failed to create virtual environment: %v", err)
			return err
		}
	} else if err != nil {
		return fmt.Errorf("error checking virtual environment: %v", err)
	}

	// Install dependencies
	for _, dep := range deps {
		pkgStr := fmt.Sprintf("%s%s", dep.Name, dep.Version)
		logrus.WithFields(logrus.Fields{
			"dependency": dep.Name,
			"version":    dep.Version,
		}).Info("Installing Python package")

		cmd := &core.Command{
			Name: p.getPipPath(),
			Args: []string{"install", pkgStr},
		}
		err := p.executor.Run(cmd) // Changed from output, err :=
		if err != nil {
			logrus.Errorf("Failed to install Python package '%s': %v", dep.Name, err)
			return err
		}
		logrus.Infof("Successfully installed Python package: %s", dep.Name)

		// Resolve transitive dependencies.
		transDeps, err := p.GetTransitiveDependencies(dep.Name, dep.Constraint.String())
		if err != nil {
			logrus.Warnf("Failed to get transitive dependencies for '%s': %v", dep.Name, err)
			continue
		}
		if len(transDeps) > 0 {
			logrus.Infof("Resolving transitive dependencies for '%s'", dep.Name)
			if err := p.Install(transDeps); err != nil {
				logrus.Errorf("Failed to install transitive dependencies for '%s': %v", dep.Name, err)
			}
		}
	}

	// Run security scans after installation.
	if err := p.RunSecurityScan(); err != nil {
		logrus.Warnf("Security scan encountered issues: %v", err)
	}

	return nil
}

// Update updates the specified Python dependencies.
func (p *PythonPlugin) Update(deps []core.Dependency) error {
	for _, dep := range deps {
		pkgStr := fmt.Sprintf("%s%s", dep.Name, dep.Version)
		logrus.WithFields(logrus.Fields{
			"dependency": dep.Name,
			"version":    dep.Version,
		}).Info("Updating Python package")

		cmd := &core.Command{
			Name: p.getPipPath(),
			Args: []string{"install", "--upgrade", pkgStr},
		}
		if err := p.executor.Run(cmd); err != nil {
			logrus.Errorf("Failed to update Python package '%s': %v", dep.Name, err)
			return err
		}
		logrus.Infof("Successfully updated Python package: %s", dep.Name)
	}

	// Run security scans after update.
	if err := p.RunSecurityScan(); err != nil {
		logrus.Warnf("Security scan encountered issues: %v", err)
	}

	return nil
}

// Remove removes the specified Python dependency.
func (p *PythonPlugin) Remove(dep core.Dependency) error {
	pkgStr := dep.Name
	logrus.WithFields(logrus.Fields{
		"dependency": dep.Name,
	}).Info("Removing Python package")

	cmd := &core.Command{
		Name: p.getPipPath(),
		Args: []string{"uninstall", "-y", pkgStr},
	}
	if err := p.executor.Run(cmd); err != nil {
		logrus.Errorf("Failed to remove Python package '%s': %v", dep.Name, err)
		return err
	}
	logrus.Infof("Successfully removed Python package: %s", dep.Name)

	return nil
}

// List lists all installed Python dependencies.
func (p *PythonPlugin) List() ([]core.Dependency, error) {
	cmd := &core.Command{
		Name: p.getPipPath(),
		Args: []string{"freeze"},
	}
	output, err := p.executor.Output(cmd)
	if err != nil {
		return nil, err
	}

	deps := []core.Dependency{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "==")
		if len(parts) != 2 {
			continue
		}
		deps = append(deps, core.Dependency{
			Name:    parts[0],
			Version: "=" + parts[1],
		})
	}

	return deps, nil
}

// ListVersions lists all available versions for a given Python package.
func (p *PythonPlugin) ListVersions(depName string) ([]string, error) {
	cmd := &core.Command{
		Name: p.getPipPath(),
		Args: []string{"install", fmt.Sprintf("%s==random", depName)}, // Intentional error to get available versions
	}

	output, err := p.executor.Output(cmd)
	if err == nil {
		return nil, fmt.Errorf("expected failure when listing versions")
	}

	outputStr := string(output)
	// Parse available versions from error message
	versions := []string{}
	prefix := "(from versions:"
	suffix := ")"
	start := strings.Index(outputStr, prefix)
	if start == -1 {
		return nil, fmt.Errorf("failed to parse available versions")
	}
	start += len(prefix)
	end := strings.Index(outputStr[start:], suffix)
	if end == -1 {
		return nil, fmt.Errorf("failed to parse available versions")
	}
	versionStr := outputStr[start : start+end]
	versionParts := strings.Split(versionStr, ",")
	for _, v := range versionParts {
		v = strings.TrimSpace(v)
		versions = append(versions, v)
	}
	return versions, nil
}

// GetTransitiveDependencies fetches transitive dependencies for a given dependency.
func (p *PythonPlugin) GetTransitiveDependencies(depName, version string) ([]core.Dependency, error) {
	cmd := &core.Command{
		Name: "pipdeptree",
		Args: []string{"--json-tree"},
	}
	output, err := p.executor.Output(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to run pipdeptree: %v", err)
	}

	var tree []map[string]interface{}
	if err := json.Unmarshal(output, &tree); err != nil {
		return nil, fmt.Errorf("failed to parse pipdeptree output: %v", err)
	}

	var transDeps []core.Dependency
	for _, pkg := range tree {
		if pkg["package"].(map[string]interface{})["name"] == depName {
			dependencies, ok := pkg["dependencies"].([]interface{})
			if !ok {
				continue
			}
			for _, d := range dependencies {
				depMap, ok := d.(map[string]interface{})
				if !ok {
					continue
				}
				name := depMap["package"].(map[string]interface{})["name"].(string)
				version := depMap["package"].(map[string]interface{})["version"].(string)
				transDeps = append(transDeps, core.Dependency{
					Name:    name,
					Version: "=" + version,
				})
			}
			break
		}
	}

	return transDeps, nil
}

// RunSecurityScan runs security scans on installed Python packages using 'safety'.
func (p *PythonPlugin) RunSecurityScan() error {
	logrus.Info("Running security scan on Python dependencies using 'safety'...")
	// Install safety if not installed
	cmd := &core.Command{
		Name: p.getPipPath(),
		Args: []string{"install", "safety"},
	}
	if err := p.executor.Run(cmd); err != nil {
		return fmt.Errorf("failed to install 'safety': %v", err)
	}

	// Run safety check
	scanCmd := &core.Command{
		Name: p.getPipPath(),
		Args: []string{"run", "safety", "check", "--json"},
	}
	scanOutput, err := p.executor.Output(scanCmd)
	if err != nil {
		return fmt.Errorf("security scan failed: %v", err)
	}

	var scanResults []core.SecurityVulnerability
	if err := json.Unmarshal(scanOutput, &scanResults); err != nil {
		return fmt.Errorf("failed to parse security scan results: %v", err)
	}

	if len(scanResults) > 0 {
		logrus.Warn("Vulnerabilities detected in Python dependencies:")
		for _, vuln := range scanResults {
			logrus.Warnf("- %s: %s", vuln.Package, vuln.Vulnerability)
		}
	} else {
		logrus.Info("No vulnerabilities found in Python dependencies.")
	}

	return nil
}

// GetVulnerabilities retrieves security vulnerabilities.
func (p *PythonPlugin) GetVulnerabilities() ([]core.SecurityVulnerability, error) {
	cmd := &core.Command{
		Name: p.getPipPath(),
		Args: []string{"run", "safety", "check", "--json"},
	}
	scanOutput, err := p.executor.Output(cmd)
	if err != nil {
		return nil, fmt.Errorf("security scan failed: %v", err)
	}

	var scanResults []core.SecurityVulnerability
	if err := json.Unmarshal(scanOutput, &scanResults); err != nil {
		return nil, fmt.Errorf("failed to parse security scan results: %v", err)
	}

	return scanResults, nil
}

// Cleanup cleans up Python plugin resources.
func (p *PythonPlugin) Cleanup() error {
	logrus.Info("Cleaning up Python plugin resources...")
	// Implement any necessary cleanup, such as removing virtual environments
	// For example:
	// return os.RemoveAll("venv")
	return nil
}
