SELECT
    s.id,
    s.survey_id,
    s.status,
    s.started_at,
    s.current_question_id,
    s.client_metadata,
    s.completed_at,
    sur.settings
FROM sessions s
         JOIN surveys sur ON s.survey_id = sur.id
WHERE s.id = $1
