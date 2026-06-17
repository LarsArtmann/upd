package upd

import (
	"strings"
	"testing"
)

func assertContains(t *testing.T, haystack, needle, label string) {
	t.Helper()

	if !strings.Contains(haystack, needle) {
		t.Errorf("%s: expected %q in:\n%s", label, needle, haystack)
	}
}

func assertNotContains(t *testing.T, haystack, needle, label string) {
	t.Helper()

	if strings.Contains(haystack, needle) {
		t.Errorf("%s: %q should not be present in:\n%s", label, needle, haystack)
	}
}
