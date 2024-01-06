package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/mwildt/ceh-utils/pkg/questions"
	"log"
	"os"
)

type JsonOption struct {
	Text   string `json:"text"`
	Answer bool   `json:"answer"`
}

type JsonQuestion struct {
	Question    string       `json:"question"`
	Media       []string     `json:"media"`
	Tags        []string     `json:"tags"`
	Options     []JsonOption `json:"options"`
	Answer      string       `json:"answer"`
	Explanation string       `json:"explanation"`
}

func main() {

	var jsonQuestions []JsonQuestion

	if data, err := os.ReadFile("json-data/set-1.json"); err != nil {
		log.Fatal(err)
	} else if err = json.Unmarshal(data, &jsonQuestions); err != nil {
		log.Fatal(err)
	} else if repo, err := questions.CreateRepo("config/custom-json/question.data"); err != nil {
		log.Fatal(err)
	} else {

		cntNew := 0
		cntOld := 0
		cntFailed := 0

		for _, jsonJquestion := range jsonQuestions {
			var options []questions.Option
			var answers []uuid.UUID
			for _, jsonOption := range jsonJquestion.Options {
				id := uuid.New()
				options = append(options, questions.Option{
					Id:     id,
					Option: jsonOption.Text,
				})
				if jsonOption.Answer {
					answers = append(answers, id)
				}
			}

			question := questions.CreateQuestion(
				jsonJquestion.Question,
				options,
				answers,
				jsonJquestion.Media,
				jsonJquestion.Tags)

			if !repo.Contains(questions.ByQuestionText(question.Question)) {
				cntNew = cntNew + 1
				_, err = repo.Save(question)
				if err != nil {
					log.Printf("\nFehler: %s\n", err)
					cntFailed = cntFailed + 1
				}
			} else {
				cntOld = cntOld + 1
			}
		}
		fmt.Printf("new %d, old %d, failed: %d, total: %d", cntNew, cntOld, cntFailed, repo.CountAll())
	}

}
