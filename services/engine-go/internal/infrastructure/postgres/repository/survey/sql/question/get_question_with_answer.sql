SELECT
    q.id,
    q.order_num,
    q.type,
    q.text,
    q.logic_rules,
    a.id AS answer_id,
    a.text AS answer_text,
    a.weight,
    COALESCE(a.category_tag, '') AS category_tag
FROM questions q
JOIN answers a ON a.question_id = q.id
WHERE q.id = $1 AND a.id = $2
