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