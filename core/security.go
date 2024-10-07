// core/security.go
package core

// SecurityVulnerability represents a security issue found in a dependency.
type SecurityVulnerability struct {
	Package       string `json:"package"`
	Vulnerability string `json:"vulnerability"`
	Severity      string `json:"severity,omitempty"`
	Description   string `json:"description,omitempty"`
}