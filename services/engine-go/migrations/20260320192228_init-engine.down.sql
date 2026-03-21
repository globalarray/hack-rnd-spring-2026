DROP TABLE IF EXISTS surveys CASCADE;
DROP TABLE IF EXISTS questions CASCADE;
DROP TABLE IF EXISTS answers CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS responses CASCADE;

DROP INDEX IF EXISTS idx_surveys_psychologist_id;
DROP INDEX IF EXISTS idx_questions_survey_id;
DROP INDEX IF EXISTS idx_questions_order;
DROP INDEX IF EXISTS idx_answers_question_id;
DROP INDEX IF EXISTS idx_sessions_survey_id;
DROP INDEX IF EXISTS idx_responses_session_id;

DROP TYPE IF EXISTS session_status;