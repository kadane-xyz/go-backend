/* test user */
INSERT into account (username) values ('test');
INSERT into solution (username, problem_id, title, body, tags) values ('test', 1, 'Test title', 'Test body', ARRAY['tag1', 'tag2']);
INSERT into solution (username, problem_id, title, body, tags) values ('test', 1, 'Test title 1', 'Test body 1', ARRAY['tag1']);
INSERT into solution (username, problem_id, title, body, tags) values ('test', 2, 'This is a great statement', 'Spoken well', ARRAY['tag3']);
INSERT into solution (username, problem_id, title, body, tags) values ('test', 3, 'Random stuff', 'Much is said by very little', ARRAY[]);
INSERT into solution_user_vote (username, 1, vote) values ('test', 1, 'up');
INSERT into solution_user_vote (username, 2, vote) values ('test', 1, 'down');
INSERT into solution_user_vote (username, 3, vote) values ('test', 1, 'up');

/* test1 user */
INSERT into account (username) values ('test1');
INSERT into solution (username, problem_id, title, body, tags) values ('test1', 1, 'Test title', 'Test body', ARRAY['tag1', 'tag2']);
INSERT into solution (username, problem_id, title, body, tags) values ('test1', 1, 'Test title 1', 'Test body 1', ARRAY['tag1']);
INSERT into solution (username, problem_id, title, body, tags) values ('test1', 2, 'This is a great statement', 'Spoken well', ARRAY['tag3']);
INSERT into solution (username, problem_id, title, body, tags) values ('test1', 3, 'Random stuff', 'Much is said by very little', ARRAY[]);
INSERT into solution_user_vote (username, 1, vote) values ('test1', 5, 'up');
INSERT into solution_user_vote (username, 2, vote) values ('test1', 6, 'down');
INSERT into solution_user_vote (username, 2, vote) values ('test1', 7, 'up');
INSERT into solution_user_vote (username, 2, vote) values ('test1', 8, 'down');
INSERT into solution_user_vote (username, 3, vote) values ('test1', 1, 'up');


/* test2 user */
INSERT into account (username) values ('test2');
INSERT into solution (username, problem_id, title, body, tags) values ('test2', 1, 'Test title', 'Test body', ARRAY['tag1', 'tag2']);
INSERT into solution (username, problem_id, title, body, tags) values ('test2', 1, 'Test title 1', 'Test body 1', ARRAY['tag1']);
INSERT into solution (username, problem_id, title, body, tags) values ('test2', 2, 'This is a great statement', 'Spoken well', ARRAY['tag3']);
INSERT into solution (username, problem_id, title, body, tags) values ('test2', 3, 'Random stuff', 'Much is said by very little', ARRAY[]);
INSERT into solution_user_vote (username, 1, vote) values ('test2', 9, 'up');
INSERT into solution_user_vote (username, 2, vote) values ('test2', 10, 'down');
INSERT into solution_user_vote (username, 2, vote) values ('test2', 11, 'up');
INSERT into solution_user_vote (username, 2, vote) values ('test2', 1, 'down');
INSERT into solution_user_vote (username, 3, vote) values ('test2', 5, 'up');
