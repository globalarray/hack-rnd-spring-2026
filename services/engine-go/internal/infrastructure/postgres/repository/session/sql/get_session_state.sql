SELECT
    s.id,
    s.survey_id,
    s.status,
    s.started_at,
    s.current_question_id,
    sur.settings
FROM sessions s
         JOIN surveys sur ON s.survey_id = sur.id
WHERE s.id = $1