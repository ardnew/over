package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// [SemVer] represents a parsed semantic version 2.0.0 string.
//
// [SemVer]: https://semver.org/
type SemVer struct {
	Major      uint64   `json:"major"`
	Minor      uint64   `json:"minor"`
	Patch      uint64   `json:"patch"`
	PreRelease []string `json:"prerelease,omitempty,omitzero"` // nil means absent, empty slice not possible per BNF
	Build      []string `json:"build,omitempty,omitzero"`      // nil means absent

	indent string `json:"-"`
}

// MakeSemVer scans the input for the first valid semver substring.
// Returns the parsed SemVer and true if found, or zero value and false.
func MakeSemVer(input []byte, opts ...func(*SemVer)) (SemVer, bool) {
	for i := range input {
		if isDigit(input[i]) {
			v, _, ok := parseSemVer(input, i)
			if ok {
				for _, opt := range opts {
					opt(&v)
				}
				return v, true
			}
		}
	}
	return SemVer{}, false
}

// WithJSONIndent returns an option to set the indentation for pretty JSON.
func WithJSONIndent(indent string) func(*SemVer) {
	return func(v *SemVer) { v.indent = indent }
}

// BumpMajor increments major, resets minor and patch to 0.
func (v *SemVer) BumpMajor() { v.Major++; v.Minor = 0; v.Patch = 0 }

// BumpMinor increments minor, resets patch to 0.
func (v *SemVer) BumpMinor() { v.Minor++; v.Patch = 0 }

// BumpPatch increments patch.
func (v *SemVer) BumpPatch() { v.Patch++ }

// StripPreRelease removes the pre-release component.
func (v *SemVer) StripPreRelease() { v.PreRelease = nil }

// StripBuild removes the build metadata component.
func (v *SemVer) StripBuild() { v.Build = nil }

// String returns the canonical semver string representation.
func (v SemVer) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != nil {
		b.WriteByte('-')
		for i, id := range v.PreRelease {
			if i > 0 {
				b.WriteByte('.')
			}
			b.WriteString(id)
		}
	}
	if v.Build != nil {
		b.WriteByte('+')
		for i, id := range v.Build {
			if i > 0 {
				b.WriteByte('.')
			}
			b.WriteString(id)
		}
	}
	return b.String()
}

// JSON returns the JSON string representation of the SemVer.
func (v SemVer) JSON() string {
	b, _ := json.Marshal(v)
	return string(b)
}

// PrettyJSON returns the pretty JSON string representation of the SemVer.
func (v SemVer) PrettyJSON() string {
	b, _ := json.MarshalIndent(v, "", v.indent)
	return string(b)
}

// --- BNF Parser ---
// All parse functions operate on a byte slice and return the advance count.
// A return of 0 (or ok=false) means no match.

// isDigit returns true for ASCII digits 0-9.
func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// isPositiveDigit returns true for ASCII digits 1-9.
func isPositiveDigit(c byte) bool { return c >= '1' && c <= '9' }

// isLetter returns true for ASCII letters A-Za-z.
func isLetter(c byte) bool { return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') }

// isNonDigit returns true for letters and hyphen.
func isNonDigit(c byte) bool { return isLetter(c) || c == '-' }

// isIdentChar returns true for digits, letters, and hyphen.
func isIdentChar(c byte) bool { return isDigit(c) || isNonDigit(c) }

// parseNumericIdentifier parses: "0" | <positive digit> | <positive digit> <digits>
// Returns the parsed uint64 value, number of bytes consumed, and success.
func parseNumericIdentifier(data []byte, pos int) (uint64, int, bool) {
	if pos >= len(data) {
		return 0, 0, false
	}
	if data[pos] == '0' {
		// "0" alone — must not be followed by another digit (no leading zeros)
		if pos+1 < len(data) && isDigit(data[pos+1]) {
			return 0, 0, false
		}
		return 0, 1, true
	}
	if !isPositiveDigit(data[pos]) {
		return 0, 0, false
	}
	end := pos + 1
	for end < len(data) && isDigit(data[end]) {
		end++
	}
	val, err := strconv.ParseUint(string(data[pos:end]), 10, 64)
	if err != nil {
		return 0, 0, false
	}
	return val, end - pos, true
}

// parsePreReleaseIdentifier parses a single pre-release identifier.
// It's either a numeric identifier (no leading zeros) or an alphanumeric identifier
// (must contain at least one non-digit).
func parsePreReleaseIdentifier(data []byte, pos int) (string, int, bool) {
	if pos >= len(data) || !isIdentChar(data[pos]) {
		return "", 0, false
	}
	end := pos
	hasNonDigit := false
	for end < len(data) && isIdentChar(data[end]) {
		if isNonDigit(data[end]) {
			hasNonDigit = true
		}
		end++
	}
	s := string(data[pos:end])
	if !hasNonDigit {
		// Purely numeric — must not have leading zeros
		if len(s) > 1 && s[0] == '0' {
			return "", 0, false
		}
	}
	return s, end - pos, true
}

// parseBuildIdentifier parses a single build identifier.
// Alphanumeric or digits (leading zeros allowed).
func parseBuildIdentifier(data []byte, pos int) (string, int, bool) {
	if pos >= len(data) || !isIdentChar(data[pos]) {
		return "", 0, false
	}
	end := pos
	for end < len(data) && isIdentChar(data[end]) {
		end++
	}
	return string(data[pos:end]), end - pos, true
}

// parseDotSeparated parses a dot-separated list using the given identifier parser.
func parseDotSeparated(data []byte, pos int, parseFn func([]byte, int) (string, int, bool)) ([]string, int, bool) {
	id, n, ok := parseFn(data, pos)
	if !ok {
		return nil, 0, false
	}
	ids := []string{id}
	total := n
	for {
		next := pos + total
		if next >= len(data) || data[next] != '.' {
			break
		}
		id, n, ok = parseFn(data, next+1)
		if !ok {
			break
		}
		ids = append(ids, id)
		total += 1 + n // dot + identifier
	}
	return ids, total, true
}

// parseSemVer attempts to parse a full semver starting at pos.
// Returns the parsed SemVer, the end position (exclusive), and success.
func parseSemVer(data []byte, pos int) (SemVer, int, bool) {
	var v SemVer
	var n int
	var ok bool

	// Major
	v.Major, n, ok = parseNumericIdentifier(data, pos)
	if !ok {
		return v, 0, false
	}
	cur := pos + n

	// "."
	if cur >= len(data) || data[cur] != '.' {
		return v, 0, false
	}
	cur++

	// Minor
	v.Minor, n, ok = parseNumericIdentifier(data, cur)
	if !ok {
		return v, 0, false
	}
	cur += n

	// "."
	if cur >= len(data) || data[cur] != '.' {
		return v, 0, false
	}
	cur++

	// Patch
	v.Patch, n, ok = parseNumericIdentifier(data, cur)
	if !ok {
		return v, 0, false
	}
	cur += n

	// Optional pre-release: "-" <pre-release>
	if cur < len(data) && data[cur] == '-' {
		ids, pn, pok := parseDotSeparated(data, cur+1, parsePreReleaseIdentifier)
		if pok {
			v.PreRelease = ids
			cur += 1 + pn
		}
		// If "-" is present but no valid pre-release follows, we don't consume it.
		// The version core is still valid; the "-" is trailing junk.
	}

	// Optional build: "+" <build>
	if cur < len(data) && data[cur] == '+' {
		ids, bn, bok := parseDotSeparated(data, cur+1, parseBuildIdentifier)
		if bok {
			v.Build = ids
			cur += 1 + bn
		}
	}

	return v, cur, true
}
