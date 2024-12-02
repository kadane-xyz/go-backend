CREATE TYPE problem_language AS ENUM ('cpp', 'go', 'java', 'javascript', 'python', 'typescript');

CREATE TABLE problem (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT,
    points INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    tags TEXT[],
    UNIQUE (id, title)
);

CREATE TABLE problem_code (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    problem_id UUID REFERENCES problem(id) ON DELETE CASCADE,
    language problem_language NOT NULL,
    code BYTEA NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (problem_id, id)
);

CREATE TABLE problem_hint (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    problem_id UUID REFERENCES problem(id) ON DELETE CASCADE,
    description BYTEA NOT NULL,
    answer BYTEA NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (problem_id, id)
);

CREATE TABLE problem_test_case (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    problem_id UUID REFERENCES problem(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    input TEXT NOT NULL,
    expected_output TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (problem_id, id)
);

CREATE TABLE problem_solution (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    problem_id UUID REFERENCES problem(id) ON DELETE CASCADE,
    expected_output BYTEA NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (problem_id, id)
);