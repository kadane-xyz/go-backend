CREATE TYPE problem_language AS ENUM ('cpp', 'go', 'java', 'javascript', 'python', 'typescript');

CREATE TYPE problem_difficulty AS ENUM ('easy', 'medium', 'hard');

CREATE TABLE problem (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
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
    input TEXT NOT NULL,
    output TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    visibility visibility NOT NULL,
    UNIQUE (problem_id, id)
);