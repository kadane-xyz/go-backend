CREATE TYPE submission_status AS ENUM ('pending', 'running', 'completed', 'failed');

CREATE TABLE submission (
    id SERIAL PRIMARY KEY,
    stdout TEXT,
    execution_time DECIMAL(10,3),
    memory_used INTEGER,
    stderr TEXT,
    token VARCHAR(36),
    compile_output TEXT,
    message TEXT,
    status submission_status,
    status_id INTEGER,
    status_description VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    account_id TEXT REFERENCES account(id),
    problem_id INTEGER REFERENCES problem(id)
);

-- Add indexes for commonly queried fields
CREATE INDEX idx_submission_token ON submission(token);
CREATE INDEX idx_submission_status_id ON submission(status_id);