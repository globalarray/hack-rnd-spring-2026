UPDATE sessions
SET current_question_id = $1,
    status = $2,
    completed_at = CASE WHEN $3 = TRUE THEN NOW() ELSE NULL END
WHERE id = $4
