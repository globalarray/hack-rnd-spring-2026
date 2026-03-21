INSERT INTO surveys (psychologist_id, title, description, settings)
VALUES ($1, $2, $3, $4)
    RETURNING id;