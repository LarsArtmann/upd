package internal

import "errors"

var (
	ErrFileNotFound    = errors.New("package configuration file not found")
	ErrInvalidJSON     = errors.New("invalid JSON in package configuration file")
	ErrPackageNotFound = errors.New("package not found in NPM registry")
	ErrVersionParse    = errors.New("failed to parse semantic version")
)
