ALTER TABLE files
ADD COLUMN trashed BOOLEAN DEFAULT false,
  ADD COLUMN trashed_at TIMESTAMPTZ;
ALTER TABLE folders
ADD COLUMN trashed BOOLEAN DEFAULT false,
  ADD COLUMN trashed_at TIMESTAMPTZ;
CREATE TABLE IF NOT EXISTS audit_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT REFERENCES users(id) ON DELETE
  SET NULL,
    action TEXT NOT NULL,
    object_type TEXT,
    object_id BIGINT,
    meta JSONB,
    created_at TIMESTAMPTZ DEFAULT now()
);