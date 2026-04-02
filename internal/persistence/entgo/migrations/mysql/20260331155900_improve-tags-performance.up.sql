-- for fast lookup by job_id
CREATE INDEX idx_tag_jobs_job_id ON tag_jobs(job_id);
