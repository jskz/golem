CREATE TABLE shops (
    `id` INTEGER PRIMARY KEY,

    `mobile_id` BIGINT NOT NULL UNIQUE,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (mobile_id) REFERENCES mobiles(id)
);

CREATE TABLE shop_object (
    `id` INTEGER PRIMARY KEY,

    `shop_id` BIGINT NOT NULL,
    `object_id` BIGINT NOT NULL,
    `price` BIGINT NOT NULL,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE,
    FOREIGN KEY (`object_id`) REFERENCES objects(id) ON DELETE CASCADE
);

INSERT INTO shops (id, mobile_id) VALUES (1, 3);

INSERT INTO shop_object (id, shop_id, `object_id`, price) VALUES (7, 1, 7, 1000);
INSERT INTO shop_object (id, shop_id, `object_id`, price) VALUES (6, 1, 12, 250);
INSERT INTO shop_object (id, shop_id, `object_id`, price) VALUES (5, 1, 6, 200);
INSERT INTO shop_object (id, shop_id, `object_id`, price) VALUES (4, 1, 9, 150);
INSERT INTO shop_object (id, shop_id, `object_id`, price) VALUES (3, 1, 8, 100);
INSERT INTO shop_object (id, shop_id, `object_id`, price) VALUES (2, 1, 10, 50);
INSERT INTO shop_object (id, shop_id, `object_id`, price) VALUES (1, 1, 5, 10);
