/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function onReload() {
    Golem.clearAllEventHandlers();
}

Golem.registerEventHandler('reload', onReload);