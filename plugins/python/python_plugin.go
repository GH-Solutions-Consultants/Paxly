// plugins/python/python_plugin.go
package python

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
    "net/http"
    "sort"
    "time"

	"github.com/GH-Solutions-Consultants/Paxly/core"
	"github.com/sirupsen/logrus"
    "github.com/Masterminds/semver/v3"
)

// Ensure the PythonPlugin implements the PackageManagerPlugin interface.
var _ core.PackageManagerPlugin = (*PythonPlugin)(nil)

// PythonPlugin is the plugin for managing Python dependencies.
type PythonPlugin struct {
	executor core.Executor
}

// Register the PythonPlugin with the PluginRegistry during initialization.
func init() {
	plugin := NewPythonPlugin(nil)
	err := core.GetPluginRegistry().RegisterPlugin(plugin.Language(), plugin)
	if err != nil {
		logrus.Fatalf("Failed to register Python plugin: %v", err)
	}
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

// getPythonPath determines the appropriate Python executable based on the OS.
func (p *PythonPlugin) getPythonPath() string {
    var pythonPath string
    if runtime.GOOS == "windows" {
        pythonPath = filepath.Join("venv", "Scripts", "python.exe")
    } else {
        pythonPath = filepath.Join("venv", "bin", "python3")
    }
    if _, err := os.Stat(pythonPath); err == nil {
        return pythonPath
    }
    // Fallback to system python
    if runtime.GOOS == "windows" {
        return "python"
    }
    return "python3"
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
	pythonCmd := p.getPythonPath()

	// Validate Python installation
	if err := p.executor.Run(&core.Command{Name: pythonCmd, Args: []string{"--version"}}); err != nil {
		return fmt.Errorf("%s is not installed or not in PATH", pythonCmd)
	}

	pipPath := p.getPipPath()
	if err := p.executor.Run(&core.Command{Name: pipPath, Args: []string{"--version"}}); err != nil {
		return fmt.Errorf("pip is not installed or not in PATH")
	}

	// Ensure pipdeptree is installed
	if err := p.ensurePipDeptree(); err != nil {
		return err
	}
	return nil
}

// ensurePipDeptree ensures that pipdeptree is installed.
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
    var pipPath string
    if runtime.GOOS == "windows" {
        pipPath = filepath.Join("venv", "Scripts", "pip.exe")
    } else {
        pipPath = filepath.Join("venv", "bin", "pip3")
    }
    if _, err := os.Stat(pipPath); err == nil {
        return pipPath
    }
    // Fallback to system pip
    if runtime.GOOS == "windows" {
        return "pip"
    }
    return "pip3"
}

// getSafetyPath returns the path to the safety executable, handling cross-platform paths.
func (p *PythonPlugin) getSafetyPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join("venv", "Scripts", "safety.exe")
	}
	return filepath.Join("venv", "bin", "safety")
}

// Install installs the specified Python dependencies along with transitive dependencies.
func (p *PythonPlugin) Install(deps []core.Dependency) error {
	// Check if virtual environment exists.
	_, err := os.Stat("venv")
	if os.IsNotExist(err) {
		logrus.Info("Creating Python virtual environment...")
		cmd := &core.Command{
			Name: p.getPythonPath(),
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
		err := p.executor.Run(cmd)
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
	logrus.Infof("Running 'pip freeze' to list dependencies")
	cmd := &core.Command{
		Name: p.getPipPath(),
		Args: []string{"freeze"},
	}
	output, err := p.executor.Output(cmd)
	if err != nil {
		logrus.Errorf("Failed to run 'pip freeze': %v", err)
		return nil, fmt.Errorf("failed to list Python dependencies: %v", err)
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

func (p *PythonPlugin) ListVersions(depName string) ([]string, error) {
    url := fmt.Sprintf("https://pypi.org/pypi/%s/json", depName)
    client := &http.Client{
        Timeout: 10 * time.Second,
    }
    logrus.Debugf("Fetching package info from PyPI: %s", url)
    resp, err := client.Get(url)
    if err != nil {
        logrus.Errorf("HTTP request failed: %v", err)
        return nil, fmt.Errorf("failed to fetch package info from PyPI for '%s': %v", depName, err)
    }
    defer resp.Body.Close()

    logrus.Debugf("Received response: Status Code %d", resp.StatusCode)
    if resp.StatusCode != http.StatusOK {
        logrus.Errorf("Non-OK HTTP status: %d", resp.StatusCode)
        return nil, fmt.Errorf("failed to fetch package info from PyPI for '%s': Status Code %d", depName, resp.StatusCode)
    }

    var data struct {
        Releases map[string][]interface{} `json:"releases"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        logrus.Errorf("JSON decoding failed: %v", err)
        return nil, fmt.Errorf("failed to parse PyPI response for '%s': %v", depName, err)
    }

    versions := make([]string, 0, len(data.Releases))
    for version := range data.Releases {
        versions = append(versions, version)
    }

    logrus.Debugf("Fetched versions for '%s': %v", depName, versions)

    // Sort versions in ascending order
    sort.Slice(versions, func(i, j int) bool {
        vi, err1 := semver.NewVersion(versions[i])
        vj, err2 := semver.NewVersion(versions[j])
        if err1 != nil || err2 != nil {
            // Fallback to string comparison if semver parsing fails
            return versions[i] < versions[j]
        }
        return vi.LessThan(vj)
    })

    logrus.Debugf("Sorted versions for '%s': %v", depName, versions)

    if len(versions) == 0 {
        logrus.Error("No versions found for the package")
        return nil, fmt.Errorf("no versions found for '%s'", depName)
    }

    return versions, nil
}



// Helper function to parse 'pip index versions' plain text output
func parsePipIndexVersionsOutput(output []byte) ([]string, error) {
    outputStr := string(output)
    lines := strings.Split(outputStr, "\n")
    var versions []string
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "* ") {
            version := strings.TrimPrefix(line, "* ")
            version = strings.TrimSpace(version)
            if version != "" {
                versions = append(versions, version)
            }
        }
    }
    if len(versions) == 0 {
        return nil, fmt.Errorf("no versions found in the output")
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
		pkgInfo, ok := pkg["package"].(map[string]interface{})
		if !ok {
			continue
		}
		if pkgInfo["name"] != depName {
			continue
		}
		dependencies, ok := pkg["dependencies"].([]interface{})
		if !ok {
			continue
		}
		for _, d := range dependencies {
			depMap, ok := d.(map[string]interface{})
			if !ok {
				continue
			}
			packageInfo, ok := depMap["package"].(map[string]interface{})
			if !ok {
				continue
			}
			name, okName := packageInfo["name"].(string)
			version, okVersion := packageInfo["version"].(string)
			if !okName || !okVersion {
				continue
			}
			transDeps = append(transDeps, core.Dependency{
				Name:    name,
				Version: "=" + version,
			})
		}
		break
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

	// Run safety check using the absolute path to the safety executable
	scanCmd := &core.Command{
		Name: p.getSafetyPath(),
		Args: []string{"check", "--json"},
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
		Name: p.getSafetyPath(),
		Args: []string{"check", "--json"},
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
	// Example: Uncomment the following line to remove the virtual environment
	// return os.RemoveAll("venv")
	return nil
}