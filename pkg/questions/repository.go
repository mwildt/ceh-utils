package questions

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/mwildt/ceh-utils/pkg/utils"
	"io"
	"math/rand"
	"os"
	"time"
)

type FileLogRepository struct {
	file   *os.File
	path   string
	values map[uuid.UUID]Question
	rand   *rand.Rand
	logger utils.Logger
}

func CreateRepo(path string) (repo *FileLogRepository, err error) {
	repo = &FileLogRepository{
		path:   path,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
		logger: utils.NewStdLogger("questions.repository"),
		values: make(map[uuid.UUID]Question),
	}
	if err := repo.init(); err != nil {
		return repo, err
	}
	if err := repo.load(); err != nil {
		return repo, err
	}
	if err := repo.open(); err != nil {
		return repo, err
	}
	return repo, err
}

func (repo *FileLogRepository) FindRandom(predicate utils.Predicate[Question]) (question Question, err error) {
	candidates := utils.FilterValues(repo.values, predicate)
	if len(candidates) == 0 {
		return question, err
	}
	randomIndex := repo.rand.Intn(len(candidates))
	return candidates[randomIndex], nil
}

func (repo *FileLogRepository) Save(question Question) (result Question, err error) {
	encoded, err := repo.encodeRecord(question)
	if err != nil {
		return question, err
	}

	buffLenBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(buffLenBytes, uint32(len(encoded)))
	if _, err := repo.file.Write(buffLenBytes); err != nil {
		return question, err
	}
	_, err = repo.file.Write(encoded)

	repo.values[question.Id] = question
	return question, err
}

func (repo *FileLogRepository) init() (err error) {
	if !fileExist(repo.filepath()) {
		_, err := os.Create(repo.filepath())
		return err
	}
	return nil
}

type QuestionPredicate func(q Question) bool

func IdNotIn(uuids []uuid.UUID) utils.Predicate[Question] {
	keyMap := make(map[uuid.UUID]uuid.UUID)
	for _, id := range uuids {
		keyMap[id] = id
	}
	return func(q Question) bool {
		_, exists := keyMap[q.Id]
		return !exists
	}
}

func IdEquals(value uuid.UUID) utils.Predicate[Question] {
	return func(q Question) bool {
		return value == q.Id
	}
}

func True() utils.Predicate[Question] {
	return func(q Question) bool {
		return true
	}
}

func (repo *FileLogRepository) FindAll(predicate utils.Predicate[Question]) (list []Question, err error) {
	for _, question := range repo.values {
		if predicate(question) {
			list = append(list, question)
		}
	}
	return list, err
}

func (repo *FileLogRepository) FindFirst(predicate utils.Predicate[Question]) (question Question, exists bool) {
	for _, question := range repo.values {
		if predicate(question) {
			return question, true
		}
	}
	return question, false
}

func (repo *FileLogRepository) Contains(predicate utils.Predicate[Question]) bool {
	_, exists := repo.FindFirst(predicate)
	return exists
}

func (repo *FileLogRepository) open() (err error) {
	if !fileExist(repo.filepath()) {
		return fmt.Errorf("could not open log segment. File %s not found", repo.filepath())
	}
	repo.file, err = os.OpenFile(repo.filepath(), os.O_APPEND|os.O_WRONLY, 0644)
	return err
}

func (repo *FileLogRepository) load() (err error) {
	repo.logger.Info("start load items from file-system (%s)", repo.filepath())
	file, err := os.OpenFile(repo.filepath(), os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for {
		lenBytes := make([]byte, 4)
		if _, err := io.ReadFull(file, lenBytes); err != nil {
			if err == io.EOF {
				repo.logger.Info("%d items loaded from file system", len(repo.values))
				return nil
			}
			return err
		}
		dataLen := binary.LittleEndian.Uint32(lenBytes)
		dataBuffer := make([]byte, int(dataLen))
		if _, err := io.ReadFull(file, dataBuffer); err != nil {
			return err
		}
		value, err := repo.decodeRecord(dataBuffer)
		if err != nil {
			return err
		}
		repo.values[value.Id] = value
	}

}

func (repo *FileLogRepository) filepath() string {
	return repo.path
}

func (repo *FileLogRepository) decodeRecord(record []byte) (question Question, err error) {
	encoding := base64.RawStdEncoding
	jsonValue := make([]byte, encoding.DecodedLen(len(record)))
	_, err = encoding.Decode(jsonValue, record)
	//fmt.Printf("JSON: \n|%s|\n-\n", jsonValue)
	if err != nil {

		return question, err
	}
	err = json.Unmarshal(jsonValue, &question)
	if err != nil {
		fmt.Printf("ERROR ON %w: \n|%s|\n-\n", err, jsonValue)
		return question, err
	}
	return question, err
}

func (repo *FileLogRepository) encodeRecord(value Question) (encoded []byte, err error) {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return encoded, err
	}
	encoder := base64.RawStdEncoding
	encoded = make([]byte, encoder.EncodedLen(len(jsonData)))
	encoder.Encode(encoded, jsonData)

	return encoded, err
}

func (repo *FileLogRepository) CountAll() int {
	return len(repo.values)
}

func fileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return true
	}
}
