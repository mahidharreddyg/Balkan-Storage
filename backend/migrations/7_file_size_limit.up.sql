ALTER TABLE files
ADD CONSTRAINT files_size_check CHECK (size > 0);
ALTER TABLE blobs
ADD CONSTRAINT blobs_size_check CHECK (size > 0);