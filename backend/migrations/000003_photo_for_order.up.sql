CREATE TABLE IF NOT EXISTS order_photos (
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    photo_name VARCHAR(255) NOT NULL,

    PRIMARY KEY (order_id, photo_name)
);
