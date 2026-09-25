CREATE TYPE user_role AS ENUM ('user', 'admin');

CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                       email VARCHAR UNIQUE CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),

                       password_hash TEXT NOT NULL
                           CHECK (btrim(password_hash) <> ''),

                       role user_role NOT NULL DEFAULT 'user'
                           CHECK (role IN ('user', 'admin')),

                       created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE auth_sessions (
                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                               user_id UUID NOT NULL
                                   REFERENCES users(id) ON DELETE CASCADE,

                               refresh_token_hash TEXT NOT NULL UNIQUE
                                   CHECK (btrim(refresh_token_hash) <> ''),

                               expires_at TIMESTAMPTZ NOT NULL,
                               revoked_at TIMESTAMPTZ,
                               created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

                               CHECK (expires_at > created_at),
                               CHECK (revoked_at IS NULL OR revoked_at >= created_at)
);


CREATE TABLE fillings (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                          name VARCHAR(150) NOT NULL
                              CHECK (btrim(name) <> ''),

                          description TEXT,

                          price_per_kg NUMERIC(12, 2) NOT NULL
                              CHECK (price_per_kg > 0),

                          image_name TEXT,
                          is_active BOOLEAN NOT NULL DEFAULT TRUE,

                          created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


CREATE TABLE orders (
                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                        user_id UUID NOT NULL
                            REFERENCES users(id) ON DELETE RESTRICT,

                        idempotency_key UUID NOT NULL,

                        weight_grams INTEGER NOT NULL
                            CHECK (weight_grams BETWEEN 1000 AND 7000),

                        decoration_wishes TEXT,

                        delivery_address TEXT NOT NULL
                            CHECK (btrim(delivery_address) <> ''),

                        total_price NUMERIC(12, 2) NOT NULL
                            CHECK (total_price > 0),

                        status VARCHAR(20) NOT NULL DEFAULT 'accepted'
                            CHECK (
                                status IN (
                                           'accepted',
                                           'processing',
                                           'ready',
                                           'cancelled'
                                    )
                                ),

                        created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                        updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

                        UNIQUE (user_id, idempotency_key)
);


CREATE TABLE order_fillings (
                                order_id UUID NOT NULL
                                    REFERENCES orders(id) ON DELETE CASCADE,

                                filling_id UUID NOT NULL
                                    REFERENCES fillings(id) ON DELETE RESTRICT,

                                filling_name VARCHAR(150) NOT NULL
                                    CHECK (btrim(filling_name) <> ''),

                                filling_price_per_kg NUMERIC(12, 2) NOT NULL
                                    CHECK (filling_price_per_kg > 0),

                                weight_grams INTEGER NOT NULL
                                    CHECK (weight_grams BETWEEN 1 AND 7000),

                                PRIMARY KEY (order_id, filling_id)
);
