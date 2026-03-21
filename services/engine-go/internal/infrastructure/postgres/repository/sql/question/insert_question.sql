INSERT INTO questions (survey_id, order_num, type, text, logic_rules)
VALUES ($1, $2, $3, $4, $5)
    RETURNING id;