// core/dependency.go

package core

import (
    "fmt"
	"regexp"
    "strconv"
    "github.com/Masterminds/semver/v3"
    "github.com/go-playground/validator/v10"
    "github.com/sirupsen/logrus"
)

// Dependency represents a package dependency.
type Dependency struct {
    Name             string              `yaml:"name" validate:"required"`
    Version          string              `yaml:"version" validate:"required"`
    PEP440Constraint string              `yaml:"-"` // New field for PEP 440 constraints
    Constraint       *semver.Constraints `yaml:"-"`
}

// Validate parses and validates the dependency.
func (d *Dependency) Validate() error {
    validate := validator.New()
    validate.RegisterValidation("semver", validateSemVer)
    err := validate.Struct(d)
    if err != nil {
        logrus.Errorf("Dependency validation failed for '%s': %v", d.Name, err)
        return err
    }

    // Translate SemVer to PEP 440
    pep440Constraint, err := translateSemVerToPEP440(d.Version)
    if err != nil {
        logrus.Errorf("Failed to translate SemVer constraint '%s' for '%s': %v", d.Version, d.Name, err)
        return err
    }
    d.PEP440Constraint = pep440Constraint

    constraint, err := semver.NewConstraint(d.Version)
    if err != nil {
        logrus.Errorf("Invalid version constraint '%s' for dependency '%s': %v", d.Version, d.Name, err)
        return fmt.Errorf("invalid version constraint '%s': %v", d.Version, err)
    }
    d.Constraint = constraint
    logrus.Debugf("Set Constraint for '%s' to '%s'", d.Name, d.Constraint.String())
    return nil
}

// Helper function to translate SemVer to PEP 440
func translateSemVerToPEP440(semverConstraint string) (string, error) {
    // Example: '^1.21' -> '>=1.21.0,<2.0.0'
    re := regexp.MustCompile(`\^(?P<major>\d+)\.(?P<minor>\d+)`)
    matches := re.FindStringSubmatch(semverConstraint)
    if len(matches) < 3 {
        return "", fmt.Errorf("unsupported SemVer constraint: %s", semverConstraint)
    }
    major := matches[1]
    minor := matches[2]
    upperMajor := fmt.Sprintf("%d", atoi(major)+1)
    pep440 := fmt.Sprintf(">=%s.%s.0,<%s.0.0", major, minor, upperMajor)
    return pep440, nil
}

// Simple helper to convert string to int
func atoi(s string) int {
    i, _ := strconv.Atoi(s)
    return i
}


// validateSemVer ensures the version string adheres to semantic versioning.
func validateSemVer(fl validator.FieldLevel) bool {
    _, err := semver.NewConstraint(fl.Field().String())
    return err == nil
}
