SELECT
    s.survey_id,
    s.id AS session_id,
    s.client_metadata::text AS client_metadata_json,
    s.status::text AS status,
    COUNT(r.id)::int AS responses_count,
    s.started_at,
    s.completed_at
FROM sessions s
LEFT JOIN responses r ON r.session_id = s.id
WHERE s.survey_id = $1
  AND s.status = 'COMPLETED'
GROUP BY s.survey_id, s.id, s.client_metadata, s.status, s.started_at, s.completed_at
ORDER BY s.completed_at DESC NULLS LAST, s.started_at DESC, s.id DESC;
