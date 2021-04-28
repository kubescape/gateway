package cautils

import "encoding/json"

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

// ObjectToString Convert an object to a string
func ObjectToString(obj interface{}) string {
	bm, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(bm)
}
