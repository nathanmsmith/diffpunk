package private

import (
	"bytes"
	"context"
	"strings"
)

// Pure renames and permission changes have no patch hunks. Their original
// blobs still let the user inspect the whole unchanged text file.
func captureMetadataContents(ctx context.Context, args []string, files []File) {
	wanted := map[string]int{}
	for i, f := range files {
		if !strings.Contains(f.Patch, "\n@@ ") && !strings.Contains(f.Patch, "Binary files ") && !strings.Contains(f.Patch, "GIT binary patch") {
			wanted[f.Name] = i
		}
	}
	if len(wanted) == 0 {
		return
	}
	raw, err := Command(ctx, diffArgs(args, []string{"--no-color", "--no-patch", "--raw", "-z", "--no-abbrev"})...).Output()
	if err != nil {
		return
	}
	records := strings.Split(string(raw), "\x00")
	for i := 0; i+1 < len(records); {
		fields := strings.Fields(records[i])
		if len(fields) != 5 || !strings.HasPrefix(fields[0], ":") {
			break
		}
		name := records[i+1]
		i += 2
		if fields[4][0] == 'R' || fields[4][0] == 'C' {
			if i >= len(records) {
				break
			}
			name = records[i]
			i++
		}
		index, ok := wanted[name]
		if !ok || fields[0] == ":160000" || fields[1] == "160000" {
			continue
		}
		object := fields[2]
		if strings.Trim(object, "0") == "" {
			object = fields[3]
		}
		var content []byte
		if strings.Trim(object, "0") != "" {
			content, err = Command(ctx, "cat-file", "blob", object).Output()
			if err != nil || bytes.IndexByte(content, 0) >= 0 {
				continue
			}
		}
		text := string(content)
		files[index].FullContent = &text
	}
}
