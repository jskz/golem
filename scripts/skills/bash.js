/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_bash(ch, args) {
    if(!args.length) {
        ch.send("Bash who?\r\n");
        return;
    }

    ch.send("Bash handler executed with arguments: " + args + "\r\n");
}

Golem.registerSkillHandler('bash', do_bash);