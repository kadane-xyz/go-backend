CREATE TABLE account (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    verified BOOLEAN DEFAULT FALSE,
    avatar_url TEXT,
    attributesId BIGINT REFERENCES account_attributes(id) ON DELETE CASCADE
);

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

CREATE TABLE account_solved_problems (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username TEXT REFERENCES account(username) ON DELETE CASCADE,
    problem_id BIGINT REFERENCES problems(id) ON DELETE CASCADE,
    solved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE account_game_stats (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username TEXT REFERENCES account(username) ON DELETE CASCADE,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    elo INTEGER 
);