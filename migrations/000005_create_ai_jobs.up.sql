CREATE TYPE job_status AS ENUM ('pending', 'running', 'completed', 'failed');

CREATE TABLE ai_jobs (
    id           BIGSERIAL PRIMARY KEY,
    issue_id     BIGINT NOT NULL REFERENCES issues(id),
    status       job_status NOT NULL DEFAULT 'pending',
    attempts     INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    started_at   TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_msg    TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_jobs_pending ON ai_jobs (status, created_at) WHERE status = 'pending';
