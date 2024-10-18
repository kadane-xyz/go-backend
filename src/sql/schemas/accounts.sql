CREATE TABLE account_attributes (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    bio TEXT,
    location TEXT,
    realName TEXT NOT NULL,
    github_url TEXT,
    linkedin_url TEXT,
    facebook_url TEXT,
    twitter_url TEXT,
    school TEXT
);

CREATE TABLE account (
    id TEXT NOT NULL UNIQUE PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    avatar_url TEXT,
    attributesId BIGINT REFERENCES account_attributes(id) ON DELETE CASCADE,
    level INTEGER DEFAULT 1
);

CREATE TABLE account_solved_problems (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id TEXT REFERENCES account(id) ON DELETE CASCADE,
    problem_id BIGINT REFERENCES problems(id) ON DELETE CASCADE,
    solved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE account_game_stats (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id TEXT REFERENCES account(id) ON DELETE CASCADE,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    elo INTEGER 
);