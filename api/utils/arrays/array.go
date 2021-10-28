package arrays

func StringSliceContains(a []string, s string) bool {
	for _, e := range a {
		if e == s {
			return true
		}
	}
	return false
}
