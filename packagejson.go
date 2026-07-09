package upd

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"strings"

	atomicwrite "github.com/larsartmann/go-atomic-write"
)

type PackageFile struct {
	raw         []byte
	fingerprint atomicwrite.Fingerprint
}

func ReadPackageFile(path string) (*PackageFile, error) {
	//nolint:gosec // path is supplied by the operator via -f/--file; not attacker-controlled.
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("cannot find NPM package configuration file %q: %w", path, ErrFileNotFound)
		}

		return nil, fmt.Errorf("failed to read package configuration file %q: %w", path, err)
	}

	if !jsontext.Value(data).IsValid() {
		return nil, fmt.Errorf("failed to parse package configuration file %q: %w", path, ErrInvalidJSON)
	}

	return &PackageFile{raw: data, fingerprint: atomicwrite.FingerprintFromBytes(data)}, nil
}

func (p *PackageFile) Raw() []byte {
	return p.raw
}

func (p *PackageFile) GetDependencySection(section string) map[string]string {
	var topLevel map[string]jsontext.Value

	err := json.Unmarshal(p.raw, &topLevel)
	if err != nil {
		return make(map[string]string)
	}

	sectionRaw, ok := topLevel[section]
	if !ok || sectionRaw.Kind() != jsontext.KindBeginObject {
		return make(map[string]string)
	}

	var deps map[string]string
	if err := json.Unmarshal(sectionRaw, &deps); err != nil {
		return make(map[string]string)
	}

	return deps
}

func (p *PackageFile) GetUpdArgs() []string {
	var v struct {
		Upd jsontext.Value `json:"upd"`
	}

	err := json.Unmarshal(p.raw, &v)
	if err != nil {
		return nil
	}

	if !v.Upd.IsValid() {
		return nil
	}

	switch v.Upd.Kind() {
	case jsontext.KindBeginArray:
		var args []string

		err := json.Unmarshal(v.Upd, &args)
		if err != nil {
			return nil
		}

		return args
	case jsontext.KindString:
		var s string

		err := json.Unmarshal(v.Upd, &s)
		if err != nil {
			return nil
		}

		return strings.Fields(s)
	default:
		return nil
	}
}

func (p *PackageFile) UpdateDependency(section, name, newValue string) error {
	dec := jsontext.NewDecoder(bytes.NewReader(p.raw))

	tok, err := dec.ReadToken()
	if err != nil {
		return fmt.Errorf("parse root: %w", err)
	}

	if tok.Kind() != jsontext.KindBeginObject {
		return fmt.Errorf("expected root object: %w", ErrInvalidJSON)
	}

	sectionFound := false

	for dec.PeekKind() != jsontext.KindEndObject {
		keyTok, err := dec.ReadToken()
		if err != nil {
			return fmt.Errorf("read key: %w", err)
		}

		if keyTok.String() == section {
			sectionFound = true

			break
		}

		if err := dec.SkipValue(); err != nil {
			return fmt.Errorf("skip value: %w", err)
		}
	}

	if !sectionFound {
		return fmt.Errorf("%w: %q", ErrSectionNotFound, section)
	}

	objTok, err := dec.ReadToken()
	if err != nil {
		return fmt.Errorf("read section start: %w", err)
	}

	if objTok.Kind() != jsontext.KindBeginObject {
		return fmt.Errorf("section %q is not an object", section)
	}

	found := false

	for dec.PeekKind() != jsontext.KindEndObject {
		keyTok, err := dec.ReadToken()
		if err != nil {
			return fmt.Errorf("read dependency key: %w", err)
		}

		keyStr := keyTok.String()

		val, err := dec.ReadValue()
		if err != nil {
			return fmt.Errorf("read dependency value: %w", err)
		}

		if keyStr == name {
			encoded, err := json.Marshal(newValue)
			if err != nil {
				return fmt.Errorf("encode new value: %w", err)
			}

			endOffset := int(dec.InputOffset())
			startOffset := endOffset - len(val)

			newRaw := make([]byte, 0, len(p.raw)-(endOffset-startOffset)+len(encoded))
			newRaw = append(newRaw, p.raw[:startOffset]...)
			newRaw = append(newRaw, encoded...)
			newRaw = append(newRaw, p.raw[endOffset:]...)
			p.raw = newRaw

			found = true

			break
		}
	}

	if !found {
		return fmt.Errorf("%w: %q in %q", ErrDependencyNotFound, name, section)
	}

	return nil
}

func (p *PackageFile) Write(path string) error {
	err := atomicwrite.Write(path, p.raw, p.fingerprint)
	if err != nil {
		if errors.Is(err, atomicwrite.ErrConcurrentModification) {
			return fmt.Errorf("%w: %q", ErrConcurrentModification, path)
		}

		return fmt.Errorf("write package configuration file %q: %w", path, err)
	}

	return nil
}
