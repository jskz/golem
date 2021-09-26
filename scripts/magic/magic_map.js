/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_magic_map(ch, args) {
    const map = ch.createMazeMap();
    const mapLines = map.split(/\r\n|\r|\n/);
    const mapWidthRequired = mapLines[0].length;
    
    let output = "{Y  ,-" + "-".repeat(mapWidthRequired) + "-.\r\n"
        + " (_\\ " + " ".repeat(mapWidthRequired) + " \\\r\n"
        + mapLines
            .filter(line => line.length === mapWidthRequired)
            .map((line) => "   |{x " + line + " {Y|\r\n")
            .join("")
        + "  _| " + " ".repeat(mapWidthRequired) + " |\r\n"
        + " (_/_" + "_".repeat(mapWidthRequired) + "_/{x\r\n";
        
    ch.send(output);
}

Golem.registerSpellHandler('magic map', spell_magic_map);