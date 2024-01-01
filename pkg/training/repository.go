package training

import (
	"context"
	"github.com/google/uuid"
	"github.com/mwildt/ceh-utils/pkg/events"
	"github.com/mwildt/ceh-utils/pkg/utils"
)

type Repository interface {
	Save(context.Context, *Training) (*Training, error)
	FindAllBy(ctx context.Context, predicate utils.Predicate[*Training]) ([]*Training, error)
	FindFirst(ctx context.Context, predicate utils.Predicate[*Training]) (*Training, bool)
}

func CreateRepository() (Repository, error) {
	return &repository{
		make(map[uuid.UUID]*Training),
	}, nil
}

type repository struct {
	values map[uuid.UUID]*Training
}

func IdEquals(value uuid.UUID) utils.Predicate[*Training] {
	return func(q *Training) bool {
		return value == q.Id
	}
}

func (repo repository) FindFirst(_ context.Context, predicate utils.Predicate[*Training]) (training *Training, exists bool) {
	for _, training := range repo.values {
		if predicate(training) {
			return training, true
		}
	}
	return training, false
}

func (repo repository) FindAllBy(_ context.Context, predicate utils.Predicate[*Training]) ([]*Training, error) {
	return utils.FilterValues(repo.values, predicate), nil
}

func (repo repository) Save(_ context.Context, training *Training) (*Training, error) {
	repo.values[training.Id] = training
	for _, event := range training.events {
		_ = events.Emit(event.Type, event.event)
	}
	training.events = training.events[:0]
	return training, nil
}
