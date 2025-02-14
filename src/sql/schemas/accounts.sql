CREATE TYPE account_type AS ENUM ('free', 'plus', 'pro');

CREATE TABLE account (
    id TEXT NOT NULL UNIQUE PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    avatar_url TEXT,
    admin BOOLEAN DEFAULT FALSE,
    type account_type DEFAULT 'free',
    level INTEGER DEFAULT 1
);

CREATE TABLE account_attribute (
    id TEXT PRIMARY KEY REFERENCES account(id) ON DELETE CASCADE,
    bio TEXT DEFAULT '',
    contact_email TEXT DEFAULT '',
    location TEXT DEFAULT '',
    real_name TEXT DEFAULT '',
    github_url TEXT DEFAULT '',
    linkedin_url TEXT DEFAULT '',
    facebook_url TEXT DEFAULT '',
    instagram_url TEXT DEFAULT '',
    twitter_url TEXT DEFAULT '',
    school TEXT DEFAULT '',
    website_url TEXT DEFAULT ''
);

CREATE TABLE account_game_stat (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id TEXT REFERENCES account(id) ON DELETE CASCADE,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    elo INTEGER 
);