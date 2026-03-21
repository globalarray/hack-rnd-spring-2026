SELECT
    q.id,
    q.survey_id,
    q.order_num,
    q.type,
    q.text,
    q.logic_rules
FROM questions q
         JOIN sessions s ON s.current_question_id = q.id
WHERE s.id = $1