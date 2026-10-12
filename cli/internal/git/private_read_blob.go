package git

import (
	"context"
	"strings"
)

func readBlob(ctx context.Context, id string) (string, bool) {
	if strings.Trim(id, "0") == "" {
		return "", false
	}
	data, err := Command(ctx, "cat-file", "blob", id).Output()
	return string(data), err == nil && !strings.ContainsRune(string(data), 0)
}
