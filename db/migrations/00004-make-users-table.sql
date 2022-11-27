CREATE TABLE users (
    user_id BIGINT PRIMARY KEY,
    email TEXT NOT NULL,
    confirmed BOOLEAN NOT NULL,
    username TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    last_login_attempt TIMESTAMP,
    updated_at TIMESTAMP
);

ALTER TABLE sessions
    ADD CONSTRAINT fk_session_user_id
    FOREIGN KEY (user_id)
    REFERENCES users (user_id)
    ON DELETE CASCADE
