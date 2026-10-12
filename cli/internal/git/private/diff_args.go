package private

func diffArgs(args, controls []string) []string {
	result := []string{"-c", "diff.noprefix=false", "-c", "diff.mnemonicPrefix=false", "diff"}
	for i, arg := range args {
		if arg == "--" {
			result = append(result, controls...)
			return append(result, args[i:]...)
		}
		result = append(result, arg)
	}
	return append(result, controls...)
}
