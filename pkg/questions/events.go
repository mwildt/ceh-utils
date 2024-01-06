package questions

import "github.com/google/uuid"

type UpdatedEvent struct {
	QuestionId uuid.UUID   `json:"questionId"`
	AnswerIds  []uuid.UUID `json:"answerIds"`
}

type event struct {
	Type  string
	event interface{}
}

func updatedEvent(question *Question) event {
	return event{"question.updated", UpdatedEvent{question.Id, question.AnswerIds}}
}
