package questions

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/mwildt/ceh-utils/pkg/events"
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
	events   []event
}

func CreateQuestion(question string, options []Option, answerId uuid.UUID, media []string, tags []string) *Question {
	return (&Question{
		Id:       uuid.New(),
		Question: question,
		Options:  options,
		AnswerId: answerId,
		Media:    media,
		Tags:     tags,
	}).init()
}

func (q *Question) init(events ...event) *Question {
	q.events = append([]event{}, events...)
	return q
}

func (q *Question) Update(text string, options []Option, answerId uuid.UUID) (updated *Question, err error) {
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
		q.AnswerId = answerId
		q.events = append(q.events, updatedEvent(q))
	}
	q.Options = options
	q.Question = text
	return q, err
}

func (q *Question) emitEvents() {
	for _, event := range q.events {
		_ = events.Emit(event.Type, event.event)
	}
	q.events = q.events[:0]
}

func ByQuestionText(text string) utils.Predicate[*Question] {
	return func(comp *Question) bool {
		return comp.Question == text
	}
}

func IdEquals(value uuid.UUID) utils.Predicate[*Question] {
	return func(q *Question) bool {
		return value == q.Id
	}
}

func True() utils.Predicate[*Question] {
	return func(q *Question) bool {
		return true
	}
}
