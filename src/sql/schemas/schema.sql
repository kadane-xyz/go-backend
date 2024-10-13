CREATE TYPE vote_type AS ENUM ('up', 'down', 'none');

CREATE TABLE game_stats (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username TEXT REFERENCES account(username) ON DELETE CASCADE,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    elo INTEGER 
);