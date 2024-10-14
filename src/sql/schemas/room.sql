CREATE TABLE room (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    problem_id BIGINT REFERENCES problems(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL DEFAULT 'open',
    max_participants INT NOT NULL DEFAULT 4,
    time_limit INT NOT NULL DEFAULT 3600,
    creator_id BIGINT REFERENCES account(id) ON DELETE SET NULL
);

CREATE TABLE room_participant (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    room_id BIGINT REFERENCES room(id) ON DELETE CASCADE,
    account_id BIGINT REFERENCES account(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL DEFAULT 'active',
    UNIQUE (room_id, account_id)
);

-- In the future it might be more wise to use Redis or some sort of cache
CREATE TABLE room_message (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    room_id BIGINT REFERENCES room(id) ON DELETE CASCADE,
    account_id BIGINT REFERENCES account(id) ON DELETE SET NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
