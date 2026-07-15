package upd

import (
	"fmt"
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
)

func TestErrorFamilyClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want errorfamily.Family
		exit int
	}{
		{"ErrFileNotFound is Rejection", ErrFileNotFound, errorfamily.Rejection, 1},
		{"ErrInvalidJSON is Corruption", ErrInvalidJSON, errorfamily.Corruption, 65},
		{"ErrPackageNotFound is Rejection", ErrPackageNotFound, errorfamily.Rejection, 1},
		{"ErrRegistryUnavailable is Transient", ErrRegistryUnavailable, errorfamily.Transient, 75},
		{"ErrVersionParse is Corruption", ErrVersionParse, errorfamily.Corruption, 65},
		{"ErrNoLatestDistTag is Corruption", ErrNoLatestDistTag, errorfamily.Corruption, 65},
		{"ErrNoValidVersions is Corruption", ErrNoValidVersions, errorfamily.Corruption, 65},
		{"ErrNoSemverVersions is Corruption", ErrNoSemverVersions, errorfamily.Corruption, 65},
		{"ErrSectionNotFound is Rejection", ErrSectionNotFound, errorfamily.Rejection, 1},
		{"ErrSectionNotObject is Corruption", ErrSectionNotObject, errorfamily.Corruption, 65},
		{"ErrDependencyNotFound is Rejection", ErrDependencyNotFound, errorfamily.Rejection, 1},
		{"ErrConcurrentModification is Conflict", ErrConcurrentModification, errorfamily.Conflict, 1},
		{"ErrPartialFailure is Rejection", ErrPartialFailure, errorfamily.Rejection, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := errorfamily.Classify(tt.err)
			if got != tt.want {
				t.Errorf("Classify = %s, want %s", got, tt.want)
			}

			exit := got.ExitCode()
			if exit != tt.exit {
				t.Errorf("ExitCode = %d, want %d", exit, tt.exit)
			}
		})
	}
}

func TestErrorFamilyWrappedChainClassification(t *testing.T) {
	t.Parallel()

	t.Run("wrapped Transient preserves family through chain", func(t *testing.T) {
		t.Parallel()

		wrapped := fmt.Errorf("fetch failed: %w", ErrRegistryUnavailable)
		got := errorfamily.Classify(wrapped)

		if got != errorfamily.Transient {
			t.Errorf("Classify(wrapped) = %s, want Transient", got)
		}

		if exit := got.ExitCode(); exit != 75 {
			t.Errorf("ExitCode = %d, want 75", exit)
		}
	})

	t.Run("wrapped Rejection preserves family through chain", func(t *testing.T) {
		t.Parallel()

		wrapped := fmt.Errorf("2 packages failed: %w", ErrPartialFailure)
		got := errorfamily.Classify(wrapped)

		if got != errorfamily.Rejection {
			t.Errorf("Classify(wrapped) = %s, want Rejection", got)
		}
	})

	t.Run("WithContext preserves family", func(t *testing.T) {
		t.Parallel()

		enriched := ErrPackageNotFound.WithContext("package", "ghost")
		got := errorfamily.Classify(enriched)

		if got != errorfamily.Rejection {
			t.Errorf("Classify(enriched) = %s, want Rejection", got)
		}
	})
}
