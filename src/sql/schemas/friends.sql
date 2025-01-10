CREATE TYPE friendship_status AS ENUM ('pending', 'accepted', 'blocked');

-- First, ensure the friendship table is properly structured
CREATE TABLE IF NOT EXISTS friendship (
    user_id_1 TEXT REFERENCES account(id),
    user_id_2 TEXT REFERENCES account(id),
    status friendship_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id_1, user_id_2),
    CONSTRAINT friendship_members_ordered CHECK (user_id_1 < user_id_2)
);