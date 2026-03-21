DROP TABLE IF EXISTS surveys;
DROP TABLE IF EXISTS questions;
DROP TABLE IF EXISTS answers;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS responses;

DROP INDEX IF EXISTS idx_surveys_psychologist_id;
DROP INDEX IF EXISTS idx_questions_survey_id;
DROP INDEX IF EXISTS idx_questions_order;
DROP INDEX IF EXISTS idx_answers_question_id;
DROP INDEX IF EXISTS idx_sessions_survey_id;
DROP INDEX IF EXISTS idx_responses_session_id;

DROP TYPE IF EXISTS session_status;
DROP TYPE IF EXISTS question_type;