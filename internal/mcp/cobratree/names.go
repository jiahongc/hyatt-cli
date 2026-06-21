// Copyright 2026 Jiahong Chen and contributors. Licensed under Apache-2.0. See LICENSE.
// Maintained in the World of Hyatt CLI repository.

package cobratree

import (
	"strings"
	"unicode"
)

func toolNameForPath(parts []string) string {
	var out []rune
	for _, part := range parts {
		for _, r := range part {
			switch {
			case unicode.IsLetter(r) || unicode.IsDigit(r):
				out = append(out, unicode.ToLower(r))
			default:
				out = append(out, '_')
			}
		}
		out = append(out, '_')
	}
	return strings.Trim(strings.Join(strings.FieldsFunc(string(out), func(r rune) bool { return r == '_' }), "_"), "_")
}
