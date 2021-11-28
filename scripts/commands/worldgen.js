function do_worldgen(ch, args) {
    try {
        const BUILDING_POSITION = [49, 418],
            BUILDING_WIDTH = 12,
            BUILDING_HEIGHT = 6;
        const p = ch.room.plane;
        const w = p.width;
        const h = p.height;
        const terrain = p.map.layers[0].terrain;

        for (let y = 0; y < h; y++) {
            for (let x = 0; x < w; x++) {
                let nx = x / w - 0.5,
                    ny = y / h - 0.5;
                let n = Golem.util.perlin2D(nx * 1.25, ny * 1.25);

                n += 1.0;
                n /= 2.0;

                let t = Golem.TerrainTypes.TerrainTypeOcean;

                if (n > 0.47) {
                    t = Golem.TerrainTypes.TerrainTypeShallowWater;
                }
                if (n > 0.485) {
                    t = Golem.TerrainTypes.TerrainTypeShore;
                }
                if (n > 0.50) {
                    t = Golem.TerrainTypes.TerrainTypePlains;
                }
                if (n > 0.53) {
                    t = Golem.TerrainTypes.TerrainTypeField;
                }
                if (n > 0.60) {
                    t = Golem.TerrainTypes.TerrainTypeLightForest;
                }
                if (n > 0.67) {
                    t = Golem.TerrainTypes.TerrainTypeDenseForest;
                }
                if (n > 0.73) {
                    t = Golem.TerrainTypes.TerrainTypeHills;
                }
                if (n > 0.85) {
                    t = Golem.TerrainTypes.TerrainTypeMountains;
                }
                if (n > 1.1) {
                    t = Golem.TerrainTypes.TerrainTypeSnowcappedMountains;
                }

                terrain[y][x] = t;
            }
        }
    } catch (err) {
        ch.send(err.toString());
    }

    ch.send("Done!\r\n");
}

Golem.registerPlayerCommand('worldgen', do_worldgen);