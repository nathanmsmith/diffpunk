package server

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Only start tools from the checkout containing this development executable,
// never from the repository whose diff the user is viewing.
func devCheckout() string {
	executable, err := os.Executable()
	if err != nil {
		return ""
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return ""
	}
	root := filepath.Dir(executable)
	if filepath.Base(root) == "bin" {
		root = filepath.Dir(root)
	}
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return ""
	}
	var pkg struct {
		Name string `json:"name"`
	}
	if json.Unmarshal(data, &pkg) != nil || pkg.Name != "diff-punk" {
		return ""
	}
	for _, name := range []string{"vite.config.mjs", "frontend/src/main.tsx"} {
		if info, err := os.Stat(filepath.Join(root, name)); err != nil || info.IsDir() {
			return ""
		}
	}
	return root
}
