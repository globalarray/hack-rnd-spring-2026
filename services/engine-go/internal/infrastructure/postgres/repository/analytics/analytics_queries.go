package analytics

import _ "embed"

//go:embed sql/get_session_data.sql
var queryGetSessionData string

//go:embed sql/get_session_responses.sql
var queryGetSessionResponses string

//go:embed sql/list_survey_sessions.sql
var queryListSurveySessions string
