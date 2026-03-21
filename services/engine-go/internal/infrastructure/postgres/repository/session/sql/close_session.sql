UPDATE sessions
SET
    status = $1,
    completed_at = NOW(),
    total_score = (SELECT SUM(weight) FROM responses WHERE session_id = $2)
WHERE id = $2 AND status = 1