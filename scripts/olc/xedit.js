/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_xedit(ch, args) {
    function displayUsage() {
        ch.send(
            `{WExit editor usage:
{Gxedit delete <direction> - {gDestroy a permanent exit
{Gxedit dig <direction> - {gTry to create a new zone room in a direction
{Gxedit flags <direction> <flag name> - {gToggle a flag for a given direction 
`);
    }

    if (!ch.room) {
        ch.send("You can't do that here.\r\n");
        return;
    }

    let [firstArgument, rest] = Golem.util.oneArgument(args);

    switch (firstArgument) {
        default:
            displayUsage();
            break;
    }
}

Golem.registerPlayerCommand('xedit', do_xedit, Golem.Levels.LevelBuilder);