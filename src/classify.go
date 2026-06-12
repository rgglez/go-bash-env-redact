// classify.go
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

import (
	"net"
	"strings"
)

// valueType is a string tag returned by classifyValue.
type valueType string

const (
	typeEmpty valueType = "empty"
	typeBool  valueType = "bool"
	typeInt   valueType = "int"
	typeFloat valueType = "float"
	typeEmail valueType = "email"
	typeIP    valueType = "ip"
	typeURL   valueType = "url"
	typePath  valueType = "path"
	typeText  valueType = "text"
)

// classifyValue returns the valueType that best describes v.
func classifyValue(v string) valueType {
	s := strings.TrimSpace(v)
	if s == "" {
		return typeEmpty
	}
	low := strings.ToLower(s)
	if boolLiterals[low] {
		return typeBool
	}
	if integerRE.MatchString(s) {
		return typeInt
	}
	if floatRE.MatchString(s) {
		return typeFloat
	}
	if emailRE.MatchString(s) {
		return typeEmail
	}
	if schemeRE.MatchString(s) {
		return typeURL
	}
	// net.ParseIP handles both IPv4 and IPv6.
	if net.ParseIP(s) != nil {
		return typeIP
	}
	if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "~/") ||
		strings.HasPrefix(s, "./") || windowsPathRE.MatchString(s) {
		return typePath
	}
	return typeText
}

// isSensitiveKey returns true when the variable name contains any of the
// substrings in sensitivePatterns (case-insensitive comparison).
func isSensitiveKey(name string) bool {
	low := strings.ToLower(name)
	for _, p := range sensitivePatterns {
		if strings.Contains(low, p) {
			return true
		}
	}
	return false
}
