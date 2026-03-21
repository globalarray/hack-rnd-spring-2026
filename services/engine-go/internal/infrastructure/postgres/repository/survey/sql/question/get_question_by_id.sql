SELECT
    q.id,
    q.order_num,
    q.type,
    q.text,
    q.logic_rules
FROM questions q
WHERE q.id = $1;
