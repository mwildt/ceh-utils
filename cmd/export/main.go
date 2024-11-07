package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/mwildt/ceh-utils/pkg/questions"
	"github.com/mwildt/ceh-utils/pkg/utils"
	"log"
	"os"
	"path"
)

type option struct {
	Id   uuid.UUID `json:"id"`
	Text string    `json:"text"`
}

type record struct {
	Id        uuid.UUID   `json:"id"`
	Text      string      `json:"text"`
	Options   []option    `json:"options"`
	AnswerIds []uuid.UUID `json:"answers"`
	Media     []string    `json:"media"`
}

func main() {

	repo, err := questions.CreateRepo(
		"data/question.data",
		path.Join("config/ceh-12-cehtest.org", "question.data"),
		path.Join("config/custom-json", "question.data"))

	if err != nil {
		log.Fatal(err)
	} else if all, err := repo.FindAll(utils.True[*questions.Question]()); err != nil {
		log.Fatal(err)
	} else {
		var records []record
		for _, q := range all {

			if rec, err := transformRecord(q); err != nil {
				log.Fatalf("fehler beim transformieren der frage: %s", err.Error())
			} else {
				records = append(records, rec)
			}
		}
		if err = writeJsonRecordsToFile(records, "questions.json"); err != nil {
			log.Fatalf("fehler beim schreiben der daten als JSON: %s", err.Error())
		}
	}

}

func transformRecord(q *questions.Question) (record, error) {

	media, err := utils.MapError(q.Media, func(m string) (string, error) {
		if imageData, err := os.ReadFile(path.Join("config/ceh-12-cehtest.org/media", m)); err != nil {
			return "", err
		} else {
			return base64.StdEncoding.EncodeToString(imageData), nil
		}
	})

	if err != nil {
		return record{}, err
	}

	return record{
		Id:   q.Id,
		Text: q.Question,
		Options: utils.Map(q.Options, func(opt questions.Option) option {
			return option{
				Id:   opt.Id,
				Text: opt.Option,
			}
		}),
		AnswerIds: q.AnswerIds,
		Media:     media,
	}, nil
}

func writeJsonRecordsToFile(records []record, path string) error {

	if jsonData, err := json.MarshalIndent(records, "", "    "); err != nil {
		return fmt.Errorf("err marshalling JSON %w", err)
	} else if file, err := os.Create(path); err != nil {
		return fmt.Errorf("error creating file: %w", err)
	} else {
		defer file.Close()
		_, err = file.Write(jsonData)
		return err
	}

}
