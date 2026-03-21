INSERT INTO sessions (survey_id, current_question_id, status, client_metadata)
VALUES (
    $1,
    (
        SELECT q.id
        FROM questions q
        WHERE q.survey_id = $1
        ORDER BY q.order_num
        LIMIT 1
    ),
    $2,
    $3::jsonb
)
RETURNING id;
