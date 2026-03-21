package dto

type SessionStateRecord struct {
	SessionRecord
	SettingsSurvey string `db:"settings"`
}
