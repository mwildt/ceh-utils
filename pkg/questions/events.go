package questions

import "github.com/google/uuid"

type UpdatedEvent struct {
	QuestionId uuid.UUID `json:"questionId"`
	AnswerId   uuid.UUID `json:"answerId"`
}

type event struct {
	Type  string
	event interface{}
}

func updatedEvent(question *Question) event {
	return event{"question.updated", UpdatedEvent{question.Id, question.AnswerId}}
}
