/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_example(ch) {
    try {
        if(!ch.room || !ch.room.plane || !ch.room.plane.map || !ch.room.plane.map.layers || !ch.room.plane.map.layers.length) {
            ch.send("You can't do that here.\r\n");
            return;
        }

        const permutation = Golem.util.generatePermutation();
        const terrain = ch.room.plane.map.layers[0].terrain;

        const w = 512;
        const h = 512;
        const values = [];
        let min = 0.0;
        let max = 0.0;

        for (let y = 0; y < h; y++) {
            for (let x = 0; x < w; x++) {
                const nx = x / w - 0.5;
                const ny = y / h - 0.5;

                const t = Golem.util.perlin2D(nx, ny, permutation);
                values.push(t);

                if(t > max) {
                    max = t;
                } else if(t < min) {
                    min = t;
                }
            }
        }

        for (let y = 0; y < h; y++) {
            for (let x = 0; x < w; x++) {
                const nx = x / w - 0.5;
                const ny = y / h - 0.5;

                const t = Golem.util.perlin2D(nx, ny, permutation);
                const normalized = (t-min)/(max-min);

                let newTerrain = Golem.TerrainTypes.TerrainTypeOcean;

                if(normalized < 0.2) {
                    newTerrain = Golem.TerrainTypes.TerrainTypeShallowWater;
                }

                terrain[x][y] = newTerrain;
            }
        }
    } catch(err) {
        ch.send(err.toString());
        return;
    }

    ch.send("Done!\r\n");
}

Golem.registerPlayerCommand('example', do_example);