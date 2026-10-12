package private

func testDirectory(s string) bool {
	switch s {
	case "test", "tests", "spec", "specs", "__tests__":
		return true
	}
	return false
}
