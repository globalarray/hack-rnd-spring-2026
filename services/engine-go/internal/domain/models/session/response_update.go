package session

type ResponseUpdate struct {
	SessionID                 string
	ExpectedCurrentQuestionID string

	QuestionID string  // ID вопроса, на который ответили
	AnswerID   *string // UUID ответа (если выбор)
	RawText    *string // Текст (если свободный ввод)
	Weight     float64 // Вес этого конкретного ответа

	NextQuestionID *string // Куда переводим сессию
	IsFinished     bool    // Флаг завершения
}
