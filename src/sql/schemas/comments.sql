CREATE TABLE comment (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    solution_id BIGINT NOT NULL, 
    user_id TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    votes INTEGER DEFAULT 0,
    parent_id BIGINT NULL,
    UNIQUE (id, solution_id),
    FOREIGN KEY (parent_id, solution_id)
        REFERENCES comment (id, solution_id)
        ON DELETE CASCADE
);

CREATE TABLE comment_user_vote (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    comment_id BIGINT REFERENCES comment(id) ON DELETE CASCADE,
    user_id TEXT REFERENCES account(id) ON DELETE CASCADE,
    vote vote_type NOT NULL,
    UNIQUE (user_id, comment_id)
);

-- Function to recalculate a comment's vote total based on the current votes.
CREATE OR REPLACE FUNCTION recalc_comment_votes(p_comment_id BIGINT)
RETURNS VOID AS $$
BEGIN
    UPDATE comment
    SET votes = (
        SELECT COALESCE(SUM(
            CASE vote
                WHEN 'up' THEN 1
                WHEN 'down' THEN -1
                ELSE 0
            END
        ), 0)
        FROM comment_user_vote
        WHERE comment_id = p_comment_id
    )
    WHERE id = p_comment_id;
END;
$$ LANGUAGE plpgsql;

-- Function that encapsulates vote logic:
-- If p_vote is 'none', any existing vote is removed.
-- Otherwise, an upsert is performed to insert or update the vote.
-- Finally, the comment's vote total is recalculated.
CREATE OR REPLACE FUNCTION set_comment_vote(
    p_user_id TEXT,
    p_comment_id BIGINT,
    p_vote vote_type
) RETURNS VOID AS $$
BEGIN
    IF p_vote = 'none' THEN
        DELETE FROM comment_user_vote
        WHERE user_id = p_user_id
          AND comment_id = p_comment_id;
    ELSE
        INSERT INTO comment_user_vote (user_id, comment_id, vote)
        VALUES (p_user_id, p_comment_id, p_vote)
        ON CONFLICT (user_id, comment_id)
        DO UPDATE SET vote = EXCLUDED.vote;
    END IF;

    PERFORM recalc_comment_votes(p_comment_id);
END;
$$ LANGUAGE plpgsql;