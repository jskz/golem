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
    const firstMapLineWithoutColourCodes = mapLines[0].replace(/{\w/gi, '');
    const mapWidthRequired = firstMapLineWithoutColourCodes.length;

    ch.send('{MA luminous ethereal scroll appears in front of you!{x\r\n\r\n');

    let output =
        '{Y  ,-' +
        '-'.repeat(mapWidthRequired) +
        '-.\r\n' +
        ' (_\\ ' +
        ' '.repeat(mapWidthRequired) +
        ' \\\r\n' +
        mapLines
            .filter(
                (line) => line.replace(/{\w/gi, '').length === mapWidthRequired
            )
            .map((line) => '   |{x ' + line + ' {Y|\r\n')
            .join('') +
        '  _| ' +
        ' '.repeat(mapWidthRequired) +
        ' |\r\n' +
        ' (_/_' +
        '_'.repeat(mapWidthRequired) +
        '_/{x\r\n';

    ch.send(output);
}

Golem.registerSpellHandler('magic map', spell_magic_map);
