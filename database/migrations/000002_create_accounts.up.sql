CREATE TYPE account_plan AS ENUM ('free', 'plus', 'pro');

CREATE TABLE account (
    id TEXT NOT NULL UNIQUE PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    avatar_url TEXT,
    admin BOOLEAN DEFAULT FALSE NOT NULL,
    plan account_plan DEFAULT 'free' NOT NULL,
    level INTEGER DEFAULT 1 NOT NULL
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

CREATE OR REPLACE FUNCTION create_account_attribute()
RETURNS trigger AS $$
BEGIN
    INSERT INTO account_attribute (id)  -- Assume `target_table` has a `source_id` column
    VALUES (NEW.id);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER after_account_insert
AFTER INSERT ON account
FOR EACH ROW
EXECUTE FUNCTION create_account_attribute();