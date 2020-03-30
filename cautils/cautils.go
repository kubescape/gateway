package cautils

// RemoveIndexFromStringSlice -
func RemoveIndexFromStringSlice(s *[]string, index int) {
	(*s)[index] = (*s)[len(*s)-1]
	(*s)[len(*s)-1] = ""
	*s = (*s)[:len(*s)-1]
}

// MergeSliceAndMap merge a list and keys of map
func MergeSliceAndMap(s []string, m map[string]string) map[string]string {
	merged := make(map[string]string)
	for i := range s {
		if v, ok := m[s[i]]; ok {
			merged[s[i]] = v
		}
	}
	return merged
}
