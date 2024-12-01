CREATE TYPE submission_status AS ENUM (
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
    token VARCHAR(36) NOT NULL,
    stdout TEXT,
    time TEXT,
    memory_used INTEGER,
    stderr TEXT,
    compile_output TEXT,
    message TEXT,
    status submission_status NOT NULL,
    status_id INTEGER NOT NULL,
    status_description TEXT NOT NULL,
    language_id INTEGER NOT NULL,
    language_name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    account_id TEXT NOT NULL REFERENCES account(id) ON DELETE CASCADE,
    problem_id UUID NOT NULL REFERENCES problem(id) ON DELETE CASCADE,
    UNIQUE(token)
);

-- Add indexes for commonly queried fields
CREATE INDEX idx_submission_token ON submission(token);
CREATE INDEX idx_submission_status_id ON submission(status_id);