INSERT INTO sessions (survey_id, current_question, status, client_metadata)
VALUES ($1, $2, $3, $3)
    RETURNING id;