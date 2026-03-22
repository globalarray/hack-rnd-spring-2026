package session

import _ "embed"

//go:embed sql/insert.sql
var queryInsertSession string

//go:embed sql/get_current_question.sql
var queryGetCurrentQuestionBySessionID string

//go:embed sql/update.sql
var queryUpdateSession string

//go:embed sql/get_session_state.sql
var queryGetSessionState string

//go:embed sql/close_session.sql
var queryCloseSession string

//go:embed sql/has_active_session.sql
var queryHasActiveSession string

//go:embed sql/has_completed_share_link_session.sql
var queryHasCompletedShareLinkSession string

//go:embed sql/insert_response.sql
var queryInsertResponse string
