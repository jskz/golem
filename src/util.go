/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"strings"
	"unicode"
)

func oneArgument(args string) (string, string) {
	var buf strings.Builder
	var quoted bool = false
	var end int = len(args)

	for index, r := range args {
		if r == '\'' || r == '"' {
			if quoted {
				end = index
				break
			}

			quoted = true
		} else {
			if r != ' ' || quoted {
				buf.WriteRune(unicode.ToLower(r))
			} else if r == ' ' && !quoted {
				end = index
				break
			}
		}
	}

	if quoted {
		end++
	}

	return buf.String(), strings.TrimLeft(args[end:], " ")
}
