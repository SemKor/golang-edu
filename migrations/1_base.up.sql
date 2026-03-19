CREATE TABLE IF NOT EXISTS widgets (
    id integer PRIMARY KEY,
    name text NOT NULL,
    weight numeric NOT NULL,
    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp DEFAULT now() NOT NULL,
    deleted_at timestamp
);

INSERT INTO widgets (id, name, weight) VALUES
    (1, 'cup', 0.1) ON CONFLICT (id) DO UPDATE SET name = excluded.name, weight = excluded.weight;
