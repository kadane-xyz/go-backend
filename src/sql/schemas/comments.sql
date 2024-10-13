CREATE TABLE comment (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    solution_id BIGINT NOT NULL, 
    username TEXT NOT NULL,
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
    username TEXT REFERENCES account(username) ON DELETE CASCADE,
    vote vote_type NOT NULL,
    UNIQUE (username, comment_id)
);

-- Update the votes count for the comment when a vote is inserted, updated, or deleted
CREATE OR REPLACE FUNCTION update_comment_votes() RETURNS TRIGGER AS $$
BEGIN
    -- Handle Insert case
    IF TG_OP = 'INSERT' THEN
        IF NEW.vote = 'up' THEN
            UPDATE comment SET votes = votes + 1 WHERE id = NEW.comment_id;
        ELSIF NEW.vote = 'down' THEN
            UPDATE comment SET votes = votes - 1 WHERE id = NEW.comment_id;
        END IF;

    -- Handle Update case (if the vote is changing)
    ELSIF TG_OP = 'UPDATE' THEN
        -- If the old vote was 'up' and the new vote is 'down'
        IF OLD.vote = 'up' AND NEW.vote = 'down' THEN
            UPDATE comment SET votes = votes - 2 WHERE id = NEW.comment_id;
        -- If the old vote was 'down' and the new vote is 'up'
        ELSIF OLD.vote = 'down' AND NEW.vote = 'up' THEN
            UPDATE comment SET votes = votes + 2 WHERE id = NEW.comment_id;
        -- If the old vote was 'up' and the new vote is 'none' (removing the upvote)
        ELSIF OLD.vote = 'up' AND NEW.vote = 'none' THEN
            UPDATE comment SET votes = votes - 1 WHERE id = NEW.comment_id;
        -- If the old vote was 'down' and the new vote is 'none' (removing the downvote)
        ELSIF OLD.vote = 'down' AND NEW.vote = 'none' THEN
            UPDATE comment SET votes = votes + 1 WHERE id = NEW.comment_id;
        -- If the old vote was 'none' and the new vote is 'up'
        ELSIF OLD.vote = 'none' AND NEW.vote = 'up' THEN
            UPDATE comment SET votes = votes + 1 WHERE id = NEW.comment_id;
        -- If the old vote was 'none' and the new vote is 'down'
        ELSIF OLD.vote = 'none' AND NEW.vote = 'down' THEN
            UPDATE comment SET votes = votes - 1 WHERE id = NEW.comment_id;
        END IF;

    -- Handle Delete case (removing the vote)
    ELSIF TG_OP = 'DELETE' THEN
        IF OLD.vote = 'up' THEN
            UPDATE comment SET votes = votes - 1 WHERE id = OLD.comment_id;
        ELSIF OLD.vote = 'down' THEN
            UPDATE comment SET votes = votes + 1 WHERE id = OLD.comment_id;
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for INSERT
CREATE TRIGGER update_comment_vote_insert
AFTER INSERT ON comment_user_vote
FOR EACH ROW
EXECUTE FUNCTION update_comment_votes();

-- Trigger for UPDATE
CREATE TRIGGER update_comment_vote_update
AFTER UPDATE ON comment_user_vote
FOR EACH ROW
EXECUTE FUNCTION update_comment_votes();

-- Trigger for DELETE
CREATE TRIGGER update_comment_vote_delete
AFTER DELETE ON comment_user_vote
FOR EACH ROW
EXECUTE FUNCTION update_comment_votes();

CREATE OR REPLACE FUNCTION get_comments_sorted(
    p_solution_id BIGINT,
    p_order_by TEXT DEFAULT 'created_at',
    p_sort_direction TEXT DEFAULT 'DESC'
) 
RETURNS SETOF comment AS $$
BEGIN
    RETURN QUERY EXECUTE format(
        'SELECT * FROM comment
         WHERE solution_id = $1
         ORDER BY %s %s',
        -- Only allow sorting by 'votes' or 'created_at'
        CASE WHEN p_order_by = 'votes' THEN 'votes'
             WHEN p_order_by = 'created_at' THEN 'created_at'
             ELSE 'created_at'  -- Default to 'created_at' if invalid input
        END,
        -- Only allow 'ASC' or 'DESC' for sort direction
        CASE WHEN UPPER(p_sort_direction) = 'DESC' THEN 'DESC' ELSE 'ASC' END
    ) USING p_solution_id;
END;
$$ LANGUAGE plpgsql;