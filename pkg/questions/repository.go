package questions

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/mwildt/ceh-utils/pkg/utils"
	"github.com/ohrenpiraten/go-collections/dictionaray"
	"github.com/ohrenpiraten/go-collections/predicates"
	"math/rand"
	"os"
	"sync"
	"time"
)

type FileLogRepository struct {
	file   *os.File
	path   string
	values map[uuid.UUID]*Question
	rand   *rand.Rand
	logger utils.Logger
	mutex  *sync.Mutex
}

func CreateRepo(path string, preloadFiles ...string) (repo *FileLogRepository, err error) {
	repo = &FileLogRepository{
		path:   path,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
		logger: utils.NewStdLogger("questions.repository"),
		values: make(map[uuid.UUID]*Question),
		mutex:  &sync.Mutex{},
	}
	if err := utils.CreateFileIfNotExists(repo.filepath()); err != nil {
		return repo, err
	}

	for _, preloadFile := range preloadFiles {
		if err = repo.loadFile(preloadFile); err != nil {
			return nil, err
		}
	}

	if err := repo.load(); err != nil {
		return repo, err
	}
	if err := repo.open(); err != nil {
		return repo, err
	}
	return repo, err
}

func (repo *FileLogRepository) FindRandom(predicate predicates.Predicate[*Question]) (question *Question, err error) {
	candidates := dictionaray.FilterValues(repo.values, predicate)
	if len(candidates) == 0 {
		return question, err
	}
	randomIndex := repo.rand.Intn(len(candidates))
	return candidates[randomIndex], nil
}

func (repo *FileLogRepository) Save(question *Question) (_ *Question, err error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	err = utils.Append(repo.file, question, repo.encodeRecord)
	if err != nil {
		return question, err
	}
	repo.values[question.Id] = question
	question.emitEvents()
	return question, err
}

type QuestionPredicate func(q Question) bool

func IdNotIn(uuids []uuid.UUID) predicates.Predicate[*Question] {
	keyMap := make(map[uuid.UUID]uuid.UUID)
	for _, id := range uuids {
		keyMap[id] = id
	}
	return func(q *Question) bool {
		_, exists := keyMap[q.Id]
		return !exists
	}
}

func (repo *FileLogRepository) FindAll(predicate predicates.Predicate[*Question]) (list []*Question, err error) {
	for _, question := range repo.values {
		if predicate(question) {
			list = append(list, question)
		}
	}
	return list, err
}

func (repo *FileLogRepository) FindFirst(predicate predicates.Predicate[*Question]) (question *Question, exists bool) {
	for _, question := range repo.values {
		if predicate(question) {
			return question, true
		}
	}
	return question, false
}

func (repo *FileLogRepository) Contains(predicate predicates.Predicate[*Question]) bool {
	_, exists := repo.FindFirst(predicate)
	return exists
}

func (repo *FileLogRepository) open() (err error) {
	if !utils.FileExist(repo.filepath()) {
		return fmt.Errorf("could not open log segment. File %s not found", repo.filepath())
	}
	repo.file, err = os.OpenFile(repo.filepath(), os.O_APPEND|os.O_WRONLY, 0644)
	return err
}

func (repo *FileLogRepository) loadFile(path string) (err error) {
	repo.logger.Info("start load items from file-system (%s)", path)
	count, err := utils.LoadFromFile(path, func(buffer []byte) error {
		value, err := repo.decodeRecord(buffer)
		if err != nil {
			return err
		}
		repo.values[value.Id] = value
		return nil
	})
	if err == nil {
		repo.logger.Info("%d items loaded from %s, %d in store", count, path, len(repo.values))
	}
	return err
}

func (repo *FileLogRepository) load() (err error) {
	return repo.loadFile(repo.filepath())
}

func (repo *FileLogRepository) filepath() string {
	return repo.path
}

func (repo *FileLogRepository) decodeRecord(record []byte) (question *Question, err error) {
	return utils.B64JsonDecoder[*Question](record)
}

func (repo *FileLogRepository) encodeRecord(value *Question) (encoded []byte, err error) {
	return utils.B64JsonEncoder(value)
}

func (repo *FileLogRepository) CountAll() int {
	return len(repo.values)
}
