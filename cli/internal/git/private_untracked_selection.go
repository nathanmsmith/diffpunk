package git

import (
	"context"
	"strings"
)

// untrackedSelection lets Git resolve revisions and pathspecs, keeping diff
// option operands out of revision resolution. Run has already validated args.
func untrackedSelection(ctx context.Context, args []string) (*UntrackedOptions, []string, string, error) {
	options := &UntrackedOptions{}
	var paths []string
	revisions := 0
	relative, _ := Command(ctx, "config", "--bool", "diff.relative").Output()
	relativeEnabled := strings.TrimSpace(string(relative)) == "true"
	relativePrefix := ""
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			paths = append(paths, args[i+1:]...)
			break
		}
		switch arg {
		case "--cached", "--staged":
			options.DefaultIncluded = true
		case "--relative":
			relativeEnabled = true
			relativePrefix = ""
		case "--no-relative":
			relativeEnabled = false
			relativePrefix = ""
		case "-U", "--unified", "--inter-hunk-context", "--diff-algorithm", "--anchored", "-S", "-G", "-I", "--ignore-matching-lines", "--find-object", "-O", "--order-file", "--skip-to", "--rotate-to", "--diff-filter", "-l":
			i++
		default:
			if strings.HasPrefix(arg, "--relative=") {
				relativeEnabled = true
				relativePrefix = strings.TrimPrefix(arg, "--relative=")
			} else if !strings.HasPrefix(arg, "-") {
				resolved, err := Command(ctx, "rev-parse", "--revs-only", "--no-flags", "--end-of-options", arg).Output()
				if err != nil {
					return nil, nil, "", err
				}
				if len(resolved) == 0 {
					paths = append(paths, arg)
				} else {
					revisions += len(strings.Fields(string(resolved)))
				}
			}
		}
	}
	if revisions > 1 {
		return nil, nil, "", nil
	}
	if relativeEnabled {
		if relativePrefix == "" {
			prefix, err := Command(ctx, "rev-parse", "--show-prefix").Output()
			if err != nil {
				return nil, nil, "", err
			}
			relativePrefix = strings.TrimSuffix(string(prefix), "\n")
		}
		if relativePrefix != "" {
			relativePrefix = strings.TrimSuffix(relativePrefix, "/") + "/"
		}
	} else {
		relativePrefix = ""
	}
	return options, paths, relativePrefix, nil
}

func pagerUntrackedOptions(files []File) *UntrackedOptions {
	var options *UntrackedOptions
	for _, file := range files {
		header, _, _ := strings.Cut(file.Patch, "\n")
		names := pagerFileHeader.FindStringSubmatch(header)
		if names == nil {
			return nil
		}
		prefixes := strings.TrimPrefix(names[1], `"`)[:1] + strings.TrimPrefix(names[2], `"`)[:1]
		staged := prefixes == "ci" || prefixes == "ic"
		if !staged && prefixes != "iw" && prefixes != "wi" && prefixes != "cw" && prefixes != "wc" {
			return nil
		}
		if options != nil && options.DefaultIncluded != staged {
			return nil
		}
		options = &UntrackedOptions{DefaultIncluded: staged}
	}
	return options
}
