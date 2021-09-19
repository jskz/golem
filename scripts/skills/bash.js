/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_bash(ch) {
    ch.send("Bash handler!\r\n");
}

Golem.registerSkillHandler('bash', do_bash);