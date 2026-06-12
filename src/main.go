// main.go
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
	"flag"
	"fmt"
	"os"
)

func main() {
	// --- Flag definitions ---------------------------------------------------
	output := flag.String("o", "-", "Output file ('-' for stdout).")
	strict := flag.Bool("strict", false,
		"Redact everything, including booleans, numbers and common enums.")
	stripComments := flag.Bool("strip-comments", false,
		"Remove comments (they can contain secrets too).")
	keepPrivateIPs := flag.Bool("keep-private-ips", false,
		"Do not redact RFC-1918/loopback addresses (10.x, 192.168.x, …).")
	keepCSV := flag.String("keep", "",
		"Comma-separated list of variable names to leave untouched.")
	forceCSV := flag.String("force", "",
		"Comma-separated list of variable names to always redact.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"Usage: redactenv [flags] <input>\n\n"+
				"  <input>  File to read, or - for stdin.\n\n"+
				"Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	inputPath := flag.Arg(0)

	// --- Build lookup sets ---------------------------------------------------
	keep := splitCSV(*keepCSV)
	force := splitCSV(*forceCSV)

	// --- Read ----------------------------------------------------------------
	lines, err := readLines(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "redactenv: cannot read %q: %v\n", inputPath, err)
		os.Exit(1)
	}

	// --- Process -------------------------------------------------------------
	anon := newAnonymiser(*strict, *keepPrivateIPs)
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = processLine(l, anon, *stripComments, keep, force)
	}

	// --- Write ---------------------------------------------------------------
	if err := writeLines(*output, out); err != nil {
		fmt.Fprintf(os.Stderr, "redactenv: cannot write %q: %v\n", *output, err)
		os.Exit(1)
	}

	// Print a summary to stderr so it does not pollute the redacted output.
	fmt.Fprintf(os.Stderr,
		"[redactenv] values redacted: %d  (by category: %v)\n",
		anon.totalRedacted(), anon.counters)
}
