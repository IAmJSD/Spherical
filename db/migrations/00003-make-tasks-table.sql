CREATE TABLE tasks (
    task_id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    job_handler TEXT NOT NULL,
    scheduled_for TIMESTAMP NOT NULL,
    job_body BYTEA NOT NULL,
    scheduler_metadata BYTEA NOT NULL
);

CREATE INDEX tasks_job_scheduled ON tasks (job_handler, scheduled_for ASC)
