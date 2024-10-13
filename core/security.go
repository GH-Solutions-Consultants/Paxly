// core/security.go
package core

// SecurityVulnerability represents a security vulnerability in a package
type SecurityVulnerability struct {
    PackageName     string
    VulnerabilityID string
    Description     string
    Severity        string
}

