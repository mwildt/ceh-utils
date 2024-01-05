package questions

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/mwildt/ceh-utils/pkg/utils"
)

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
	Media    []string
}

func (question Question) Update(text string, options []Option, answerId uuid.UUID) (updated Question, err error) {
	if len(text) == 0 {
		return updated, fmt.Errorf("text must not be empty")
	}

	if len(options) < 2 {
		return updated, fmt.Errorf("options must be min 2")
	}

	if utils.AnyMatch(options, func(o Option) bool {
		return len(o.Option) == 0
	}) {
		return updated, fmt.Errorf("options must not be empty")
	}

	var empty uuid.UUID
	if answerId != empty {
		if !utils.AnyMatch(options, func(o Option) bool {
			return o.Id == answerId
		}) {
			return updated, fmt.Errorf("answer must exist in options")
		}
		question.AnswerId = answerId
	}
	question.Options = options
	question.Question = text
	return question, err

}

func ByQuestionText(question Question) func(Question) bool {
	return func(comp Question) bool {
		return comp.Question == question.Question
	}
}
