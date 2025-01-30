CREATE TYPE mode as ENUM ('classic', 'team');

CREATE TABLE room (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    admin text NOT NULL REFERENCES account(id),
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    max_players int NOT NULL DEFAULT 8,
    visibility visibility NOT NULL DEFAULT 'public',
    difficulty TEXT NOT NULL DEFAULT 'easy',
    mode mode NOT NULL DEFAULT 'classic',
    whitelist uuid[] NOT NULL DEFAULT '{}'
);

CREATE TABLE room_problems (
    room_id uuid REFERENCES room(id) ON DELETE CASCADE,
    problem_id int REFERENCES problem(id) ON DELETE CASCADE,
    problem_order int NOT NULL,
    PRIMARY KEY (room_id, problem_id)
);