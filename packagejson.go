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
	errorfamily "github.com/larsartmann/go-error-family"
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
			return nil, ErrFileNotFound.WithContext("path", path)
		}

		return nil, errorfamily.WrapRejection(
			err,
			"file.read_failed",
			fmt.Sprintf("read package configuration file %q", path),
		)
	}

	if !jsontext.Value(data).IsValid() {
		return nil, ErrInvalidJSON.WithContext("path", path)
	}

	return &PackageFile{raw: data, fingerprint: atomicwrite.FingerprintFromBytes(data)}, nil
}

func (p *PackageFile) Raw() []byte {
	return p.raw
}

func (p *PackageFile) GetDependencySection(section string) (map[string]string, error) {
	var topLevel map[string]jsontext.Value

	err := json.Unmarshal(p.raw, &topLevel)
	if err != nil {
		return nil, errorfamily.WrapCorruption(
			err,
			"json.parse_toplevel",
			fmt.Sprintf("parse top-level JSON for section %q", section),
		)
	}

	sectionRaw, ok := topLevel[section]
	if !ok {
		return make(map[string]string), nil
	}

	if sectionRaw.Kind() != jsontext.KindBeginObject {
		return nil, ErrSectionNotObject.WithContext("section", section).WithContextf("kind", "%s", sectionRaw.Kind())
	}

	var deps map[string]string

	err = json.Unmarshal(sectionRaw, &deps)
	if err != nil {
		return nil, errorfamily.WrapCorruption(
			err,
			"json.parse_section",
			fmt.Sprintf("parse section %q: expected object of name→version strings", section),
		)
	}

	return deps, nil
}

func (p *PackageFile) GetUpdArgs() ([]string, error) {
	var v struct {
		Upd jsontext.Value `json:"upd"`
	}

	err := json.Unmarshal(p.raw, &v)
	if err != nil {
		return nil, errorfamily.WrapCorruption(err, "json.parse_upd_field", "parse upd field")
	}

	if !v.Upd.IsValid() {
		return nil, nil
	}

	kind := v.Upd.Kind()
	if kind == jsontext.KindBeginArray {
		return parseUpdArray(v.Upd)
	}

	if kind == jsontext.KindString {
		return parseUpdString(v.Upd)
	}

	return nil, nil
}

func parseUpdArray(raw jsontext.Value) ([]string, error) {
	var args []string

	err := json.Unmarshal(raw, &args)
	if err != nil {
		return nil, errorfamily.WrapCorruption(err, "json.parse_upd_array", "parse upd array")
	}

	return args, nil
}

func parseUpdString(raw jsontext.Value) ([]string, error) {
	var s string

	err := json.Unmarshal(raw, &s)
	if err != nil {
		return nil, errorfamily.WrapCorruption(err, "json.parse_upd_string", "parse upd string")
	}

	return strings.Fields(s), nil
}

func (p *PackageFile) UpdateDependency(section, name, newValue string) error {
	dec := jsontext.NewDecoder(bytes.NewReader(p.raw))

	err := p.navigateToSection(dec, section)
	if err != nil {
		return err
	}

	return p.replaceDependency(dec, name, newValue)
}

func (p *PackageFile) navigateToSection(dec *jsontext.Decoder, section string) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return errorfamily.WrapCorruption(err, "json.read_root", "parse root token")
	}

	if tok.Kind() != jsontext.KindBeginObject {
		return ErrInvalidJSON.WithContext("kind", tok.Kind().String())
	}

	for dec.PeekKind() != jsontext.KindEndObject {
		keyTok, err := dec.ReadToken()
		if err != nil {
			return errorfamily.WrapCorruption(err, "json.read_key", "read key token")
		}

		if keyTok.String() == section {
			return p.enterSection(dec, section)
		}

		err = dec.SkipValue()
		if err != nil {
			return errorfamily.WrapCorruption(err, "json.skip_value", "skip value")
		}
	}

	return ErrSectionNotFound.WithContext("section", section)
}

func (p *PackageFile) enterSection(dec *jsontext.Decoder, section string) error {
	objTok, err := dec.ReadToken()
	if err != nil {
		return errorfamily.WrapCorruption(err, "json.read_section_start", "read section start token")
	}

	if objTok.Kind() != jsontext.KindBeginObject {
		return ErrSectionNotObject.WithContext("section", section).WithContextf("kind", "%s", objTok.Kind())
	}

	return nil
}

func (p *PackageFile) replaceDependency(dec *jsontext.Decoder, name, newValue string) error {
	for dec.PeekKind() != jsontext.KindEndObject {
		keyTok, err := dec.ReadToken()
		if err != nil {
			return errorfamily.WrapCorruption(err, "json.read_dependency_key", "read dependency key")
		}

		keyStr := keyTok.String()

		val, err := dec.ReadValue()
		if err != nil {
			return errorfamily.WrapCorruption(err, "json.read_dependency_value", "read dependency value")
		}

		if keyStr == name {
			return p.spliceDependency(dec, val, newValue)
		}
	}

	return ErrDependencyNotFound.WithContext("dependency", name)
}

func (p *PackageFile) spliceDependency(dec *jsontext.Decoder, val jsontext.Value, newValue string) error {
	encoded, err := json.Marshal(newValue)
	if err != nil {
		return errorfamily.WrapCorruption(err, "json.encode_value", "encode new value")
	}

	endOffset := int(dec.InputOffset())
	startOffset := endOffset - len(val)

	newRaw := make([]byte, 0, len(p.raw)-(endOffset-startOffset)+len(encoded))
	newRaw = append(newRaw, p.raw[:startOffset]...)
	newRaw = append(newRaw, encoded...)
	newRaw = append(newRaw, p.raw[endOffset:]...)
	p.raw = newRaw

	return nil
}

func (p *PackageFile) Write(path string) error {
	err := atomicwrite.Write(path, p.raw, p.fingerprint)
	if err != nil {
		if errors.Is(err, atomicwrite.ErrConcurrentModification) {
			return ErrConcurrentModification.WithContext("path", path)
		}

		return errorfamily.WrapRejection(
			err,
			"file.write_failed",
			fmt.Sprintf("write package configuration file %q", path),
		)
	}

	return nil
}
