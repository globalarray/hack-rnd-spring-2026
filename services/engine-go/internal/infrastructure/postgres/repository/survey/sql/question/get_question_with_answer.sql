SELECT q.id, q.order_num, q.logic_rules, a.weight
FROM questions q
JOIN answers a ON a.question_id = q.id
WHERE q.id = $1 AND a.id = $2