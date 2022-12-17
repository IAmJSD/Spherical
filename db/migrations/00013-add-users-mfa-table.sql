CREATE TABLE users_mfa (
    user_id BIGINT PRIMARY KEY,
    totp_token TEXT
);

ALTER TABLE users_mfa
    ADD CONSTRAINT fk_users_mfa_user_id
    FOREIGN KEY (user_id)
    REFERENCES users (user_id)
    ON DELETE CASCADE
