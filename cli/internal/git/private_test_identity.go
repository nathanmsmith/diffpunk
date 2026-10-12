package git

import (
	"path"
	"strings"
)

// testIdentity recognizes filename and directory conventions without requiring
// configuration. Matching also considers mirrored source/test paths, avoiding
// accidental pairing of identically named files in different modules.
func testIdentity(name string) (string, bool) {
	dir, base := path.Split(name)
	ext := path.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	isTest := false
	for _, suffix := range []string{".test", ".spec", "_test", "_spec", "Test", "Tests", "Spec"} {
		if strings.HasSuffix(stem, suffix) {
			stem = strings.TrimSuffix(stem, suffix)
			isTest = true
			break
		}
	}
	if strings.HasPrefix(stem, "test_") {
		stem = strings.TrimPrefix(stem, "test_")
		isTest = true
	}
	parts := strings.Split(strings.TrimSuffix(dir, "/"), "/")
	for _, part := range parts {
		if testDirectory(part) {
			isTest = true
		}
	}
	return normalizedDir(dir) + stem + ext, isTest
}
