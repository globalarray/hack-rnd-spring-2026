SELECT EXISTS (
    SELECT 1
    FROM sessions
    WHERE survey_id = $1
      AND client_metadata = $2::jsonb
      AND status IN ('CREATED', 'IN_PROGRESS')
);
