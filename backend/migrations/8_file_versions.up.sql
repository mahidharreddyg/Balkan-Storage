CREATE TABLE IF NOT EXISTS audit_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
  action TEXT NOT NULL,
  object_type TEXT NOT NULL,
  object_id TEXT NOT NULL,
  meta JSONB,
  created_at TIMESTAMPTZ DEFAULT now()
);
ALTER TABLE users
ADD COLUMN IF NOT EXISTS role TEXT DEFAULT 'user';
CREATE TABLE IF NOT EXISTS file_versions (
  id BIGSERIAL PRIMARY KEY,
  file_id BIGINT REFERENCES files(id) ON DELETE CASCADE,
  version INT NOT NULL,
  blob_id BIGINT REFERENCES blobs(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ DEFAULT now()
);
ALTER TABLE files
ADD CONSTRAINT filename_length CHECK (char_length(filename) <= 255);
ALTER TABLE folders
ADD CONSTRAINT foldername_length CHECK (char_length(name) <= 255);
ALTER TABLE users
ADD CONSTRAINT username_length CHECK (
    char_length(username) BETWEEN 3 AND 64
  );
ALTER TABLE users
ADD CONSTRAINT email_length CHECK (char_length(email) <= 255);