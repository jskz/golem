/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_bash(ch, args) {
    let victim = ch.fighting !== null 
        ? ch.fighting
        : ch.findCharacterInRoom(args);
    if(!victim) {
        ch.send("Bash who?\r\n");
        return;
    }

    ch.send("You bash into " + victim.getShortDescription(ch) + ", sending them flying!\r\n");
}

Golem.registerSkillHandler('bash', do_bash);