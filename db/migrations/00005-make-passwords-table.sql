CREATE TABLE password_authentication_users (
    user_id BIGINT PRIMARY KEY,
    password TEXT NOT NULL,
    CONSTRAINT fk_passwords_user
        FOREIGN KEY (user_id)
            REFERENCES users (user_id)
            ON DELETE CASCADE
)
