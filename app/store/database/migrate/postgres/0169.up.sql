-- Covers: WHERE job_state = 'scheduled' AND job_scheduled <= ? ORDER BY job_priority DESC, job_scheduled ASC
CREATE INDEX jobs_scheduled_priority
    ON jobs(job_priority DESC, job_scheduled ASC)
    WHERE job_state = 'scheduled';
