package upd

import (
	errorfamily "github.com/larsartmann/go-error-family"
)

// Domain errors — classified by behavioral family.
// Rejection = caller's fault (not found, bad input). Exit 1.
// Transient = temporary, retryable. Exit 75.
// Corruption = data damaged. Exit 65 (EX_DATAERR).
// Conflict = state mismatch. Exit 1.
//
// Control-flow signals (ErrHelp, ErrVersion) live in config.go.
var (
	ErrFileNotFound    = errorfamily.NewRejection("file.not_found", "package configuration file not found")
	ErrInvalidJSON     = errorfamily.NewCorruption("json.invalid", "invalid JSON in package configuration file")
	ErrPackageNotFound = errorfamily.NewRejection(
		"registry.package_not_found",
		"package not found in NPM registry",
	)
	ErrRegistryUnavailable = errorfamily.NewTransient("registry.unavailable", "NPM registry is unavailable")
	ErrVersionParse        = errorfamily.NewCorruption("version.parse_failed", "failed to parse semantic version")
	ErrNoLatestDistTag     = errorfamily.NewCorruption("version.no_latest", "no \"latest\" dist-tag found")
	ErrNoValidVersions     = errorfamily.NewCorruption("version.no_versions", "no valid versions found")
	ErrNoSemverVersions    = errorfamily.NewCorruption("version.no_semver", "no valid semver versions found")
	ErrSectionNotFound     = errorfamily.NewRejection(
		"json.section_missing",
		"section not found in package configuration file",
	)
	ErrSectionNotObject   = errorfamily.NewCorruption("json.section_not_object", "section is not a JSON object")
	ErrDependencyNotFound = errorfamily.NewRejection(
		"json.dependency_missing",
		"dependency not found in package configuration file",
	)
	ErrConcurrentModification = errorfamily.NewConflict(
		"file.concurrent_modification",
		"package configuration file was modified concurrently since read",
	)
	ErrPartialFailure = errorfamily.NewRejection(
		"update.partial_failure",
		"one or more dependencies could not be resolved",
	)
)
