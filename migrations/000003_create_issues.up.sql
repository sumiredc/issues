CREATE TYPE issue_status AS ENUM ('open', 'in_progress', 'completed', 'closed');

CREATE TABLE issues (
    id            BIGSERIAL PRIMARY KEY,
    project_id    BIGINT NOT NULL REFERENCES projects(id),
    title         TEXT NOT NULL,
    body          TEXT,
    status        issue_status NOT NULL DEFAULT 'open',
    ai_session_id TEXT,
    ai_result     TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_issues_project_id ON issues (project_id);
CREATE INDEX idx_issues_status ON issues (project_id, status);
