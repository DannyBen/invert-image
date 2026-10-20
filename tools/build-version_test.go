package main

import "testing"

func TestNextPatch(t *testing.T) {
	tests := []struct {
		version string
		want    string
		wantOK  bool
	}{
		{version: "0.1.0", want: "0.1.1", wantOK: true},
		{version: "12.34.56", want: "12.34.57", wantOK: true},
		{version: "1.2", wantOK: false},
		{version: "1.2.beta", wantOK: false},
	}
	for _, test := range tests {
		got, ok := nextPatch(test.version)
		if got != test.want || ok != test.wantOK {
			t.Errorf("nextPatch(%q) = (%q, %t), want (%q, %t)", test.version, got, ok, test.want, test.wantOK)
		}
	}
}

func TestIsVersionTag(t *testing.T) {
	for _, tag := range []string{"v0.1.0", "v12.34.56"} {
		if !isVersionTag(tag) {
			t.Errorf("isVersionTag(%q) = false, want true", tag)
		}
	}
	for _, tag := range []string{"0.1.0", "v1.2", "latest", "v1.2.3-rc1"} {
		if isVersionTag(tag) {
			t.Errorf("isVersionTag(%q) = true, want false", tag)
		}
	}
}
