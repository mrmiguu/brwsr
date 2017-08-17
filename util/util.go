package util

// Min calculates the minimum of the two values.
func Min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

// Max calculates the maximum of the two values.
func Max(i, j int) int {
	if i > j {
		return i
	}
	return j
}
