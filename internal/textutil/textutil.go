// Package textutil contains small text helpers: slugs, excerpts, reading time.
package textutil

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// Slugify turns "Hello, Wörld!" into "hello-world".
func Slugify(s string) string {
	s = norm.NFKD.String(strings.ToLower(s))
	var b strings.Builder
	dash := false
	for _, r := range s {
		switch {
		case unicode.Is(unicode.Mn, r): // drop combining marks (accents)
			continue
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			// Non-Latin letters (e.g. CJK) are kept; browsers handle them in URLs.
			b.WriteRune(r)
			dash = false
		default:
			if !dash && b.Len() > 0 {
				b.WriteByte('-')
				dash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > 80 {
		out = strings.Trim(out[:80], "-")
		for !utf8.ValidString(out) { // don't cut a multi-byte rune in half
			out = out[:len(out)-1]
		}
	}
	return out
}

// Truncate shortens s to at most n runes, breaking on a word boundary and
// adding an ellipsis when it cuts.
func Truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	runes := []rune(s)
	cut := string(runes[:n])
	if i := strings.LastIndexAny(cut, " \n\t"); i > n/2 {
		cut = cut[:i]
	}
	return strings.TrimRight(cut, " ,.;:-") + "…"
}

// CollapseSpaces replaces runs of whitespace with a single space.
func CollapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// ReadingMinutes estimates reading time at ~220 words per minute (min 1).
func ReadingMinutes(plain string) int {
	words := len(strings.Fields(plain))
	m := (words + 219) / 220
	if m < 1 {
		return 1
	}
	return m
}

// ParseTags splits "go, Web Dev ,go" into ["go", "Web Dev"]: trimmed,
// de-duplicated (case-insensitively) and capped at max entries.
func ParseTags(raw string, max int) []string {
	seen := map[string]bool{}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		name := CollapseSpaces(part)
		if name == "" || utf8.RuneCountInString(name) > 32 {
			continue
		}
		key := Slugify(name)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, name)
		if len(out) == max {
			break
		}
	}
	return out
}
