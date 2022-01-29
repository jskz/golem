/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_zedit(ch, args) {
    function displayUsage() {
        ch.send(
            `{WZone editor usage:
                
{Gzedit save             - {gSave zone properties
{x`);
    }

    if (!ch.room) {
        ch.send("You can't do that here.\r\n");
        return;
    }

    let [firstArgument, xs] = Golem.util.oneArgument(args);
    let [secondArgument, xxs] = Golem.util.oneArgument(xs);

    if(!args.length) {
        displayUsage();
        return;
    }

    switch (firstArgument) {
        case 'save':
            ch.send("Ok.\r\n");
            return;

        default:
            displayUsage();
            return;
    }
}

Golem.registerPlayerCommand('zedit', do_zedit, Golem.Levels.LevelAdmin);