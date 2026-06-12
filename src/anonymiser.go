// anonymiser.go
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
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"strings"
)

// cacheKey uniquely identifies a (sensitive-flag, type, raw-value) triple so
// that the same original value always produces the same placeholder.
type cacheKey struct {
	sensitive bool
	vtype     valueType
	raw       string
}

// -----------------------------------------------------------------------------

// Anonymiser replaces values with type-preserving placeholders.
type Anonymiser struct {
	strict         bool
	keepPrivateIPs bool
	cache          map[cacheKey]string
	counters       map[string]int // keyed by a human-readable category name
}

// -----------------------------------------------------------------------------

// newAnonymiser allocates a ready-to-use Anonymiser.
func newAnonymiser(strict, keepPrivateIPs bool) *Anonymiser {
	return &Anonymiser{
		strict:         strict,
		keepPrivateIPs: keepPrivateIPs,
		cache:          make(map[cacheKey]string),
		counters:       make(map[string]int),
	}
}

// -----------------------------------------------------------------------------

// next increments the counter for category cat and returns the new value.
func (a *Anonymiser) next(cat string) int {
	a.counters[cat]++
	return a.counters[cat]
}

// -----------------------------------------------------------------------------

// totalRedacted returns the sum of all counter values.
func (a *Anonymiser) totalRedacted() int {
	total := 0
	for _, v := range a.counters {
		total += v
	}
	return total
}

// -----------------------------------------------------------------------------

// anonEmail returns a fake but structurally valid email address.
func (a *Anonymiser) anonEmail() string {
	return fmt.Sprintf("user%d@example.com", a.next("email"))
}

// -----------------------------------------------------------------------------

// anonIP returns an address from a documentation-reserved range so it
// cannot be confused with real infrastructure.
func (a *Anonymiser) anonIP(original string) string {
	ip := net.ParseIP(original)
	if ip == nil {
		return fmt.Sprintf("203.0.113.%d", a.next("ip"))
	}
	// Preserve private-range addresses if the flag is set.
	if a.keepPrivateIPs && (ip.IsPrivate() || ip.IsLoopback()) {
		return original
	}
	n := a.next("ip")
	if ip.To4() != nil {
		// 203.0.113.0/24 — TEST-NET-3, RFC 5737.
		return fmt.Sprintf("203.0.113.%d", (n%254)+1)
	}
	// 2001:db8::/32 — documentation prefix, RFC 3849.
	return fmt.Sprintf("2001:db8::%x", n)
}

// -----------------------------------------------------------------------------

// anonURL returns a URL that keeps the scheme and port but replaces the host,
// path, credentials, query string and fragment.
func (a *Anonymiser) anonURL(original string) string {
	n := a.next("url")
	placeholder := fmt.Sprintf("https://host%d.example.com", n)

	parsed, err := url.Parse(original)
	if err != nil {
		return placeholder
	}
	scheme := parsed.Scheme
	if scheme == "" {
		scheme = "https"
	}
	host := fmt.Sprintf("host%d.example.com", n)
	// Keep the port number — it usually conveys which service is running.
	if port := parsed.Port(); port != "" {
		host = host + ":" + port
	}
	// Replace any non-trivial path with a generic token.
	path := ""
	if parsed.Path != "" && parsed.Path != "/" {
		path = "/redacted-path"
	}
	// Drop credentials, query string and fragment for safety.
	out := &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}
	return out.String()
}

// -----------------------------------------------------------------------------

// anonPath returns a generic path that preserves the file extension.
func (a *Anonymiser) anonPath(original string) string {
	n := a.next("path")
	ext := filepath.Ext(original) // returns "" when there is no extension
	return fmt.Sprintf("/redacted/path-%d%s", n, ext)
}

// -----------------------------------------------------------------------------

// placeholder returns the replacement string for the given (key, type, value)
// triple, generating and caching a new one if necessary.
func (a *Anonymiser) placeholder(sensitive bool, vt valueType, raw string) string {
	k := cacheKey{sensitive, vt, raw}
	if cached, ok := a.cache[k]; ok {
		return cached
	}
	var result string
	if sensitive {
		result = fmt.Sprintf("REDACTED_%d", a.next("secret"))
	} else {
		switch vt {
		case typeEmail:
			result = a.anonEmail()
		case typeIP:
			// anonIP may return the original unchanged (keepPrivateIPs).
			result = a.anonIP(raw)
		case typeURL:
			result = a.anonURL(raw)
		case typePath:
			result = a.anonPath(raw)
		case typeBool:
			// Only reached in strict mode.
			result = "false"
		case typeInt:
			result = "0"
		case typeFloat:
			result = "0.0"
		default: // typeText
			result = fmt.Sprintf("value_anon_%d", a.next("text"))
		}
	}
	a.cache[k] = result
	return result
}

// -----------------------------------------------------------------------------

// Redact decides whether value should be replaced and, if so, returns the
// placeholder; otherwise it returns the original value unchanged.
// quoteChar is the quoting style that surrounded the value in the source
// ("", `"`, or `'`).  Returns the (possibly updated) quote character and
// the (possibly replaced) value.
func (a *Anonymiser) Redact(key, quoteChar, value string) (string, string) {
	if value == "" {
		return quoteChar, value
	}

	sensitive := isSensitiveKey(key)
	vt := classifyValue(value)

	// In non-strict mode, keep values that are obviously safe.
	if !sensitive && !a.strict {
		switch vt {
		case typeBool, typeInt, typeFloat:
			return quoteChar, value
		case typeText:
			if safeTextValues[strings.ToLower(strings.TrimSpace(value))] {
				return quoteChar, value
			}
		}
	}

	repl := a.placeholder(sensitive, vt, value)

	// Choose the output quote character.
	outQuote := quoteChar
	if outQuote == "" && needsQuoting(repl) {
		outQuote = `"`
	}
	return outQuote, repl
}

// -----------------------------------------------------------------------------

// needsQuoting returns true when v contains characters that would be
// interpreted by the shell if left unquoted.
func needsQuoting(v string) bool {
	return v == "" || needsQuotingRE.MatchString(v)
}
