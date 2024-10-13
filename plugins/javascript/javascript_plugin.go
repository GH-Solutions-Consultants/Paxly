// plugins/javascript/javascript_plugin.go
package javascript

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/GH-Solutions-Consultants/Paxly/core"
	"github.com/sirupsen/logrus"
)

// Ensure the JavaScriptPlugin implements the PackageManagerPlugin interface.
var _ core.PackageManagerPlugin = (*JavaScriptPlugin)(nil)

// JavaScriptPlugin is the plugin for managing JavaScript dependencies.
type JavaScriptPlugin struct {
	executor core.Executor
}

// Register the JavaScriptPlugin with the PluginRegistry during initialization.
func init() {
	plugin := NewJavaScriptPlugin(nil)
	err := core.GetPluginRegistry().RegisterPlugin(plugin.Language(), plugin)
	if err != nil {
		logrus.Fatalf("Failed to register JavaScript plugin: %v", err)
	}
}

// NewJavaScriptPlugin creates a new instance of JavaScriptPlugin with the given executor.
func NewJavaScriptPlugin(executor core.Executor) *JavaScriptPlugin {
	if executor == nil {
		executor = &core.RealExecutor{}
	}
	return &JavaScriptPlugin{
		executor: executor,
	}
}

// APIVersion returns the plugin API version.
func (p *JavaScriptPlugin) APIVersion() string {
	return core.PluginAPIVersion
}

// Language returns the name of the language this plugin manages.
func (p *JavaScriptPlugin) Language() string {
	return "javascript"
}

// Initialize sets up the JavaScript plugin with necessary configurations.
func (p *JavaScriptPlugin) Initialize(config core.Config) error {
	logrus.Info("Initializing JavaScript plugin...")
	// Validate npm installation
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("npm is not installed or not in PATH")
	}
	// Ensure package.json exists; if not, initialize it
	if _, err := os.Stat("package.json"); os.IsNotExist(err) {
		logrus.Info("package.json not found. Initializing npm project...")
		cmd := exec.Command("npm", "init", "-y")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to initialize npm project: %v", err)
		}
	}
	return nil
}

// Install installs the specified JavaScript dependencies.
func (p *JavaScriptPlugin) Install(deps []core.Dependency) error {
	for _, dep := range deps {
		pkgStr := fmt.Sprintf("%s@%s", dep.Name, dep.Version)
		logrus.WithFields(logrus.Fields{
			"dependency": dep.Name,
			"version":    dep.Version,
		}).Info("Installing JavaScript package")

		cmd := core.Command{
			Name: "npm",
			Args: []string{"install", pkgStr},
		}
		err := p.executor.Run(&cmd) // Correct assignment
		if err != nil {
			logrus.Errorf("Failed to install JavaScript package '%s': %v", dep.Name, err)
			return err
		}
		logrus.Infof("Successfully installed JavaScript package: %s", dep.Name)
	}

	// Optionally, run 'npm audit' here if integrated

	return nil
}

// Update updates the specified JavaScript dependencies.
func (p *JavaScriptPlugin) Update(deps []core.Dependency) error {
	for _, dep := range deps {
		pkgStr := fmt.Sprintf("%s@%s", dep.Name, dep.Version)
		logrus.WithFields(logrus.Fields{
			"dependency": dep.Name,
			"version":    dep.Version,
		}).Info("Updating JavaScript package")

		cmd := core.Command{
			Name: "npm",
			Args: []string{"install", pkgStr},
		}
		err := p.executor.Run(&cmd) // Correct assignment
		if err != nil {
			logrus.Errorf("Failed to update JavaScript package '%s': %v", dep.Name, err)
			return err
		}
		logrus.Infof("Successfully updated JavaScript package: %s", dep.Name)
	}

	return nil
}

// Remove removes the specified JavaScript dependency.
func (p *JavaScriptPlugin) Remove(dep core.Dependency) error {
	pkgStr := dep.Name
	logrus.WithFields(logrus.Fields{
		"dependency": dep.Name,
	}).Info("Removing JavaScript package")

	cmd := core.Command{
		Name: "npm",
		Args: []string{"uninstall", pkgStr},
	}
	err := p.executor.Run(&cmd) // Correct assignment
	if err != nil {
		logrus.Errorf("Failed to remove JavaScript package '%s': %v", dep.Name, err)
		return err
	}

	logrus.Infof("Successfully removed JavaScript package: %s", dep.Name)
	return nil
}

