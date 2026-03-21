SELECT
    s.id AS survey_id,
    s.title,
    COUNT(sess.id) FILTER (WHERE sess.status = 'COMPLETED')::int AS completions_count
FROM surveys s
LEFT JOIN sessions sess ON sess.survey_id = s.id
WHERE s.psychologist_id = $1
GROUP BY s.id, s.title
ORDER BY s.title ASC;
