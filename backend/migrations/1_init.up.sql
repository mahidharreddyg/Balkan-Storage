-- Users table
CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  username TEXT UNIQUE NOT NULL,
  email TEXT UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT DEFAULT 'user',
  quota BIGINT DEFAULT 10485760,
  created_at TIMESTAMPTZ DEFAULT now()
);
-- Blobs table
CREATE TABLE IF NOT EXISTS blobs (
  id BIGSERIAL PRIMARY KEY,
  hash TEXT UNIQUE NOT NULL,
  size BIGINT NOT NULL,
  path TEXT NOT NULL,
  ref_count INT DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT now()
);
-- Files table
CREATE TABLE IF NOT EXISTS files (
  id BIGSERIAL PRIMARY KEY,
  blob_id BIGINT REFERENCES blobs(id) ON DELETE CASCADE,
  owner_id BIGINT REFERENCES users(id),
  filename TEXT NOT NULL,
  mime_type TEXT,
  size BIGINT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now(),
  download_count BIGINT DEFAULT 0,
  is_public BOOLEAN DEFAULT false
);
-- Shares table (optional)
CREATE TABLE IF NOT EXISTS shares (
  id BIGSERIAL PRIMARY KEY,
  file_id BIGINT REFERENCES files(id) ON DELETE CASCADE,
  token TEXT UNIQUE NOT NULL,
  expires_at TIMESTAMPTZ
);