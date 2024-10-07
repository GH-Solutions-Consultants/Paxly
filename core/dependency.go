// core/dependency.go
package core

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/go-playground/validator/v10"
)

// Dependency represents a package dependency.
type Dependency struct {
	Name       string              `yaml:"name" validate:"required"`
	Version    string              `yaml:"version" validate:"required"`
	Constraint *semver.Constraints `yaml:"-"`
}

// Validate parses and validates the dependency.
func (d *Dependency) Validate() error {
	validate := validator.New()
	// Register custom validation for semantic versioning
	validate.RegisterValidation("semver", validateSemVer)

	err := validate.Struct(d)
	if err != nil {
		return err
	}

	// Parse semantic version constraint
	constraint, err := semver.NewConstraint(d.Version)
	if err != nil {
		return fmt.Errorf("invalid version constraint '%s': %v", d.Version, err)
	}
	d.Constraint = constraint

	return nil
}

// validateSemVer ensures the version string adheres to semantic versioning.
func validateSemVer(fl validator.FieldLevel) bool {
	_, err := semver.NewConstraint(fl.Field().String())
	return err == nil
}
