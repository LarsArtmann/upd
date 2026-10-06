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
// Sentinels double as parameterized error templates via the WithContext
// factory idiom, which requires the concrete type; declaring them as the
// error interface (erraudit's suggestion) would break all 11 call sites.
var (
	ErrFileNotFound = errorfamily.NewRejection(
		"file.not_found",
		"package configuration file not found",
	) //nolint:erraudit
	ErrInvalidJSON = errorfamily.NewCorruption(
		"json.invalid",
		"invalid JSON in package configuration file",
	) //nolint:erraudit
	ErrPackageNotFound = errorfamily.NewRejection( //nolint:erraudit
		"registry.package_not_found",
		"package not found in NPM registry",
	)
	ErrRegistryUnavailable = errorfamily.NewTransient(
		"registry.unavailable",
		"NPM registry is unavailable",
	) //nolint:erraudit
	ErrVersionParse = errorfamily.NewCorruption(
		"version.parse_failed",
		"failed to parse semantic version",
	) //nolint:erraudit
	ErrNoLatestDistTag = errorfamily.NewCorruption(
		"version.no_latest",
		"no \"latest\" dist-tag found",
	) //nolint:erraudit
	ErrNoValidVersions = errorfamily.NewCorruption(
		"version.no_versions",
		"no valid versions found",
	) //nolint:erraudit
	ErrNoSemverVersions = errorfamily.NewCorruption(
		"version.no_semver",
		"no valid semver versions found",
	) //nolint:erraudit
	ErrSectionNotFound = errorfamily.NewRejection( //nolint:erraudit
		"json.section_missing",
		"section not found in package configuration file",
	)
	ErrSectionNotObject = errorfamily.NewCorruption(
		"json.section_not_object",
		"section is not a JSON object",
	) //nolint:erraudit
	ErrDependencyNotFound = errorfamily.NewRejection( //nolint:erraudit
		"json.dependency_missing",
		"dependency not found in package configuration file",
	)
	ErrConcurrentModification = errorfamily.NewConflict( //nolint:erraudit
		"file.concurrent_modification",
		"package configuration file was modified concurrently since read",
	)
	ErrPartialFailure = errorfamily.NewRejection( //nolint:erraudit
		"update.partial_failure",
		"one or more dependencies could not be resolved",
	)
)
