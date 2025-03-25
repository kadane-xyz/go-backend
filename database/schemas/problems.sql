CREATE TYPE problem_language AS ENUM ('cpp', 'go', 'java', 'javascript', 'python', 'typescript');

CREATE TYPE problem_difficulty AS ENUM ('easy', 'medium', 'hard');

CREATE TYPE problem_sort AS ENUM ('alpha', 'index');

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
    language problem_language NOT NULL,
    code TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (problem_id, language)
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
);