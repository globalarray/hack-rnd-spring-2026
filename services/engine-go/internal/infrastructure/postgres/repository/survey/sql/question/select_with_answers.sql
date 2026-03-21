SELECT q.id, q.text, q.logic_rules,
       COLEASCE((SELECT jsonb_agg(a) FROM answers a WHERE a.question_id = q.id), [])
questions q WHERE q.order_num = $1 AND q.survey_id = $2;