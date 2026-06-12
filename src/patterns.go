// patterns.go
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

// sensitivePatterns holds lowercase substrings that, when found inside a
// variable name, indicate the value must always be redacted.  The list is
// intentionally broad — better safe than sorry.  Override individual
// variables with --keep when the heuristic fires on a non-sensitive key.
var sensitivePatterns = []string{
	"pass", "passwd", "pwd", "secret", "token", "api_key", "apikey",
	"key", "private", "credential", "cred", "auth", "access",
	"session", "salt", "cert", "ssh", "dsn", "jwt", "bearer",
	"client_secret", "client_id", "signature", "sign", "encrypt",
	"license", "otp", "pin", "webhook", "database_url", "db_url",
	"smtp", "mail_pass", "aws", "gcp", "azure", "sentry",
}

// -----------------------------------------------------------------------------

// safeTextValues lists lowercase plain-text values that carry no personal
// information and are therefore left as-is in non-strict mode.
var safeTextValues = map[string]bool{
	"production": true, "staging": true, "development": true,
	"dev": true, "prod": true, "test": true, "local": true,
	"debug": true, "info": true, "warn": true, "warning": true,
	"error": true, "critical": true, "trace": true, "verbose": true,
	"silent": true, "none": true, "null": true, "default": true,
	"auto": true, "enabled": true, "disabled": true, "utf-8": true,
	"utf8": true, "json": true, "text": true, "http": true,
	"https": true, "tcp": true, "udp": true, "localhost": true,
}

// -----------------------------------------------------------------------------

// boolLiterals lists the values treated as boolean in non-strict mode.
var boolLiterals = map[string]bool{
	"true": true, "false": true, "yes": true, "no": true,
	"on": true, "off": true, "0": true, "1": true,
}
