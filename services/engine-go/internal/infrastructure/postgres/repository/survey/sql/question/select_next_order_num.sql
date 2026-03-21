SELECT id FROM questions
WHERE survey_id = $1 AND order_num > $2
ORDER BY order_num ASC LIMIT 1;