// List lists all installed JavaScript dependencies.
func (p *JavaScriptPlugin) List() ([]core.Dependency, error) {
	cmd := core.Command{
		Name: "npm",
		Args: []string{"list", "--json", "--depth=0"},
	}
	output, err := p.executor.Output(&cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list JavaScript packages: %v", err)
	}

	var listOutput struct {
		Dependencies map[string]interface{} `json:"dependencies"`
	}

	if err := json.Unmarshal(output, &listOutput); err != nil {
		return nil, fmt.Errorf("failed to parse npm list output: %v", err)
	}

	deps := []core.Dependency{}
	for name, info := range listOutput.Dependencies {
		version, ok := info.(map[string]interface{})["version"].(string)
		if !ok {
			continue
		}
		deps = append(deps, core.Dependency{
			Name:    name,
			Version: "=" + version,
		})
	}

	return deps, nil
}

// ListVersions lists all available versions for a given JavaScript dependency.
func (p *JavaScriptPlugin) ListVersions(depName string) ([]string, error) {
	cmd := core.Command{
		Name: "npm",
		Args: []string{"view", depName, "versions", "--json"},
	}
	output, err := p.executor.Output(&cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions for '%s': %v", depName, err)
	}

	var versions []string
	if err := json.Unmarshal(output, &versions); err != nil {
		return nil, fmt.Errorf("failed to parse npm versions: %v", err)
	}

	return versions, nil
}

// GetTransitiveDependencies fetches transitive dependencies for a given dependency.
func (p *JavaScriptPlugin) GetTransitiveDependencies(depName, version string) ([]core.Dependency, error) {
    logrus.Infof("Running 'npm ls' for package %s", depName)
    cmd := core.Command{
        Name: "npm",
        Args: []string{"ls", depName, "--json", "--depth=0"},
    }
    output, err := p.executor.Output(&cmd)
    if err != nil {
        logrus.Warnf("Failed to list dependencies for '%s', skipping transitive dependencies: %v", depName, err)
        logrus.Warnf("Output: %s", string(output))
        return []core.Dependency{}, nil
    }

    // Parse the output
    var lsOutput struct {
        Dependencies map[string]struct {
            Version string `json:"version"`
        } `json:"dependencies"`
    }
    if err := json.Unmarshal(output, &lsOutput); err != nil {
        logrus.Warnf("Failed to parse npm ls output for '%s', skipping transitive dependencies: %v", depName, err)
        return []core.Dependency{}, nil
    }

    var transDeps []core.Dependency
    for name, info := range lsOutput.Dependencies {
        transDeps = append(transDeps, core.Dependency{
            Name:    name,
            Version: "=" + info.Version,
        })
    }

    return transDeps, nil
}

// GetVulnerabilities retrieves security vulnerabilities using 'npm audit --json'.
func (p *JavaScriptPlugin) GetVulnerabilities() ([]core.SecurityVulnerability, error) {
	cmd := core.Command{
		Name: "npm",
		Args: []string{"audit", "--json"},
	}
	output, err := p.executor.Output(&cmd)
	if err != nil {
		// npm audit returns non-zero exit code if vulnerabilities are found
		// We still need to parse the output
		if exitErr, ok := err.(*exec.ExitError); ok {
			output = exitErr.Stderr
		} else {
			return nil, fmt.Errorf("failed to run npm audit: %v", err)
		}
	}

	var auditOutput struct {
		Vulnerabilities map[string]struct {
			Title          string   `json:"title"`
			ModuleName     string   `json:"module_name"`
			Severity       string   `json:"severity"`
			Overview       string   `json:"overview"`
			PatchedBy      []string `json:"patched_by"`
			Recommendation string   `json:"recommendation"`
		} `json:"vulnerabilities"`
	}

	if err := json.Unmarshal(output, &auditOutput); err != nil {
		return nil, fmt.Errorf("failed to parse npm audit output: %v", err)
	}

	vulns := []core.SecurityVulnerability{}
	for _, vuln := range auditOutput.Vulnerabilities {
		vulns = append(vulns, core.SecurityVulnerability{
			PackageName:     vuln.ModuleName,
			VulnerabilityID: vuln.Title,
			Severity:        vuln.Severity,
			Description:     vuln.Overview,
		})
	}

	return vulns, nil
}

// Cleanup performs any necessary cleanup operations.
func (p *JavaScriptPlugin) Cleanup() error {
	logrus.Info("Cleaning up JavaScript plugin resources...")
	// Implement any necessary cleanup
	return nil
}