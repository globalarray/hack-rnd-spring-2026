SELECT
    r.question_id,
    q.type AS question_type,
    q.text AS question_text,
    COALESCE(a.weight, 0) AS selected_weight,
    COALESCE(a.category_tag, '') AS category_tag,
    COALESCE(r.raw_text, '') AS raw_text
FROM responses r
JOIN questions q ON q.id = r.question_id
LEFT JOIN answers a ON a.id = r.answer_id
WHERE r.session_id = $1
ORDER BY r.created_at, r.id;
