package training

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/mwildt/ceh-utils/pkg/events"
	"github.com/mwildt/ceh-utils/pkg/utils"
	"os"
)

type Repository interface {
	Save(context.Context, *Training) (*Training, error)
	FindAllBy(ctx context.Context, predicate utils.Predicate[*Training]) ([]*Training, error)
	FindFirst(ctx context.Context, predicate utils.Predicate[*Training]) (*Training, bool)
}

func IdEquals(value uuid.UUID) utils.Predicate[*Training] {
	return func(q *Training) bool {
		return value == q.Id
	}
}

type fileRepository struct {
	values  map[uuid.UUID]*Training
	path    string
	logger  utils.Logger
	file    *os.File
	decoder utils.Decoder[Training]
	encoder utils.Encoder[Training]
}

func CreateFileRepository(path string) (Repository, error) {
	repo := &fileRepository{
		values:  make(map[uuid.UUID]*Training),
		path:    path,
		logger:  utils.NewStdLogger("trainings.repository"),
		encoder: utils.B64JsonEncoder[Training],
		decoder: utils.B64JsonDecoder[Training],
	}

	if err := utils.CreateFileIfNotExists(repo.filepath()); err != nil {
		return repo, err
	}
	if err := repo.load(); err != nil {
		return repo, err
	}
	if err := repo.open(); err != nil {
		return repo, err
	}
	return repo, nil
}

func (repo *fileRepository) filepath() string {
	return repo.path
}

func (repo *fileRepository) open() (err error) {
	if !utils.FileExist(repo.filepath()) {
		return fmt.Errorf("could not open log segment. File %s not found", repo.filepath())
	}
	repo.file, err = os.OpenFile(repo.filepath(), os.O_APPEND|os.O_WRONLY, 0644)
	return err
}

func (repo *fileRepository) load() (err error) {
	repo.logger.Info("start load items from file-system (%s)", repo.filepath())

	count, err := utils.LoadFromFile(repo.filepath(), func(buffer []byte) error {
		value, err := repo.decoder(buffer)
		if err != nil {
			return err
		}
		value.init()
		repo.values[value.Id] = &value
		return nil
	})
	if err == nil {
		repo.logger.Info("%d items loaded from file system, %d in store", count, len(repo.values))
	}
	return err
}

func (repo *fileRepository) Save(ctx context.Context, training *Training) (_ *Training, err error) {
	err = utils.Append(repo.file, *training, repo.encoder)
	if err != nil {
		return training, err
	}
	repo.values[training.Id] = training
	return training, err
}

func (repo *fileRepository) FindAllBy(ctx context.Context, predicate utils.Predicate[*Training]) (list []*Training, err error) {
	for _, training := range repo.values {
		if predicate(training) {
			list = append(list, training)
		}
	}
	return list, err
}

func (repo *fileRepository) FindFirst(ctx context.Context, predicate utils.Predicate[*Training]) (*Training, bool) {
	for _, training := range repo.values {
		if predicate(training) {
			return training, true
		}
	}
	return nil, false
}

// IN MEMORY
// IN MEMORY
// IN MEMORY
// IN MEMORY
// IN MEMORY
// IN MEMORY
// IN MEMORY

func CreateInMemoryRepository() (Repository, error) {
	return &inMemoryRepository{
		make(map[uuid.UUID]*Training),
	}, nil
}

type inMemoryRepository struct {
	values map[uuid.UUID]*Training
}

func (repo inMemoryRepository) FindFirst(_ context.Context, predicate utils.Predicate[*Training]) (training *Training, exists bool) {
	for _, training := range repo.values {
		if predicate(training) {
			return training, true
		}
	}
	return training, false
}

func (repo inMemoryRepository) FindAllBy(_ context.Context, predicate utils.Predicate[*Training]) ([]*Training, error) {
	return utils.FilterValues(repo.values, predicate), nil
}

func (repo inMemoryRepository) Save(_ context.Context, training *Training) (*Training, error) {
	repo.values[training.Id] = training
	for _, event := range training.events {
		_ = events.Emit(event.Type, event.event)
	}
	training.events = training.events[:0]
	return training, nil
}
