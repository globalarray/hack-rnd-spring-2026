SELECT EXISTS (
    SELECT 1
    FROM sessions
    WHERE survey_id = $1
      AND client_metadata ->> '__shareLinkId' = $2
      AND status = 'COMPLETED'
);
