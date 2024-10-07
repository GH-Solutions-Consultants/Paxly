// core/resolver.go
package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// Resolver handles dependency resolution.
type Resolver struct {
	Config         Config
	PluginRegistry *PluginRegistry
	resolvedDeps   map[string]*semver.Version
	processingDeps map[string]bool // To detect cycles
}

// NewResolver creates a new Resolver instance.
func NewResolver(config Config, pr *PluginRegistry) *Resolver {
	return &Resolver{
		Config:         config,
		PluginRegistry: pr,
		resolvedDeps:   make(map[string]*semver.Version),
		processingDeps: make(map[string]bool),
	}
}

// ResolveDependencies resolves all dependencies recursively for a given environment.
func (r *Resolver) ResolveDependencies(env string) error {
	envConfig, exists := r.Config.Environments[env]
	if !exists {
		return fmt.Errorf("environment '%s' not found in configuration", env)
	}

	for lang, deps := range envConfig.Dependencies {
		for _, dep := range deps {
			if err := r.resolveDependency(lang, dep); err != nil {
				return err
			}
		}
	}

	return nil
}

// resolveDependency resolves a single dependency and its transitive dependencies.
func (r *Resolver) resolveDependency(lang string, dep Dependency) error {
	if err := dep.Validate(); err != nil {
		return fmt.Errorf("invalid dependency '%s' in language '%s': %v", dep.Name, lang, err)
	}

	// Check for cycles.
	if r.processingDeps[dep.Name] {
		return fmt.Errorf("cyclic dependency detected on '%s'", dep.Name)
	}

	// If already resolved, verify version compatibility.
	if existingVersion, exists := r.resolvedDeps[dep.Name]; exists {
		if !dep.Constraint.Check(existingVersion) {
			return fmt.Errorf("version conflict for '%s': existing version '%s' does not satisfy constraint '%s'", dep.Name, existingVersion, dep.Version)
		}
		return nil // Already resolved and compatible.
	}

	// Mark as processing.
	r.processingDeps[dep.Name] = true
	defer delete(r.processingDeps, dep.Name)

	// Fetch the latest compatible version.
	version, err := r.getLatestCompatibleVersion(lang, dep)
	if err != nil {
		return err
	}
	r.resolvedDeps[dep.Name] = version

	// Fetch transitive dependencies.
	transDeps, err := r.getTransitiveDependencies(lang, dep.Name, version)
	if err != nil {
		return err
	}

	// Recursively resolve transitive dependencies.
	for _, tDep := range transDeps {
		if err := r.resolveDependency(lang, tDep); err != nil {
			return err
		}
	}

	return nil
}

// getLatestCompatibleVersion retrieves the latest compatible version of a dependency.
func (r *Resolver) getLatestCompatibleVersion(lang string, dep Dependency) (*semver.Version, error) {
	plugin, exists := r.PluginRegistry.GetPlugin(lang)
	if !exists {
		return nil, fmt.Errorf("no plugin found for language '%s'", lang)
	}

	availableVersions, err := plugin.ListVersions(dep.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to list available versions for '%s': %v", dep.Name, err)
	}

	var compatibleVersions []*semver.Version
	for _, vStr := range availableVersions {
		v, err := semver.NewVersion(vStr)
		if err != nil {
			continue
		}
		if dep.Constraint.Check(v) {
			compatibleVersions = append(compatibleVersions, v)
		}
	}

	if len(compatibleVersions) == 0 {
		return nil, fmt.Errorf("no compatible versions found for '%s' with constraint '%s'", dep.Name, dep.Version)
	}

	// Sort the versions in ascending order and select the latest.
	sort.Slice(compatibleVersions, func(i, j int) bool {
		return compatibleVersions[i].LessThan(compatibleVersions[j])
	})
	latestVersion := compatibleVersions[len(compatibleVersions)-1]
	return latestVersion, nil
}

// getTransitiveDependencies fetches transitive dependencies for a given dependency.
func (r *Resolver) getTransitiveDependencies(lang, depName string, version *semver.Version) ([]Dependency, error) {
	plugin, exists := r.PluginRegistry.GetPlugin(lang)
	if !exists {
		return nil, fmt.Errorf("no plugin found for language '%s'", lang)
	}

	transDeps, err := plugin.GetTransitiveDependencies(depName, version.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get transitive dependencies for '%s@%s': %v", depName, version, err)
	}

	return transDeps, nil
}

// SecurityReport represents a security report for a language.
type SecurityReport struct {
	Language        string                 `json:"language"`
	Vulnerabilities []SecurityVulnerability `json:"vulnerabilities"`
}

// CollectSecurityReports collects security vulnerabilities from each plugin and generates reports.
func (r *Resolver) CollectSecurityReports() error {
	var securityReports []SecurityReport

	plugins := r.PluginRegistry.GetAllPlugins()
	for lang, plugin := range plugins {
		// Each plugin should expose a method to retrieve vulnerabilities
		vulns, err := plugin.GetVulnerabilities()
		if err != nil {
			logrus.Errorf("Failed to get vulnerabilities for '%s': %v", lang, err)
			continue
		}

		securityReports = append(securityReports, SecurityReport{
			Language:        lang,
			Vulnerabilities: vulns,
		})
	}

	// Generate security report in JSON
	if err := GenerateSecurityReport(securityReports, "json", "security_report.json"); err != nil {
		logrus.Errorf("Failed to generate JSON security report: %v", err)
		return err
	}

	// Generate security report in HTML
	if err := GenerateSecurityReport(securityReports, "html", "security_report.html"); err != nil {
		logrus.Errorf("Failed to generate HTML security report: %v", err)
		return err
	}

	return nil
}

// GenerateSecurityReport generates a security report in the specified format.
func GenerateSecurityReport(reports []SecurityReport, format, outputPath string) error {
	var data []byte
	var err error

	switch format {
	case "json":
		data, err = json.MarshalIndent(reports, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal security report to JSON: %v", err)
		}
	case "html":
		// Simple HTML template
		var buffer bytes.Buffer
		buffer.WriteString("<html><head><title>Security Report</title></head><body><h1>Security Report</h1>")
		for _, report := range reports {
			buffer.WriteString(fmt.Sprintf("<h2>%s</h2>", report.Language))
			if len(report.Vulnerabilities) == 0 {
				buffer.WriteString("<p>No vulnerabilities found.</p>")
				continue
			}
			buffer.WriteString("<ul>")
			for _, vuln := range report.Vulnerabilities {
				buffer.WriteString(fmt.Sprintf("<li><strong>%s:</strong> %s (Severity: %s)</li>", vuln.Package, vuln.Vulnerability, vuln.Severity))
			}
			buffer.WriteString("</ul>")
		}
		buffer.WriteString("</body></html>")
		data = buffer.Bytes()
	default:
		return fmt.Errorf("unsupported report format '%s'", format)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write security report to '%s': %v", outputPath, err)
	}

	logrus.Infof("Security report generated at '%s'", outputPath)
	return nil
}
