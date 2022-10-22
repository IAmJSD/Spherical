CREATE TABLE sessions (
    token TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL
);

CREATE INDEX sessions_user_id ON sessions (user_id);
