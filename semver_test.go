package main

import (
	"testing"
)

func TestParseSemVer_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"0.0.0", "0.0.0"},
		{"1.2.3", "1.2.3"},
		{"0.0.4", "0.0.4"},
		{"10.20.30", "10.20.30"},
		{"99999999999.99999999999.99999999999", "99999999999.99999999999.99999999999"},
		// Pre-release
		{"1.0.0-alpha", "1.0.0-alpha"},
		{"1.0.0-alpha.1", "1.0.0-alpha.1"},
		{"1.0.0-0.3.7", "1.0.0-0.3.7"},
		{"1.0.0-x.7.z.92", "1.0.0-x.7.z.92"},
		{"1.0.0-x-y-z.--", "1.0.0-x-y-z.--"},
		{"1.0.0-0", "1.0.0-0"},
		// Build
		{"1.0.0+001", "1.0.0+001"},
		{"1.0.0+20130313144700", "1.0.0+20130313144700"},
		{"1.0.0+21AF26D3----117B344092BD", "1.0.0+21AF26D3----117B344092BD"},
		// Pre-release + build
		{"1.0.0-alpha+001", "1.0.0-alpha+001"},
		{"1.0.0-beta+exp.sha.5114f85", "1.0.0-beta+exp.sha.5114f85"},
		// Edge: build with leading zeros (allowed)
		{"1.0.0+01", "1.0.0+01"},
		{"1.0.0-alpha+01.02.03", "1.0.0-alpha+01.02.03"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v, end, ok := parseSemVer([]byte(tt.input), 0)
			if !ok {
				t.Fatalf("parseSemVer(%q) failed", tt.input)
			}
			if end != len(tt.input) {
				t.Errorf("parseSemVer(%q) consumed %d of %d bytes", tt.input, end, len(tt.input))
			}
			if got := v.String(); got != tt.want {
				t.Errorf("parseSemVer(%q).String() = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseSemVer_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"leading zero major", "01.2.3"},
		{"leading zero minor", "1.02.3"},
		{"leading zero patch", "1.2.03"},
		{"missing patch", "1.2"},
		{"missing minor+patch", "1"},
		{"empty", ""},
		{"letters only", "abc"},
		{"pre-release leading zero numeric", "1.0.0-01"},
		{"pre-release empty ident", "1.0.0-"},
		{"pre-release double dot", "1.0.0-alpha..1"},
		{"build empty ident", "1.0.0+"},
		{"build double dot", "1.0.0+a..b"},
		{"four components", "1.2.3.4"},
		{"negative", "-1.2.3"},
		{"space in version", "1. 2.3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, end, ok := parseSemVer([]byte(tt.input), 0)
			// For some invalid inputs, the parser may still parse a valid prefix
			// (e.g., "1.2.3.4" parses "1.2.3" successfully). We check that it
			// doesn't consume the entire invalid string as-is.
			if ok && end == len(tt.input) {
				// Parsed the whole string — only valid if the parsed version
				// has a different representation (meaning we matched a prefix)
				t.Errorf("parseSemVer(%q) unexpectedly consumed entire input", tt.input)
			}
		})
	}
}

func TestFindSemVer(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		ok    bool
	}{
		{"bare version", "1.2.3", "1.2.3", true},
		{"v prefix", "v1.2.3", "1.2.3", true},
		{"leading text", "version: 1.2.3", "1.2.3", true},
		{"trailing text", "1.2.3-beta+build foo", "1.2.3-beta+build", true},
		{"surrounded", "foo1.2.3bar", "1.2.3", true},
		{"no version", "hello world", "", false},
		{"empty", "", "", false},
		{"with pre-release", "tag: 2.0.0-rc.1", "2.0.0-rc.1", true},
		{"with build", "ver 3.1.4+build.42", "3.1.4+build.42", true},
		{"invalid pre-release consumed as core", "1.0.0-01", "1.0.0", true},
		{"multiple digits before version", "abc123 4.5.6", "4.5.6", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, ok := MakeSemVer([]byte(tt.input))
			if ok != tt.ok {
				t.Fatalf("FindSemVer(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if ok {
				if got := v.String(); got != tt.want {
					t.Errorf("FindSemVer(%q) = %q, want %q", tt.input, got, tt.want)
				}
			}
		})
	}
}

