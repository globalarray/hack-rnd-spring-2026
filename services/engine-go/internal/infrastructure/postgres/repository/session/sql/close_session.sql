UPDATE sessions
SET
    status = $1,
    completed_at = NOW()
WHERE id = $2 AND status IN ('CREATED', 'IN_PROGRESS')
