package cautils

import "testing"

func TestRemoveIndexFromStringList(t *testing.T) {
	s := []string{"a", "b", "c"}
	RemoveIndexFromStringSlice(&s, 1)
	if len(s) != 2 || s[0] != "a" || s[1] != "c" {
		t.Errorf("did not delete from slice")
	}
}

func TestMergeSliceAndMap(t *testing.T) {
	s := []string{"a", "b", "c"}
	m := map[string]string{"a": "A", "c": "C", "d": "D"}
	f := MergeSliceAndMap(s, m)
	if len(f) != 2 || f["a"] != "A" || f["c"] != "C" {
		t.Errorf("did not MergeSliceAndMap")
	}
}
