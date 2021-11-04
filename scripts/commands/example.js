/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_example(ch) {
    Golem.StringEditor(ch.client,
        "Testing string editor test text...\r\nSecond line\r\nHey",
        function(_, string) {
            ch.send('{CResulting string was: ' + string + '{x\r\n');
        });
}

Golem.registerPlayerCommand('example', do_example);
