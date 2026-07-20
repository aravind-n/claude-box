package vhrn

import "testing"

func TestHistoryKey(t *testing.T) {
	cases := map[string]string{
		"/Users/aravind/projects/vhrn": "-Users-aravind-projects-vhrn",
		"/a/b_c.d":                     "-a-b-c-d",
		"/x/y-z":                       "-x-y-z",
	}
	for in, want := range cases {
		if got := historyKey(in); got != want {
			t.Errorf("historyKey(%q) = %q, want %q", in, got, want)
		}
	}
}
