package git

import (
	"fmt"
	"strings"
)

// parseFiles splits a Git patch into file changes.
func parseFiles(patch string) ([]File, error) {
	starts := fileStart.FindAllStringIndex(patch, -1)
	if len(starts) == 0 || starts[0][0] != 0 || strings.Contains(patch, "\ndiff --cc ") || strings.Contains(patch, "\ndiff --combined ") {
		return nil, fmt.Errorf("unsupported patch format")
	}
	files := make([]File, 0, len(starts))
	for i, start := range starts {
		end := len(patch)
		if i+1 < len(starts) {
			end = starts[i+1][0]
		}
		raw := patch[start[0]:end]
		lines := strings.Split(raw, "\n")
		oldName, name, err := headerNames(strings.TrimPrefix(lines[0], "diff --git "))
		if err != nil {
			return nil, err
		}
		f := File{Name: name, PreviousName: oldName, Kind: "modified", Patch: raw}
		for _, line := range lines[1:] {
			if strings.HasPrefix(line, "@@") {
				break
			}
			switch {
			case strings.HasPrefix(line, "--- "):
				if n := patchName(line[4:], true); n != "/dev/null" {
					f.PreviousName = n
				}
			case strings.HasPrefix(line, "+++ "):
				if n := patchName(line[4:], true); n != "/dev/null" {
					f.Name = n
				}
			case strings.HasPrefix(line, "new file mode "):
				f.Kind = "added"
			case strings.HasPrefix(line, "deleted file mode "):
				f.Kind = "deleted"
			case strings.HasPrefix(line, "rename from "):
				f.Kind = "renamed"
				f.PreviousName = patchName(line[len("rename from "):], false)
			case strings.HasPrefix(line, "rename to "):
				f.Name = patchName(line[len("rename to "):], false)
			case strings.HasPrefix(line, "copy from "):
				f.Kind = "copied"
				f.PreviousName = patchName(line[len("copy from "):], false)
			case strings.HasPrefix(line, "copy to "):
				f.Name = patchName(line[len("copy to "):], false)
			}
		}
		if f.PreviousName == f.Name {
			f.PreviousName = ""
		}
		files = append(files, f)
	}
	return files, nil
}
