/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"testing"
)

type oneArgumentTest struct {
	input, expectedFirstOutput, expectedSecondOutput string
}

var oneArgumentTests = []oneArgumentTest{
	{
		`unquoted "something else" here`,
		`unquoted`,
		`"something else" here`,
	},
	{
		`"something else" here`,
		`something else`,
		`here`,
	},
	{
		`cast 'power word test' target`,
		`cast`,
		`'power word test' target`,
	},
	{
		`'power word test' target`,
		`power word test`,
		`target`,
	},
	{
		`'quoted arg' 'second arg'`,
		`quoted arg`,
		`'second arg'`,
	},
	{
		`'second arg'`,
		`second arg`,
		``,
	},
}

func TestOneArgument(t *testing.T) {
	for _, test := range oneArgumentTests {
		if arg, rest := OneArgument(test.input); arg != test.expectedFirstOutput || rest != test.expectedSecondOutput {
			t.Errorf("OneArgument of %s returned %s and %s, expected %s and %s.\r\n", test.input, arg, rest, test.expectedFirstOutput, test.expectedSecondOutput)
		}
	}
}
