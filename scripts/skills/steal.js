/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_steal(ch, args) {
    let victim =
        ch.fighting !== null ? ch.fighting : ch.findCharacterInRoom(args);

    if (!victim) {
        ch.send('Steal from who?\r\n');
        return;
    }
}

Golem.registerSkillHandler('steal', do_steal);
