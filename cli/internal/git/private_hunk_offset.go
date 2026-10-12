package git

func hunkOffset(start, count int) int {
	if count == 0 {
		return start
	}
	return start - 1
}
