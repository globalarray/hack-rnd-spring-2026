package session

type SessionStatus string

const (
	CreatedStatus    SessionStatus = "CREATED"
	InProgressStatus SessionStatus = "IN_PROGRESS"
	CompletedStatus  SessionStatus = "COMPLETED"
	RevokedStatus    SessionStatus = "REVOKED"
)
