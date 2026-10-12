package private

import (
	"context"
	"strings"
)

func matchesBlob(ctx context.Context, content, id string) bool {
	if strings.Trim(id, "0") == "" {
		return true // Git sometimes emits zero for an unstored working-tree side.
	}
	cmd := Command(ctx, "hash-object", "--stdin")
	cmd.Stdin = strings.NewReader(content)
	hash, err := cmd.Output()
	return err == nil && strings.HasPrefix(strings.TrimSpace(string(hash)), id)
}