func TestBump(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		major      int
		minor      int
		patch      int
		prerelease bool
		build      bool
		want       string
	}{
		{"patch bump", "1.2.3", 0, 0, 1, false, false, "1.2.4"},
		{"minor bump", "1.2.3", 0, 1, 0, false, false, "1.3.0"},
		{"major bump", "1.2.3", 1, 0, 0, false, false, "2.0.0"},
		{"double major", "1.2.3", 2, 0, 0, false, false, "3.0.0"},
		{"major + minor", "1.2.3", 1, 1, 0, false, false, "2.1.0"},
		{"major + minor + patch", "1.2.3", 1, 1, 1, false, false, "2.1.1"},
		{"minor + patch", "1.2.3", 0, 1, 1, false, false, "1.3.1"},
		{"from zero", "0.0.0", 0, 0, 1, false, false, "0.0.1"},
		// Pre-release stripping
		{"strip pre-release on bump", "1.0.0-alpha", 0, 0, 1, false, false, "1.0.1"},
		{"retain pre-release on bump", "1.0.0-alpha", 0, 0, 1, true, false, "1.0.1-alpha"},
		// Build stripping
		{"strip build on bump", "1.0.0+build", 0, 0, 1, false, false, "1.0.1"},
		{"retain build on bump", "1.0.0+build", 0, 0, 1, false, true, "1.0.1+build"},
		// Both
		{"retain both", "1.0.0-beta+exp", 1, 0, 0, true, true, "2.0.0-beta+exp"},
		{"strip both", "1.0.0-beta+exp", 1, 0, 0, false, false, "2.0.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, ok := MakeSemVer([]byte(tt.input))
			if !ok {
				t.Fatalf("FindSemVer(%q) failed", tt.input)
			}
			for range tt.major {
				v.BumpMajor()
			}
			for range tt.minor {
				v.BumpMinor()
			}
			for range tt.patch {
				v.BumpPatch()
			}
			if !tt.prerelease {
				v.StripPreRelease()
			}
			if !tt.build {
				v.StripBuild()
			}
			if got := v.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNumericIdentifier_NoLeadingZero(t *testing.T) {
	// "01" should fail as a numeric identifier
	_, n, ok := parseNumericIdentifier([]byte("01"), 0)
	if ok && n == 2 {
		t.Error("parseNumericIdentifier(\"01\") should not consume both digits")
	}
}

func TestPreReleaseIdentifier_NoLeadingZero(t *testing.T) {
	// "01" as purely numeric pre-release identifier is invalid
	_, _, ok := parsePreReleaseIdentifier([]byte("01"), 0)
	if ok {
		t.Error("parsePreReleaseIdentifier(\"01\") should fail")
	}
}

func TestBuildIdentifier_LeadingZeroAllowed(t *testing.T) {
	s, n, ok := parseBuildIdentifier([]byte("01"), 0)
	if !ok || n != 2 || s != "01" {
		t.Errorf("parseBuildIdentifier(\"01\") = (%q, %d, %v), want (\"01\", 2, true)", s, n, ok)
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		v    SemVer
		want string
	}{
		{SemVer{1, 2, 3, nil, nil, ""}, "1.2.3"},
		{SemVer{1, 0, 0, []string{"alpha"}, nil, ""}, "1.0.0-alpha"},
		{SemVer{1, 0, 0, nil, []string{"001"}, ""}, "1.0.0+001"},
		{SemVer{1, 0, 0, []string{"beta", "1"}, []string{"exp", "sha", "5114f85"}, ""}, "1.0.0-beta.1+exp.sha.5114f85"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.v.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
