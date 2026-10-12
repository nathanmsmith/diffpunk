package git

import "regexp"

var terminalColor = regexp.MustCompile(`\x1b\[[0-9;:]*m`)

// parsePagerFiles strips terminal colors and validates supplied Git patches.
func parsePagerFiles(patch string) ([]File, error) {
	files, err := parseFiles(terminalColor.ReplaceAllString(patch, ""))
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if _, _, err := parseHunks(file.Patch); err != nil {
			return nil, err
		}
	}
	return files, nil
}
