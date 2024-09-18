CREATE TYPE vote_type AS ENUM ('up', 'down', 'none');

CREATE TABLE account (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE solution (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT REFERENCES account(email) ON DELETE CASCADE,
    problem_id BIGINT,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    votes INTEGER DEFAULT 0,
    tags TEXT[],
    UNIQUE (email, id)
);

CREATE TABLE solution_user_vote (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    solution_id BIGINT REFERENCES solution(id) ON DELETE CASCADE NOT NULL,
    email TEXT REFERENCES account(email) ON DELETE CASCADE NOT NULL,
    vote vote_type NOT NULL,
    UNIQUE (email, solution_id)
);

CREATE TABLE comment (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    solution_id BIGINT NOT NULL, 
    email TEXT NOT NULL,
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
    email TEXT REFERENCES account(email) ON DELETE CASCADE,
    vote vote_type NOT NULL,
    UNIQUE (email, comment_id)
);

-- Update the votes count for the solution when a vote is inserted, updated, or deleted
CREATE OR REPLACE FUNCTION update_solution_votes() RETURNS TRIGGER AS $$
BEGIN
    -- Handle Insert case
    IF TG_OP = 'INSERT' THEN
        IF NEW.vote = 'up' THEN
            UPDATE solution SET votes = votes + 1 WHERE id = NEW.solution_id;
        ELSIF NEW.vote = 'down' THEN
            UPDATE solution SET votes = votes - 1 WHERE id = NEW.solution_id;
        END IF;
    
    -- Handle Update case (if the vote is changing)
    ELSIF TG_OP = 'UPDATE' THEN
        -- If the old vote was 'up' and the new vote is 'down'
        IF OLD.vote = 'up' AND NEW.vote = 'down' THEN
            UPDATE solution SET votes = votes - 2 WHERE id = NEW.solution_id;
        -- If the old vote was 'down' and the new vote is 'up'
        ELSIF OLD.vote = 'down' AND NEW.vote = 'up' THEN
            UPDATE solution SET votes = votes + 2 WHERE id = NEW.solution_id;
        -- If the old vote was 'up' and the new vote is 'none' (removing the upvote)
        ELSIF OLD.vote = 'up' AND NEW.vote = 'none' THEN
            UPDATE solution SET votes = votes - 1 WHERE id = NEW.solution_id;
        -- If the old vote was 'down' and the new vote is 'none' (removing the downvote)
        ELSIF OLD.vote = 'down' AND NEW.vote = 'none' THEN
            UPDATE solution SET votes = votes + 1 WHERE id = NEW.solution_id;
        -- If the old vote was 'none' and the new vote is 'up'
        ELSIF OLD.vote = 'none' AND NEW.vote = 'up' THEN
            UPDATE solution SET votes = votes + 1 WHERE id = NEW.solution_id;
        -- If the old vote was 'none' and the new vote is 'down'
        ELSIF OLD.vote = 'none' AND NEW.vote = 'down' THEN
            UPDATE solution SET votes = votes - 1 WHERE id = NEW.solution_id;
        END IF;

    -- Handle Delete case (removing the vote)
    ELSIF TG_OP = 'DELETE' THEN
        IF OLD.vote = 'up' THEN
            UPDATE solution SET votes = votes - 1 WHERE id = OLD.solution_id;
        ELSIF OLD.vote = 'down' THEN
            UPDATE solution SET votes = votes + 1 WHERE id = OLD.solution_id;
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for INSERT
CREATE TRIGGER update_solution_vote_insert
AFTER INSERT ON solution_user_vote
FOR EACH ROW
EXECUTE FUNCTION update_solution_votes();

-- Trigger for UPDATE
CREATE TRIGGER update_solution_vote_update
AFTER UPDATE ON solution_user_vote
FOR EACH ROW
EXECUTE FUNCTION update_solution_votes();

-- Trigger for DELETE
CREATE TRIGGER update_solution_vote_delete
AFTER DELETE ON solution_user_vote
FOR EACH ROW
EXECUTE FUNCTION update_solution_votes();

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