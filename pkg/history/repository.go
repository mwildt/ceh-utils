package history

import (
	"context"
	"github.com/google/uuid"
	"github.com/mwildt/ceh-utils/pkg/utils"
)

type Repository struct {
	values map[uuid.UUID]History
}

func CreateRepo() (repo Repository, err error) {
	return Repository{make(map[uuid.UUID]History)}, nil
}

func (repo *Repository) Save(_ context.Context, hist History) (History, error) {
	repo.values[hist.Id] = hist
	return hist, nil
}

func IdEquals(value uuid.UUID) utils.Predicate[History] {
	return func(q History) bool {
		return value == q.Id
	}
}

func (repo *Repository) FindFirst(_ context.Context, predicate utils.Predicate[History]) (history History, exists bool) {
	for _, history := range repo.values {
		if predicate(history) {
			return history, true
		}
	}
	return history, false
}
