package upd

import "errors"

var (
	ErrFileNotFound           = errors.New("package configuration file not found")
	ErrInvalidJSON            = errors.New("invalid JSON in package configuration file")
	ErrPackageNotFound        = errors.New("package not found in NPM registry")
	ErrVersionParse           = errors.New("failed to parse semantic version")
	ErrNoLatestDistTag        = errors.New("no \"latest\" dist-tag found")
	ErrNoValidVersions        = errors.New("no valid versions found")
	ErrNoSemverVersions       = errors.New("no valid semver versions found")
	ErrSectionNotFound        = errors.New("section not found in package configuration file")
	ErrSectionNotObject       = errors.New("section is not a JSON object")
	ErrDependencyNotFound     = errors.New("dependency not found in package configuration file")
	ErrConcurrentModification = errors.New("package configuration file was modified concurrently since read")
)
