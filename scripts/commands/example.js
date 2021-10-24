/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_example(ch) {
    ch.send('Hello from an example command!\r\n');
}

Golem.registerPlayerCommand('example', do_example);
