// regexps.go
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

import "regexp"

var (
	// assignmentRE matches a full assignment line.  Named groups:
	//   indent  — leading whitespace
	//   export  — "export " keyword (optional)
	//   key     — variable name
	//   rest    — everything after "="
	assignmentRE = regexp.MustCompile(
		`^(?P<indent>\s*)(?P<export>export\s+)?(?P<key>[A-Za-z_][A-Za-z0-9_]*)=(?P<rest>.*)$`,
	)

	// commentLineRE matches lines that are purely a comment (possibly indented).
	commentLineRE = regexp.MustCompile(`^\s*#`)

	// inlineCommentRE matches a trailing inline comment (space(s) then #…).
	inlineCommentRE = regexp.MustCompile(`(\s+#.*)$`)

	// integerRE matches an optionally-signed integer literal.
	integerRE = regexp.MustCompile(`^[+-]?\d+$`)

	// floatRE matches an optionally-signed decimal literal.
	floatRE = regexp.MustCompile(`^[+-]?\d*\.\d+$`)

	// emailRE matches a minimal email address (no spaces, one @).
	emailRE = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

	// schemeRE matches the leading "scheme://" of a URL.
	schemeRE = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.\-]*://`)

	// windowsPathRE matches Windows-style absolute paths (C:\…).
	windowsPathRE = regexp.MustCompile(`^[A-Za-z]:\\`)

	// needsQuotingRE lists shell metacharacters that require the value to be
	// quoted when it appears unquoted on the right-hand side of an assignment.
	needsQuotingRE = regexp.MustCompile(`[\s<>#$&|;()*?\[\]{}!` + "`" + `"'\\~=]`)
)
