package util

import "testing"

func TestNormalizeMobile(t *testing.T) {
	cases := map[string]string{
		"+91 98765 43210": "9876543210",
		"098765-43210":    "9876543210",
		"919876543210":    "9876543210",
		"9876543210":      "9876543210",
		" (98765) 43210 ": "9876543210",
		"abc":             "abc",
		"-":               "",
	}
	for in, want := range cases {
		if got := NormalizeMobile(in); got != want {
			t.Errorf("NormalizeMobile(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeEmail(t *testing.T) {
	if got := NormalizeEmail("  Abc@X.com "); got != "abc@x.com" {
		t.Errorf("got %q", got)
	}
}
