-- Insert test accounts
INSERT INTO account_attributes (realName, bio, location, github_url, linkedin_url, school) VALUES
('John Doe', 'Passionate coder', 'New York', 'https://github.com/johndoe', 'https://linkedin.com/in/johndoe', 'MIT'),
('Jane Smith', 'AI enthusiast', 'San Francisco', 'https://github.com/janesmith', 'https://linkedin.com/in/janesmith', 'Stanford'),
('Bob Johnson', 'Full-stack developer', 'London', 'https://github.com/bobjohnson', 'https://linkedin.com/in/bobjohnson', 'Oxford');

INSERT INTO account (username, email, verified, avatar_url, attributesId) VALUES
('johndoe', 'john@example.com', TRUE, 'https://example.com/avatars/johndoe.jpg', 1),
('janesmith', 'jane@example.com', TRUE, 'https://example.com/avatars/janesmith.jpg', 2),
('bobjohnson', 'bob@example.com', FALSE, 'https://example.com/avatars/bobjohnson.jpg', 3);

-- Insert test problems (assuming you have a problems table)
INSERT INTO problems (title, description, difficulty) VALUES
('Two Sum', 'Find two numbers in an array that add up to a target', 'Easy'),
('Reverse Linked List', 'Reverse a singly linked list', 'Medium'),
('Merge K Sorted Lists', 'Merge k sorted linked lists into one sorted list', 'Hard');

-- Insert test solutions
INSERT INTO solution (username, problem_id, title, body, tags) VALUES
('johndoe', 1, 'Efficient Two Sum Solution', 'Here''s an O(n) solution using a hash map...', ARRAY['array', 'hash table']),
('janesmith', 1, 'Two-pointer approach for Two Sum', 'We can solve this using two pointers...', ARRAY['array', 'two pointers']),
('bobjohnson', 2, 'Iterative Reverse Linked List', 'An iterative approach to reverse a linked list...', ARRAY['linked list', 'iterative']),
('johndoe', 2, 'Recursive Reverse Linked List', 'Here''s a recursive solution to reverse a linked list...', ARRAY['linked list', 'recursive']),
('janesmith', 3, 'Merge K Sorted Lists using Priority Queue', 'We can use a priority queue to efficiently merge...', ARRAY['linked list', 'heap', 'divide and conquer']);

-- Insert test solution votes
INSERT INTO solution_user_vote (username, solution_id, vote) VALUES
('janesmith', 1, 'up'),
('bobjohnson', 1, 'up'),
('johndoe', 2, 'up'),
('bobjohnson', 2, 'down'),
('johndoe', 3, 'up'),
('janesmith', 4, 'up'),
('bobjohnson', 5, 'up');

-- Insert test comments
INSERT INTO comment (solution_id, username, body) VALUES
(1, 'janesmith', 'Great solution! Very efficient.'),
(1, 'bobjohnson', 'I like the use of a hash map here.'),
(2, 'johndoe', 'Interesting approach with two pointers.'),
(3, 'janesmith', 'Clear explanation of the iterative method.'),
(4, 'bobjohnson', 'Recursive solutions can be tricky, but this is well-explained.'),
(5, 'johndoe', 'Excellent use of a priority queue for this problem.');

-- Insert test comment votes
INSERT INTO comment_user_vote (username, comment_id, vote) VALUES
('johndoe', 1, 'up'),
('bobjohnson', 1, 'up'),
('johndoe', 2, 'up'),
('janesmith', 3, 'up'),
('bobjohnson', 4, 'up'),
('janesmith', 5, 'up'),
('johndoe', 6, 'up');

-- Insert test solved problems
INSERT INTO account_solved_problems (username, problem_id) VALUES
('johndoe', 1),
('johndoe', 2),
('janesmith', 1),
('janesmith', 3),
('bobjohnson', 2);

-- Insert test game stats
INSERT INTO account_game_stats (username, wins, losses, elo) VALUES
('johndoe', 10, 5, 1200),
('janesmith', 15, 3, 1350),
('bobjohnson', 8, 7, 1100);