SELECT
    s.survey_id,
    s.id AS session_id,
    s.client_metadata::text AS client_metadata_json
FROM sessions s
WHERE s.id = $1;
