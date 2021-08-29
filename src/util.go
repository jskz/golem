/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

func SeverityColourFromPercentage(percentage int) string {
	if percentage < 10 {
		return "{D"
	} else if percentage < 20 {
		return "{r"
	} else if percentage < 30 {
		return "{R"
	} else if percentage < 50 {
		return "{y"
	} else if percentage < 75 {
		return "{Y"
	} else if percentage < 90 {
		return "{g"
	}

	return "{G"
}
