# redactenv

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go dev](https://pkg.go.dev/badge/github.com/rgglez/go-storage/v5)](https://pkg.go.dev/gitub.com/rgglez/bash-env-redact/v5)
[![Go Report Card](https://goreportcard.com/badge/github.com/rgglez/bash-env-redact/v5)](https://goreportcard.com/report/github.com/rgglez/bash-env-redact/v5)
![GitHub stars](https://img.shields.io/github/stars/rgglez/bash-env-redact?style=social)
![GitHub forks](https://img.shields.io/github/forks/rgglez/bash-env-redact?style=social)

*redactenv* reads a bash/`.env`-style file where each line is either:

```bash
VARIABLE=value
export VARIABLE=value
```

and writes a copy with every sensitive VALUE replaced by a safe placeholder,
keeping the structure intact (variable names, comments, quoting style,
indentation, inline comments, blank lines, etc.).

The result is safe to paste into a chatbot or send over a messaging app.

## Usage

```bash
redactenv [flags] <input>
```

- `<input>` — file to read, or `-` for stdin.
- `-o <file>` — output file (default: stdout).
- `--strict` — redact everything, including booleans, numbers and common enums.
- `--strip-comments` — remove comments (they can contain secrets too).
- `--keep-private-ips` — do not redact RFC-1918 / loopback addresses.
- `--keep <list>` — comma-separated variable names to leave untouched.
- `--force <list>` — comma-separated variable names to always redact.

## Build

### Using Make

| Target | Output | Description |
|---|---|---|
| `make` / `make build` | `build/redactenv[.exe]` | Builds for the current OS/arch (auto-detected via `go env`) |
| `make linux` | `build/redactenv_linux` | Cross-compiles for `linux/amd64` |
| `make darwin` | `build/redactenv_mac` | Cross-compiles for `darwin/arm64` |
| `make windows` | `build/redactenv.exe` | Cross-compiles for `windows/amd64` |
| `make all-platforms` | all three above | Builds for all platforms in one step |
| `make clean` | — | Removes the `build/` directory |

Override OS or architecture: `make build GOOS=linux GOARCH=arm64`.

### Manually

```sh
go build -o redactenv redactenv.go
```

Cross-compile examples:

```sh
GOOS=linux   GOARCH=amd64  go build -o redactenv_linux   redactenv.go
GOOS=darwin  GOARCH=arm64  go build -o redactenv_mac     redactenv.go
GOOS=windows GOARCH=amd64  go build -o redactenv.exe     redactenv.go
```



## Architecture

The program is a single-pass pipeline with six source files, each with a single
responsibility:

```
┌─────────────────────────────────────────────────────────┐
│                        main.go                          │
│  Parse flags → read → process each line → write         │
└───────┬─────────────────────────┬───────────────────────┘
        │                         │
        ▼                         ▼
   ┌─────────┐             ┌─────────────┐
   │  io.go  │             │  parser.go  │
   │ readLines│             │ processLine │
   │writeLines│             │  splitRHS   │
   └─────────┘             └──────┬──────┘
                                  │
                    ┌─────────────┴──────────────┐
                    ▼                            ▼
             ┌─────────────┐           ┌──────────────────┐
             │ classify.go │           │  anonymiser.go   │
             │classifyValue│◄──────────│  Anonymiser      │
             │isSensitiveKey│          │  Redact()        │
             └─────────────┘           │  placeholder()   │
                                       └──────────────────┘
                    ▲                            ▲
                    └──────────┬─────────────────┘
                               │
                   ┌───────────┴──────────┐
                   │      patterns.go     │
                   │  sensitivePatterns   │
                   │  safeTextValues      │
                   │  boolLiterals        │
                   └──────────────────────┘
                               ▲
                   ┌───────────┴──────────┐
                   │      regexps.go      │
                   │  assignmentRE        │
                   │  emailRE, schemeRE…  │
                   └──────────────────────┘
```

### Processing pipeline

For each line in the input file:

1. **`parser.go` — `processLine`**
   - Skip or strip pure comment lines (`# …`).
   - Match the line against `assignmentRE` to extract `indent`, optional
     `export` keyword, `key`, and `rest`.
   - Call `splitRHS` to separate the raw value from its quoting style (`"`, `'`,
     or unquoted) and any trailing inline comment.
   - Check the `--keep` / `--force` override lists; otherwise delegate to
     `Anonymiser.Redact`.
   - Reconstruct the output line preserving original structure.

2. **`classify.go` — `classifyValue` + `isSensitiveKey`**
   - `isSensitiveKey` does a case-insensitive substring search of the variable
     name against `sensitivePatterns` (e.g. `password`, `token`, `key`…).
   - `classifyValue` probes the value with compiled regexps and `net.ParseIP` to
     return one of: `empty`, `bool`, `int`, `float`, `email`, `ip`, `url`,
     `path`, `text`.

3. **`anonymiser.go` — `Anonymiser.Redact`**
   - If `sensitive=false` and `strict=false`, harmless types (`bool`, `int`,
     `float`, `safeTextValues`) pass through unchanged.
   - Otherwise calls `placeholder(sensitive, type, rawValue)`.
   - A `map[cacheKey]string` cache ensures the same raw value always maps to the
     same placeholder (consistent substitution across the file).
   - Type-specific generators produce realistic-looking but safe replacements:
     RFC 5737 IPs, `@example.com` emails, sanitised URLs that keep scheme and
     port, paths that keep file extension.
   - If the generated placeholder contains shell metacharacters and the original
     was unquoted, double-quotes are added automatically.

4. **`io.go`** — thin I/O layer; supports `-` (stdin/stdout) for pipe-friendly
   use.

5. **`patterns.go` / `regexps.go`** — static data and pre-compiled regexps, kept
   separate to make the heuristic lists easy to audit and extend.

## Value detection strategy

Each value is classified by type (bool, integer, float, email, IP, URL, path,
plain text) and replaced by a placeholder of the same type so the redacted file
still conveys enough context to be useful:

- A variable whose **name** suggests a secret (`PASSWORD`, `TOKEN`, `KEY`,
  `SECRET`, `JWT`, …) is always redacted as `REDACTED_N`, regardless of the
  value.
- Emails become `userN@example.com`.
- IPv4 addresses become addresses from `203.0.113.0/24` (RFC 5737 documentation
  range). IPv6 becomes `2001:db8::` (RFC 3849).
- URLs keep their scheme and port but the host/path/credentials/query are
  replaced.
- Filesystem paths keep their extension.
- Repeated identical values always map to the same placeholder (consistent
  substitution), so relationships between variables are preserved.

In non-strict mode, boolean literals, numbers, and common configuration enums
(`production`, `debug`, `info`, …) are left as-is because they carry no personal
information.

### Limitations

* Sensitive pattern detection is based on English variable names only.

## License

Copyright (C) 2026 Rodolfo González González.

Released under the
[GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.html).
Please read the [LICENSE](LICENSE) file.