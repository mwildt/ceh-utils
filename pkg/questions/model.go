package questions

import "github.com/google/uuid"

type Option struct {
	Id     uuid.UUID
	Option string
}

type Question struct {
	Id       uuid.UUID
	Question string
	Options  []Option
	AnswerId uuid.UUID
	Tags     []string
}

func ByQuestionText(question Question) func(Question) bool {
	return func(comp Question) bool {
		return comp.Question == question.Question
	}
}
