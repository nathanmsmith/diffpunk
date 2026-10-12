package private

// PatchArgs inserts format controls before the path separator, leaving all Git selection
// options, revisions, and pathspecs intact.
func PatchArgs(args []string, full bool) []string {
	controls := []string{"--no-color", "--src-prefix=a/", "--dst-prefix=b/"}
	if full {
		controls = append(controls, "--unified=2147483647")
	}
	return diffArgs(args, controls)
}
