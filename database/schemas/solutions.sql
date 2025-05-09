CREATE TABLE solution (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id TEXT REFERENCES account(id) ON DELETE CASCADE,
    problem_id INT,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    votes INTEGER DEFAULT 0,
    tags TEXT[],
    UNIQUE (user_id, id)
);

CREATE TABLE solution_user_vote (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    solution_id BIGINT REFERENCES solution(id) ON DELETE CASCADE NOT NULL,
    user_id TEXT REFERENCES account(id) ON DELETE CASCADE NOT NULL,
    vote vote_type NOT NULL,
    UNIQUE (user_id, solution_id)
);

-- Function to recalculate a solution's vote total based on the current votes.
CREATE OR REPLACE FUNCTION recalc_solution_votes(p_solution_id BIGINT)
RETURNS VOID AS $$
BEGIN
    UPDATE solution
    SET votes = (
        SELECT COALESCE(SUM(
            CASE vote
                WHEN 'up' THEN 1
                WHEN 'down' THEN -1
                ELSE 0
            END
        ), 0)
        FROM solution_user_vote
        WHERE solution_id = p_solution_id
    )
    WHERE id = p_solution_id;
END;
$$ LANGUAGE plpgsql;

-- Function that encapsulates vote logic:
-- If p_vote is 'none', any existing vote is removed.
-- Otherwise, an upsert is performed to insert or update the vote.
-- Finally, the comment's vote total is recalculated.
CREATE OR REPLACE FUNCTION set_solution_vote(
    p_user_id TEXT,
    p_solution_id BIGINT,
    p_vote vote_type
) RETURNS VOID AS $$
BEGIN
    IF p_vote = 'none' THEN
        DELETE FROM solution_user_vote
        WHERE user_id = p_user_id
          AND solution_id = p_solution_id;
    ELSE
        INSERT INTO solution_user_vote (user_id, solution_id, vote)
        VALUES (p_user_id, p_solution_id, p_vote)
        ON CONFLICT (user_id, solution_id)
        DO UPDATE SET vote = EXCLUDED.vote;
    END IF;

    PERFORM recalc_solution_votes(p_solution_id);
END;
$$ LANGUAGE plpgsql;