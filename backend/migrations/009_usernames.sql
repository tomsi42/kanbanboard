-- Add username field to users (optional, for agent accounts)
ALTER TABLE users ADD COLUMN username TEXT;

-- Case-insensitive unique index: only enforced when username is not null
CREATE UNIQUE INDEX users_username_key ON users(lower(username)) WHERE username IS NOT NULL;

-- Make email nullable: agents may not have an email address
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;
