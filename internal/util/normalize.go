package util

import "strings"

// NormalizeEmail trims and lowercases an email so lookups and uniqueness
// aren't case-sensitive.
func NormalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// NormalizeMobile strips spaces, dashes and parentheses, then a +91 / 91 / 0
// country or trunk prefix, leaving the bare 10 digits for a valid number.
// Garbage is returned cleaned but otherwise unchanged, so the `mobile`
// validator still rejects it.
func NormalizeMobile(s string) string {
	var b strings.Builder
	for i, r := range strings.TrimSpace(s) {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '+' && i == 0:
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '(' || r == ')':
		default:
			b.WriteRune(r)
		}
	}
	out := b.String()
	switch {
	case strings.HasPrefix(out, "+91") && len(out) == 13:
		return out[3:]
	case strings.HasPrefix(out, "91") && len(out) == 12:
		return out[2:]
	case strings.HasPrefix(out, "0") && len(out) == 11:
		return out[1:]
	}
	return out
}
