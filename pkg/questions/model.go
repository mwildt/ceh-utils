package questions

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/mwildt/ceh-utils/pkg/events"
	"github.com/ohrenpiraten/go-collections/collections"
	"github.com/ohrenpiraten/go-collections/predicates"
)

type Option struct {
	Id     uuid.UUID
	Option string
}

type Question struct {
	Id        uuid.UUID
	Question  string
	Options   []Option
	AnswerIds []uuid.UUID
	Tags      []string
	Media     []string
	events    []event
}

func CreateQuestion(question string, options []Option, answerIds []uuid.UUID, media []string, tags []string) *Question {
	return (&Question{
		Id:        uuid.New(),
		Question:  question,
		Options:   options,
		AnswerIds: answerIds,
		Media:     media,
		Tags:      tags,
	}).init()
}

func (q *Question) init(events ...event) *Question {
	q.events = append([]event{}, events...)
	return q
}

func (q *Question) Update(text string, options []Option, answer []uuid.UUID) (updated *Question, err error) {
	if len(text) == 0 {
		return updated, fmt.Errorf("text must not be empty")
	}

	if len(options) < 2 {
		return updated, fmt.Errorf("options must be min 2")
	}

	if collections.AnyMatch(options, func(o Option) bool {
		return len(o.Option) == 0
	}) {
		return updated, fmt.Errorf("options must not be empty")
	}

	if len(answer) > 0 {
		optionAnswerIds := collections.Map(options, func(opt Option) uuid.UUID {
			return opt.Id
		})
		if !collections.ContainsAll(optionAnswerIds, answer) {
			return updated, fmt.Errorf("answers must all exist in options")
		}
		q.AnswerIds = answer
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

func ByQuestionText(text string) predicates.Predicate[*Question] {
	return func(comp *Question) bool {
		return comp.Question == text
	}
}

func IdEquals(value uuid.UUID) predicates.Predicate[*Question] {
	return func(q *Question) bool {
		return value == q.Id
	}
}
