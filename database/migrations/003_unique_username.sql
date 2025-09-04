-- Add unique constraint to users.name column
ALTER TABLE users ADD CONSTRAINT users_name_unique UNIQUE (name);

-- Create index for faster lookups by name
CREATE INDEX idx_users_name ON users(name);
