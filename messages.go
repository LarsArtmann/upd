package upd

import (
	errorfamily "github.com/larsartmann/go-error-family"
)

// registerFileAndJSONTemplates registers message templates for file and JSON errors.
//
//nolint:exhaustruct // MessageTemplate fields Why/WayOut are optional
func registerFileAndJSONTemplates() {
	errorfamily.DefaultRegistry.RegisterTemplates(map[string]errorfamily.MessageTemplate{
		"file.not_found": {
			What: "Package configuration file was not found at {path}.",
			Fix:  "Check that the file path is correct and the file exists.",
		},
		"file.write_failed": {
			What: "Could not write to the package configuration file at {path}.",
			Fix:  "Check file permissions and disk space.",
		},
		"file.concurrent_modification": {
			What:   "The package configuration file was modified by another process while upd was running.",
			Why:    "Another tool (npm install, IDE formatter, etc.) changed the file during upd's network fetch window.",
			Fix:    "Your file was not changed. Re-run upd to try again.",
			WayOut: "If this keeps happening, close other tools that watch package.json before running upd.",
		},
		"json.invalid": {
			What: "The package configuration file contains invalid JSON.",
			Fix:  "Validate the JSON syntax (e.g. run 'npx jsonlint package.json').",
		},
		"json.section_not_object": {
			What: "Section {section} is {kind}, expected a JSON object.",
			Fix:  "Ensure the section contains key-value pairs of package names to version strings.",
		},
		"json.section_missing": {
			What: "Section {section} was not found in the package configuration file.",
			Fix:  "Add the section or remove it from upd's scope.",
		},
		"json.dependency_missing": {
			What: "Dependency {dependency} was not found in section {section}.",
			Fix:  "Check the dependency name and section.",
		},
	})
}

// registerRegistryAndVersionTemplates registers message templates for registry and version errors.
//
//nolint:exhaustruct // MessageTemplate fields Why/WayOut are optional
func registerRegistryAndVersionTemplates() {
	errorfamily.DefaultRegistry.RegisterTemplates(map[string]errorfamily.MessageTemplate{
		"registry.package_not_found": {
			What: "Package {package} was not found in the NPM registry.",
			Why:  "The registry returned status {status} for this package name.",
			Fix:  "Check the package name for typos. If the package was renamed or removed, update your dependency.",
		},
		"registry.unavailable": {
			What:   "The NPM registry is temporarily unavailable (status {status} for package {package}).",
			Why:    "This is a transient infrastructure failure — no data was lost.",
			Fix:    "Wait a moment and try again.",
			WayOut: "If this persists, check https://status.npmjs.org/ or try a mirror with -r <registry-url>.",
		},
		"registry.fetch_aborted": {
			What:   "The fetch was cancelled (likely by Ctrl+C).",
			Fix:    "Re-run upd when ready.",
			WayOut: "Use --timeout to increase the per-request timeout if the registry is slow.",
		},
		"version.parse_failed": {
			What: "Failed to parse a semantic version string.",
			Fix:  "Check the version constraint in package.json for malformed semver.",
		},
		"version.no_latest": {
			What: "The packument has no \"latest\" dist-tag.",
			Fix:  "Use --greatest to select the highest semver version instead.",
		},
		"version.no_versions": {
			What: "The packument has no versions field.",
			Fix:  "The package may be empty or deprecated. Try --greatest or remove the dependency.",
		},
		"version.no_semver": {
			What: "The packument has no valid semver versions.",
			Fix:  "The package may only contain pre-release or tag versions. Remove the dependency if unused.",
		},
		"update.partial_failure": {
			What:   "{error_count} package(s) could not be resolved.",
			Why:    "Some dependencies encountered errors during version resolution. Others were updated successfully.",
			Fix:    "Check the error details above. The file was written with successful updates.",
			WayOut: "Re-run upd after fixing the failing dependencies.",
		},
		"cli.parse_flags": {
			What: "Could not parse command-line flags.",
			Fix:  "Run 'upd --help' for usage information.",
		},
	})
}

//
//nolint:gochecknoinits // Register error templates before main() runs.
func init() {
	registerFileAndJSONTemplates()
	registerRegistryAndVersionTemplates()
}
