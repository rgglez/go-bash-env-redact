// parser.go
//
// Copyright (C) 2026 Rodolfo González González
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
package main

import "strings"

// splitRHS splits the right-hand side of an assignment into three parts:
//
//	quoteChar — quoting style: "", `"`, or `'`
//	value     — the inner content (without surrounding quotes)
//	tail      — everything that follows the closing quote or the value
//	            (trailing spaces, inline comments, …)
func splitRHS(rhs string) (quoteChar, value, tail string) {
	if len(rhs) == 0 {
		return "", "", ""
	}

	switch rhs[0] {
	case '"':
		// Double-quoted: handle backslash escapes (\" \\ \$ …).
		var buf strings.Builder
		i := 1
		for i < len(rhs) {
			c := rhs[i]
			if c == '\\' && i+1 < len(rhs) {
				buf.WriteByte(c)
				buf.WriteByte(rhs[i+1])
				i += 2
				continue
			}
			if c == '"' {
				return `"`, buf.String(), rhs[i+1:]
			}
			buf.WriteByte(c)
			i++
		}
		// Unclosed quote — treat remainder as the value.
		return `"`, rhs[1:], ""

	case '\'':
		// Single-quoted: content is entirely literal (no escapes in bash).
		close := strings.Index(rhs[1:], "'")
		if close >= 0 {
			return "'", rhs[1 : close+1], rhs[close+2:]
		}
		return "'", rhs[1:], ""

	default:
		// Unquoted: an inline comment starts with at least one space then "#".
		if loc := inlineCommentRE.FindStringIndex(rhs); loc != nil {
			return "", rhs[:loc[0]], rhs[loc[0]:]
		}
		// No inline comment: strip trailing whitespace, save it as tail.
		trimmed := strings.TrimRight(rhs, " \t")
		return "", trimmed, rhs[len(trimmed):]
	}
}

// processLine anonymises a single line and returns the result.
// Lines that are comments, blank, or not assignments are returned unchanged
// (subject to the stripComments flag).
func processLine(
	line string,
	anon *Anonymiser,
	stripComments bool,
	keep map[string]bool,
	force map[string]bool,
) string {
	// Pure comment lines.
	if commentLineRE.MatchString(line) {
		if stripComments {
			return ""
		}
		return line
	}

	m := assignmentRE.FindStringSubmatch(line)
	if m == nil {
		// Not an assignment (blank line, malformed line, etc.) — pass through.
		return line
	}

	// Extract named capture groups by index.
	names := assignmentRE.SubexpNames()
	groups := make(map[string]string, len(names))
	for i, name := range names {
		if name != "" {
			groups[name] = m[i]
		}
	}

	indent := groups["indent"]
	export := groups["export"]
	key := groups["key"]
	rhs := groups["rest"]

	quoteChar, value, tail := splitRHS(rhs)

	// Explicit control lists take precedence over heuristics.
	var outQuote, outValue string
	switch {
	case keep[key]:
		outQuote, outValue = quoteChar, value
	case force[key]:
		// Force redaction by pretending the key name is SECRET_<key>.
		outQuote, outValue = anon.Redact("SECRET_"+key, quoteChar, value)
	default:
		outQuote, outValue = anon.Redact(key, quoteChar, value)
	}

	// Reconstruct the value token with its (possibly updated) quotes.
	var valueToken string
	if outQuote != "" {
		valueToken = outQuote + outValue + outQuote
	} else {
		valueToken = outValue
	}

	// Drop inline comments if requested.
	if stripComments {
		tail = strings.TrimRight(tail, " \t")
		if strings.HasPrefix(strings.TrimLeft(tail, " \t"), "#") {
			tail = ""
		}
	}

	return indent + export + key + "=" + valueToken + tail
}
