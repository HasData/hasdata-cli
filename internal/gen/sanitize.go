package main

import (
	"go/token"
	"strings"
	"unicode"
)

// reservedFlags are global flags on root that must not collide with a generated
// per-command flag. Generator renames collisions to "<name>-param".
var reservedFlags = map[string]bool{
	"help":     true,
	"h":        true,
	"api-key":  true,
	"config":   true,
	"endpoint": true,
	"output":   true,
	"o":        true,
	"pretty":   true,
	"raw":      true,
	"verbose":  true,
	"v":        true,
	"timeout":  true,
	"retries":  true,
	"version":  true,
}

// camelToKebab converts "blockAds" -> "block-ads", "URL" -> "url",
// "extractEmails" -> "extract-emails". Digits are treated as lowercase.
// Bracket indexing like "baths[max]" -> "baths-max", "home_types[]" -> "home-types".
func camelToKebab(s string) string {
	// First, normalize bracket-indexed forms used by some schemas
	// (e.g. "baths[max]" → "baths-max", "homeTypes[]" → "homeTypes").
	s = strings.ReplaceAll(s, "[]", "")
	s = strings.ReplaceAll(s, "[", "-")
	s = strings.ReplaceAll(s, "]", "")

	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := runes[i-1]
				next := rune(0)
				if i+1 < len(runes) {
					next = runes[i+1]
				}
				// Insert a dash before an uppercase rune if preceded by a
				// lowercase/digit, or if followed by a lowercase (e.g. "URLPath" -> "url-path").
				if unicode.IsLower(prev) || unicode.IsDigit(prev) ||
					(unicode.IsUpper(prev) && next != 0 && unicode.IsLower(next)) {
					b.WriteByte('-')
				}
			}
			b.WriteRune(unicode.ToLower(r))
		} else if r == '_' || r == '-' {
			b.WriteByte('-')
		} else {
			b.WriteRune(r)
		}
	}
	// Collapse double dashes that can result from bracket normalization.
	out := b.String()
	for strings.Contains(out, "--") {
		out = strings.ReplaceAll(out, "--", "-")
	}
	return strings.Trim(out, "-")
}

// flagName returns the CLI flag name for a given schema property. Collisions
// with global/reserved flags are suffixed with "-param".
func flagName(propName string) string {
	n := camelToKebab(propName)
	if reservedFlags[n] {
		n = n + "-param"
	}
	return n
}

// slugToIdent converts "google-serp" -> "GoogleSerp" for use as a Go identifier.
func slugToIdent(slug string) string {
	var b strings.Builder
	capitalizeNext := true
	for _, r := range slug {
		switch {
		case r == '-' || r == '_' || r == '.' || r == ' ':
			capitalizeNext = true
		case capitalizeNext:
			b.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		default:
			b.WriteRune(r)
		}
	}
	s := b.String()
	if s == "" {
		s = "X"
	}
	if unicode.IsDigit(rune(s[0])) {
		s = "X" + s
	}
	if token.IsKeyword(strings.ToLower(s)) {
		s = s + "_"
	}
	return s
}

// slugToSnake converts "google-serp" -> "google_serp" for filenames.
func slugToSnake(slug string) string {
	return strings.ReplaceAll(slug, "-", "_")
}

// varName converts a property name to a Go-safe variable name.
// "blockAds" -> "blockAdsVar", "2fa" -> "x2faVar", "type" -> "typeVar",
// "baths[max]" -> "bathsMaxVar", "homeTypes[]" -> "homeTypesVar".
func varName(propName string) string {
	propName = strings.ReplaceAll(propName, "[]", "")
	propName = strings.ReplaceAll(propName, "[", "_")
	propName = strings.ReplaceAll(propName, "]", "")
	var b strings.Builder
	for i, r := range propName {
		if i == 0 && !unicode.IsLetter(r) && r != '_' {
			b.WriteByte('x')
		}
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	s := b.String() + "Var"
	return s
}
