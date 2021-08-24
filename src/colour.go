/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"regexp"
)

const AnsiBoldRed = "\u001b[31;1m"
const AnsiRed = "\u001b[0m\u001b[31m"
const AnsiGreen = "\u001b[0m\u001b[32m"
const AnsiBoldGreen = "\u001b[32;1m"
const AnsiYellow = "\u001b[0m\u001b[33m"
const AnsiBoldYellow = "\u001b[33;1m"
const AnsiBoldBlue = "\u001b[34;1m"
const AnsiBlue = "\u001b[0m\u001b[34m"
const AnsiBoldMagenta = "\u001b[35;1m"
const AnsiMagenta = "\u001b[0m\u001b[35m"
const AnsiCyan = "\u001b[0m\u001b[36m"
const AnsiBoldCyan = "\u001b[36;1m"
const AnsiWhite = "\u001b[0m\u001b[37m"
const AnsiBoldWhite = "\u001b[37;1m"
const AnsiBoldBlack = "\u001b[30;1m"
const AnsiReset = "\u001b[0m"

type AnsiColourCodeTableEntry struct {
	ansiEscapeSequence string
	regularExpression  *regexp.Regexp
}

var AnsiColourCodeTable map[string]AnsiColourCodeTableEntry

func init() {
	AnsiColourCodeTable = map[string]AnsiColourCodeTableEntry{
		"{D": {
			ansiEscapeSequence: AnsiBoldBlack,
			regularExpression:  regexp.MustCompile("{D"),
		},
		"{R": {
			ansiEscapeSequence: AnsiBoldRed,
			regularExpression:  regexp.MustCompile("{R"),
		},
		"{r": {
			ansiEscapeSequence: AnsiRed,
			regularExpression:  regexp.MustCompile("{r"),
		},
		"{G": {
			ansiEscapeSequence: AnsiBoldGreen,
			regularExpression:  regexp.MustCompile("{G"),
		},
		"{g": {
			ansiEscapeSequence: AnsiGreen,
			regularExpression:  regexp.MustCompile("{g"),
		},
		"{B": {
			ansiEscapeSequence: AnsiBoldBlue,
			regularExpression:  regexp.MustCompile("{B"),
		},
		"{b": {
			ansiEscapeSequence: AnsiBlue,
			regularExpression:  regexp.MustCompile("{b"),
		},
		"{Y": {
			ansiEscapeSequence: AnsiBoldYellow,
			regularExpression:  regexp.MustCompile("{Y"),
		},
		"{y": {
			ansiEscapeSequence: AnsiYellow,
			regularExpression:  regexp.MustCompile("{y"),
		},

		"{C": {
			ansiEscapeSequence: AnsiBoldCyan,
			regularExpression:  regexp.MustCompile("{C"),
		},
		"{c": {
			ansiEscapeSequence: AnsiCyan,
			regularExpression:  regexp.MustCompile("{c"),
		},
		"{M": {
			ansiEscapeSequence: AnsiBoldMagenta,
			regularExpression:  regexp.MustCompile("{M"),
		},
		"{m": {
			ansiEscapeSequence: AnsiMagenta,
			regularExpression:  regexp.MustCompile("{m"),
		},
		"{W": {
			ansiEscapeSequence: AnsiBoldWhite,
			regularExpression:  regexp.MustCompile("{W"),
		},
		"{w": {
			ansiEscapeSequence: AnsiWhite,
			regularExpression:  regexp.MustCompile("{w"),
		},
		"{x": {
			ansiEscapeSequence: AnsiReset,
			regularExpression:  regexp.MustCompile("{x"),
		},
	}
}

func TranslateColourCodes(s string) string {
	input := string(s)

	for colour := range AnsiColourCodeTable {
		input = AnsiColourCodeTable[colour].regularExpression.ReplaceAllString(
			input,
			AnsiColourCodeTable[colour].ansiEscapeSequence)
	}

	return input
}
