package main

import "testing"

func TestReleasePattern(t *testing.T) {
	tests := map[string]bool{
		"0.6.0":                               true,
		"1.12.3":                              true,
		"dev":                                 false,
		"0.6.0+dirty":                         false,
		"0.6.1-0.20261105120000-abcdef123456": false,
		"0.7.0-rc.1":                          false,
		"":                                    false,
	}
	for v, want := range tests {
		if got := releasePattern.MatchString(v); got != want {
			t.Errorf("releasePattern(%q) = %v, want %v", v, got, want)
		}
	}
}
