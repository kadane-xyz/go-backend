-- account vote type --
CREATE TYPE vote_type AS ENUM ('up', 'down', 'none');

-- visibility type --
CREATE TYPE visibility AS ENUM ('public', 'private');
CREATE TABLE account (
    id TEXT NOT NULL UNIQUE PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    avatar_url TEXT,
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
);CREATE TYPE friendship_status AS ENUM ('pending', 'accepted', 'blocked');

-- First, ensure the friendship table is properly structured
CREATE TABLE IF NOT EXISTS friendship (
    user_id_1 TEXT REFERENCES account(id),
    user_id_2 TEXT REFERENCES account(id),
    initiator_id TEXT REFERENCES account(id), -- the user who initiated the friend request
    status friendship_status NOT NULL DEFAULT 'pending',
    accepted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- default to now and update on friend request change
    PRIMARY KEY (user_id_1, user_id_2),
    CONSTRAINT friendship_members_ordered CHECK (user_id_1 < user_id_2)
);CREATE TABLE solution (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id TEXT REFERENCES account(id) ON DELETE CASCADE,
    problem_id BIGINT,
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
EXECUTE FUNCTION update_comment_votes();CREATE TYPE problem_language AS ENUM ('cpp', 'go', 'java', 'javascript', 'python', 'typescript');

CREATE TYPE problem_difficulty AS ENUM ('easy', 'medium', 'hard');

CREATE TABLE problem (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    function_name TEXT NOT NULL,
    points INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    difficulty problem_difficulty NOT NULL,
    tags TEXT[],
    UNIQUE (id, title)
);

CREATE TABLE problem_solution (
    id SERIAL PRIMARY KEY,
    problem_id INT REFERENCES problem(id) ON DELETE CASCADE,
    solution TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (problem_id, id)
);

CREATE TABLE problem_code (
    id SERIAL PRIMARY KEY,
    problem_id INT REFERENCES problem(id) ON DELETE CASCADE,
    language problem_language NOT NULL,
    code TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (problem_id, id)
);

CREATE TABLE problem_hint (
    id SERIAL PRIMARY KEY,
    problem_id INT REFERENCES problem(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    answer TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (problem_id, id)
);

CREATE TABLE problem_test_case (
    id SERIAL PRIMARY KEY,
    problem_id INT REFERENCES problem(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    visibility visibility NOT NULL,
    UNIQUE (problem_id, id)
);

CREATE TYPE problem_test_case_type AS ENUM (
    'int',
    'int[]',
    'string',
    'string[]',
    'float',
    'float[]',
    'double',
    'double[]',
    'boolean',
    'boolean[]',
    'null'
);

CREATE TABLE problem_test_case_input (
    id SERIAL PRIMARY KEY,
    problem_test_case_id INT REFERENCES problem_test_case(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type problem_test_case_type NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (problem_test_case_id, id)
);

CREATE TABLE problem_test_case_output (
    id SERIAL PRIMARY KEY,
    problem_test_case_id INT REFERENCES problem_test_case(id) ON DELETE CASCADE,
    value TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (problem_test_case_id, id)
);CREATE TABLE account_solved_problem (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id TEXT REFERENCES account(id) ON DELETE CASCADE,
    problem_id BIGINT REFERENCES problem(id) ON DELETE CASCADE,
    solved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);CREATE TYPE submission_status AS ENUM (
    'In Queue',                    -- 1
    'Processing',                  -- 2
    'Accepted',                    -- 3
    'Wrong Answer',                -- 4
    'Time Limit Exceeded',         -- 5
    'Compilation Error',           -- 6
    'Runtime Error (SIGSEGV)',     -- 7  Segmentation fault
    'Runtime Error (SIGXFSZ)',     -- 8  File size limit exceeded
    'Runtime Error (SIGFPE)',      -- 9  Floating point error
    'Runtime Error (SIGABRT)',     -- 10 Abort signal
    'Runtime Error (NZEC)',        -- 11 Non-zero exit code
    'Runtime Error (Other)',       -- 12 Other runtime errors
    'Internal Error',              -- 13 Judge0 internal error
    'Exec Format Error'            -- 14
);

CREATE TABLE submission (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stdout TEXT,
    time TEXT,
    memory INTEGER,
    stderr TEXT,
    compile_output TEXT,
    message TEXT,
    status submission_status NOT NULL,
    language_id INTEGER NOT NULL,
    language_name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    account_id TEXT NOT NULL REFERENCES account(id) ON DELETE CASCADE,
    submitted_code TEXT NOT NULL,
    submitted_stdin TEXT DEFAULT NULL,
    problem_id INTEGER NOT NULL REFERENCES problem(id) ON DELETE CASCADE
);CREATE TABLE starred_problem (
    id             SERIAL PRIMARY KEY,
    user_id        TEXT NOT NULL,
    problem_id     INT NOT NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES account (id),
    FOREIGN KEY (problem_id) REFERENCES problem (id),
    UNIQUE (user_id, problem_id)
);

CREATE TABLE starred_solution (
    id             SERIAL PRIMARY KEY,
    user_id        TEXT NOT NULL,
    solution_id    INT NOT NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES account (id),
    FOREIGN KEY (solution_id) REFERENCES solution (id),
    UNIQUE (user_id, solution_id)
);

CREATE TABLE starred_submission (
    id             SERIAL PRIMARY KEY,
    user_id        TEXT NOT NULL,
    submission_id  UUID NOT NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES account (id),
    FOREIGN KEY (submission_id) REFERENCES submission (id),
    UNIQUE (user_id, submission_id)
);
-- Insert test accounts
INSERT INTO account (id, username, email, avatar_url) VALUES
('123abc', 'johndoe', 'john@example.com', 'https://example.com/avatars/johndoe.jpg'),
('456def', 'janesmith', 'jane@example.com', 'https://example.com/avatars/janesmith.jpg'),
('789ghi', 'bobjohnson', 'bob@example.com', 'https://example.com/avatars/bobjohnson.jpg');

INSERT INTO account_attribute (id, bio, contact_email, location, real_name, github_url, linkedin_url, facebook_url, instagram_url, twitter_url, school, website_url) VALUES
('123abc', 'Passionate coder', 'john@example.com', 'New York', 'John Doe', 'https://github.com/johndoe', 'https://linkedin.com/in/johndoe', 'https://facebook.com/johndoe', 'https://instagram.com/johndoe', 'https://twitter.com/johndoe', 'MIT', 'https://www.johndoe.com'),
('456def', 'AI enthusiast', 'jane@example.com', 'San Francisco', 'Jane Smith', 'https://github.com/janesmith', 'https://linkedin.com/in/janesmith', 'https://facebook.com/janesmith', 'https://instagram.com/janesmith', 'https://twitter.com/janesmith', 'Stanford', 'https://www.janesmith.com'),
('789ghi', 'Full-stack developer', 'bob@example.com', 'London', 'Bob Johnson', 'https://github.com/bobjohnson', 'https://linkedin.com/in/bobjohnson', 'https://facebook.com/bobjohnson', 'https://instagram.com/bobjohnson', 'https://twitter.com/bobjohnson', 'Oxford', 'https://www.bobjohnson.com');

-- Insert problem
INSERT INTO problem (title, description, function_name, points, difficulty, tags) VALUES
('Two Sum', 'Given an array of integers nums and an integer target, return indices of the two numbers such that they add up to target.', 'twoSum', 10, 'easy', ARRAY['array', 'hash table']),
('Reverse Linked List', 'Given the head of a singly linked list, reverse the list, and return the reversed list.', 'reverseList', 10, 'medium', ARRAY['linked list', 'iterative']),
('Merge k Sorted Lists', 'You are given an array of k linked-lists lists, each linked-list is sorted in ascending order. Merge all the linked-lists into one sorted linked-list and return it.', 'mergeKLists', 10, 'hard', ARRAY['linked list', 'heap', 'divide and conquer']);

-- Insert problem_solution
INSERT INTO problem_solution (problem_id, solution) VALUES
(1, 'Efficient Two Sum Solution'),
(2, 'Recursive Reverse Linked List'),
(3, 'Merge K Sorted Lists using Priority Queue');

-- Insert problem_code
INSERT INTO problem_code (problem_id, language, code) VALUES
(1, 'python', 'def twoSum(nums, target):'),
(2, 'python', 'def reverseList(head):'),
(3, 'python', 'def mergeKLists(lists):');

-- Insert problem_hint
INSERT INTO problem_hint (problem_id, description, answer) VALUES
(1, 'Use a hash map to store the indices of the numbers.', ''),
(2, 'Use a recursive function to reverse the linked list.', ''),
(3, 'Use a priority queue to merge the linked lists.', '');

-- Insert problem_test_case
INSERT INTO problem_test_case (problem_id, description, visibility) VALUES
(1, 'Test case 1', 'public'),
(1, 'Test case 2', 'public'),
(2, 'Test case 1', 'public'),
(2, 'Test case 2', 'public'),
(3, 'Test case 1', 'public'),
(3, 'Test case 2', 'public');

-- Insert problem_test_case_input with correct enum types
INSERT INTO problem_test_case_input (problem_test_case_id, name, type, value) VALUES
(1, 'nums', 'int[]', '[2, 7, 11, 15]'),
(1, 'target', 'int', '9'),
(2, 'head', 'string', '1 -> 2 -> 3 -> 4 -> 5'),
(3, 'lists', 'string[]', '[1 -> 4 -> 5, 1 -> 3 -> 4, 2 -> 6]');

-- Insert problem_test_case_output
INSERT INTO problem_test_case_output (problem_test_case_id, value) VALUES
(1, '[0, 1]'),
(2, '5 -> 4 -> 3 -> 2 -> 1'),
(3, '1 -> 1 -> 2 -> 3 -> 4 -> 4 -> 5 -> 6');

-- Insert test solutions
INSERT INTO solution (user_id, problem_id, title, body, tags) VALUES
('123abc', 1, 'Efficient Two Sum Solution', 'Here''s an O(n) solution using a hash map...', ARRAY['array', 'hash table']),
('456def', 1, 'Two-pointer approach for Two Sum', 'We can solve this using two pointers...', ARRAY['array', 'two pointers']),
('789ghi', 2, 'Iterative Reverse Linked List', 'An iterative approach to reverse a linked list...', ARRAY['linked list', 'iterative']),
('123abc', 2, 'Recursive Reverse Linked List', 'Here''s a recursive solution to reverse a linked list...', ARRAY['linked list', 'recursive']),
('456def', 3, 'Merge K Sorted Lists using Priority Queue', 'We can use a priority queue to efficiently merge...', ARRAY['linked list', 'heap', 'divide and conquer']);

-- Insert test solution votes
INSERT INTO solution_user_vote (user_id, solution_id, vote) VALUES
('456def', 1, 'up'),
('789ghi', 1, 'up'),
('123abc', 2, 'up'),
('789ghi', 2, 'down'),
('456def', 3, 'up'),
('123abc', 4, 'up'),
('789ghi', 5, 'up');

-- Insert test comments
INSERT INTO comment (solution_id, user_id, body) VALUES
(1, '456def', 'Great solution! Very efficient.'),
(1, '789ghi', 'I like the use of a hash map here.'),
(2, '123abc', 'Interesting approach with two pointers.'),
(3, '456def', 'Clear explanation of the iterative method.'),
(4, '789ghi', 'Recursive solutions can be tricky, but this is well-explained.'),
(5, '123abc', 'Excellent use of a priority queue for this problem.');

-- Insert test comment votes
INSERT INTO comment_user_vote (user_id, comment_id, vote) VALUES
('123abc', 1, 'up'),
('789ghi', 1, 'up'),
('123abc', 2, 'up'),
('456def', 3, 'up'),
('789ghi', 4, 'up'),
('456def', 5, 'up'),
('123abc', 6, 'up');

-- Insert test solved problems
INSERT INTO account_solved_problem (user_id, problem_id) VALUES
('123abc', 1),
('123abc', 2),
('456def', 1),
('456def', 3),
('789ghi', 2);

-- Insert test game stats
INSERT INTO account_game_stat (user_id, wins, losses, elo) VALUES
('123abc', 10, 5, 1200),
('456def', 15, 3, 1350),
('789ghi', 8, 7, 1100);

-- Insert test friends (ensuring unique pairs and user_id_1 < user_id_2)
INSERT INTO friendship (user_id_1, user_id_2, initiator_id, status) VALUES
    ('123abc', '456def', '123abc', 'accepted'),  -- Friendship between John and Jane
    ('123abc', '789ghi', '123abc', 'pending'),   -- Friendship between John and Bob
    ('456def', '789ghi', '456def', 'pending');   -- Friendship between Jane and Bob

-- Insert test starred problems
INSERT INTO starred_problem (user_id, problem_id) VALUES
('123abc', 1),
('123abc', 2),
('456def', 1),
('456def', 3),
('789ghi', 2);

-- Insert test starred solutions
INSERT INTO starred_solution (user_id, solution_id) VALUES
('123abc', 1),
('123abc', 2),
('456def', 3),
('456def', 4),
('789ghi', 5);