package git

import (
	"strings"
)

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
