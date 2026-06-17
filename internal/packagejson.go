package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/tidwall/gjson"
)

type PackageFile struct {
	raw []byte
}

func ReadPackageFile(path string) (*PackageFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("cannot find NPM package configuration file %q: %w", path, ErrFileNotFound)
		}

		return nil, fmt.Errorf("failed to read package configuration file %q: %w", path, err)
	}

	if !gjson.ValidBytes(data) {
		return nil, fmt.Errorf("failed to parse package configuration file %q: %w", path, ErrInvalidJSON)
	}

	return &PackageFile{raw: data}, nil
}

func (p *PackageFile) Raw() []byte {
	return p.raw
}

func (p *PackageFile) GetDependencySection(section string) map[string]string {
	deps := make(map[string]string)

	result := gjson.GetBytes(p.raw, section)
	if !result.IsObject() {
		return deps
	}

	result.ForEach(func(key, value gjson.Result) bool {
		if value.Type == gjson.String {
			deps[key.String()] = value.String()
		}

		return true
	})

	return deps
}

func (p *PackageFile) GetUpdArgs() []string {
	result := gjson.GetBytes(p.raw, "upd")
	if !result.Exists() {
		return nil
	}

	if result.IsArray() {
		var args []string

		result.ForEach(func(_, v gjson.Result) bool {
			if v.Type == gjson.String {
				args = append(args, v.String())
			}

			return true
		})

		return args
	}

	if result.Type == gjson.String {
		return strings.Fields(result.String())
	}

	return nil
}

func (p *PackageFile) UpdateDependency(section, name, newValue string) error {
	sectionResult := gjson.GetBytes(p.raw, section)
	if !sectionResult.Exists() {
		return fmt.Errorf("section %q not found", section)
	}

	found := false

	sectionResult.ForEach(func(key, value gjson.Result) bool {
		if key.String() != name {
			return true
		}

		found = true

		encoded, err := json.Marshal(newValue)
		if err != nil {
			return false
		}

		start := value.Index
		end := value.Index + len(value.Raw)
		replacement := append([]byte(encoded), p.raw[end:]...)
		p.raw = append(p.raw[:start], replacement...)

		return false
	})

	if !found {
		return fmt.Errorf("dependency %q not found in section %q", name, section)
	}

	return nil
}

func (p *PackageFile) Write(path string) error {
	return os.WriteFile(path, p.raw, 0o644)
}
