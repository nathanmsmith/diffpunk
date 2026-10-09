package main

import (
	"path"
	"sort"
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

func testDirectory(s string) bool {
	switch s {
	case "test", "tests", "spec", "specs", "__tests__":
		return true
	}
	return false
}

func normalizedDir(dir string) string {
	var parts []string
	original := strings.Split(strings.TrimSuffix(dir, "/"), "/")
	for i, part := range original {
		// Maven/Gradle mirror src/main and src/test.
		if part == "main" && i > 0 && original[i-1] == "src" {
			continue
		}
		if part != "" && !testDirectory(part) && part != "src" && part != "lib" && part != "app" {
			parts = append(parts, part)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "/") + "/"
}

func sortFiles(files []File) []File {
	less := func(a, b File) bool {
		da, db := strings.Count(a.Name, "/"), strings.Count(b.Name, "/")
		if da != db {
			return da < db
		}
		return a.Name < b.Name
	}
	var implementations, tests []File
	for _, f := range files {
		_, test := testIdentity(f.Name)
		if test {
			tests = append(tests, f)
		} else {
			implementations = append(implementations, f)
		}
	}
	sort.Slice(implementations, func(i, j int) bool { return less(implementations[i], implementations[j]) })
	sort.Slice(tests, func(i, j int) bool { return less(tests[i], tests[j]) })
	owners := make(map[string][]int)
	for i, f := range implementations {
		dir, base := path.Split(f.Name)
		owners[normalizedDir(dir)+base] = append(owners[normalizedDir(dir)+base], i)
	}
	paired := make(map[int][]File)
	var unpaired []File
	for _, test := range tests {
		key, _ := testIdentity(test.Name)
		candidates := owners[key]
		owner := -1
		if len(candidates) == 1 {
			owner = candidates[0]
		} else {
			// An adjacent test is unambiguous even if another module normalizes to
			// the same path; otherwise keep the test unpaired.
			for _, i := range candidates {
				if path.Dir(test.Name) == path.Dir(implementations[i].Name) {
					if owner != -1 {
						owner = -1
						break
					}
					owner = i
				}
			}
		}
		if owner >= 0 {
			paired[owner] = append(paired[owner], test)
		} else {
			unpaired = append(unpaired, test)
		}
	}
	result := make([]File, 0, len(files))
	for i, f := range implementations {
		result = append(result, f)
		result = append(result, paired[i]...)
	}
	return append(result, unpaired...)
}
