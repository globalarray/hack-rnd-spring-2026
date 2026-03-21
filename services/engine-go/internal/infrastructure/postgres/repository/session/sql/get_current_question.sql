SELECT
    q.id,
    q.order_num,
    q.type,
    q.text,
    q.logic_rules,
    COALESCE(
        (
            SELECT jsonb_agg(
                jsonb_build_object(
                    'id', a.id,
                    'text', a.text,
                    'weight', a.weight,
                    'category_tag', a.category_tag
                )
                ORDER BY a.id
            )
            FROM answers a
            WHERE a.question_id = q.id
        ),
        '[]'::jsonb
    ) AS answers_json
FROM questions q
         JOIN sessions s ON s.current_question_id = q.id
WHERE s.id = $1
