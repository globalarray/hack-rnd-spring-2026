CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS surveys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    psychologist_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    settings JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
    );

CREATE INDEX idx_surveys_psychologist_id ON surveys(psychologist_id);

CREATE TYPE session_status as ENUM (
    'CREATED',
    'IN_PROGRESS',
    'REVOKED',
    'COMPLETED'
);

-- example: "rules": {"uuid-ans-1": {"action": "JUMP", "next": "uuid-ans-3}}

CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_id UUID NOT NULL REFERENCES surveys(id) ON DELETE CASCADE,
    order_num INT NOT NULL,
    type int NOT NULL,
    text TEXT NOT NULL,
    logic_rules JSONB DEFAULT '{"rules": [], "default_next": "linear"}'::jsonb
    );

CREATE INDEX idx_questions_survey_id ON questions(survey_id);
CREATE INDEX idx_questions_order ON questions(survey_id, order_num);

CREATE TABLE IF NOT EXISTS answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    weight NUMERIC DEFAULT 0,
    category_tag VARCHAR(100)
);

CREATE INDEX idx_answers_question_id ON answers(question_id);

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_id UUID NOT NULL REFERENCES surveys(id) ON DELETE CASCADE,
    client_metadata JSONB DEFAULT '{}'::jsonb,
    status session_status NOT NULL DEFAULT 'CREATED',
    started_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ
    );

CREATE INDEX idx_sessions_survey_id ON sessions(survey_id);

CREATE TABLE IF NOT EXISTS responses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    answer_id UUID REFERENCES answers(id) ON DELETE SET NULL,
    raw_text TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
    );

CREATE INDEX idx_responses_session_id ON responses(session_id);