CREATE TYPE mode as ENUM ('CLASSIC', 'TEAM');

CREATE TABLE room (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    admin uuid NOT NULL REFERENCES account(id),
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    max_players int NOT NULL DEFAULT 8,
    visibility visibility NOT NULL DEFAULT 'PUBLIC',
    problems uuid[] NOT NULL REFERENCES problem(id),
    difficulty TEXT NOT NULL DEFAULT 'EASY',
    mode mode NOT NULL DEFAULT 'CLASSIC',
    whitelist uuid[] NOT NULL REFERENCES account(id)
);
