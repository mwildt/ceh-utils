package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/mwildt/ceh-utils/pkg/questions"
	"github.com/mwildt/ceh-utils/pkg/utils"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"strings"
)

type NextQuestionResponseDTO struct {
	Question Question `json:"question"`
}

type optionValue string

func (f *optionValue) UnmarshalJSON(data []byte) error {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	switch v := raw.(type) {
	case float64:
		*f = optionValue(strconv.FormatFloat(v, 'f', -1, 64))
	case string:
		*f = optionValue(v)
	default:
		return fmt.Errorf("unexpected type %T for CustomIntOrString", v)
	}
	return nil
}

type NewSessionRequestDTO struct {
	QuestionCount int   `json:"question_count"`
	Versions      []int `json:"versions"`
}

type Question struct {
	Question    string      `json:"question"`
	Media       string      `json:"media"`
	A           optionValue `json:"A"`
	B           optionValue `json:"B"`
	C           optionValue `json:"C"`
	D           optionValue `json:"D"`
	E           optionValue `json:"E"`
	F           optionValue `json:"F"`
	G           optionValue `json:"G"`
	Answer      string      `json:"answer"`
	Version     string      `json:"version"`
	Explanation string      `json:"explanation"`
}

func (dto NewSessionRequestDTO) MustJson() []byte {
	data, err := json.Marshal(&dto)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

type Loader struct {
	BaseUrl string
}

func (loader *Loader) LoadAll(dto NewSessionRequestDTO, repo *questions.FileLogRepository, tags ...string) (cntNew int, cntOld int, cntFailed int, err error) {
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: cookieJar}

	if err = loader.create(client, dto); err != nil {
		return cntNew, cntOld, cntFailed, err
	}

	for i := 0; i < dto.QuestionCount; i++ {
		if apiQuestion, err := loader.nextQuestion(client); err != nil {
			log.Printf("Fehler: %s", err)
			cntFailed = cntFailed + 1
		} else {

			fmt.Print("*")

			//fmt.Println(apiQuestion.Question)

			question := mapToModel(apiQuestion, tags...)
			if !repo.Contains(questions.ByQuestionText(question.Question)) {

				if len(question.Media) > 0 {
					for _, media := range question.Media {
						if err = Download(loader.BaseUrl+"/media/"+media, "ceh-12-cehtest.org/media/"+media); err != nil {
							log.Fatal(err.Error())
						}
					}
				}

				_, err = repo.Save(question)
				cntNew = cntNew + 1
				if err != nil {
					log.Printf("\nFehler: %s\n", err)
					cntFailed = cntFailed + 1
				}
			} else {
				cntOld = cntOld + 1
			}
		}
	}
	return cntNew, cntOld, cntFailed, err
}

func (loader *Loader) create(client *http.Client, dto NewSessionRequestDTO) (err error) {
	req, _ := http.NewRequest("POST", loader.BaseUrl+"start_test", bytes.NewReader(dto.MustJson()))

	req.Header.Set("Accept", "*")
	req.Header.Set("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	return err
}

func (loader *Loader) nextQuestion(client *http.Client) (question Question, err error) {
	req, err := http.NewRequest("GET", "https://cehtest.org/next_question", nil)
	if err != nil {
		return question, err
	}
	req.Header.Set("Accept", "*")
	resp, err := client.Do(req)
	if err != nil {
		return question, err
	}
	//fmt.Println("Status Code:", resp.Status)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return question, err
	}

	var apiResponse NextQuestionResponseDTO
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		fmt.Println(string(body))
		return question, err
	}
	return apiResponse.Question, err
}

func Download(url, outputFilePath string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Check if the request was successful (status code 200)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file, status code: %d", response.StatusCode)
	}

	// Create the output file
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Copy the response body to the output file
	_, err = io.Copy(outputFile, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func (loader *Loader) LoadFile(repo *questions.FileLogRepository, filePath string) (cntNew int, cntOld int, cntFailed int, err error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return cntNew, cntOld, cntFailed, err
	}

	var apiQuestion NextQuestionResponseDTO
	err = json.Unmarshal(fileContent, &apiQuestion)
	if err != nil {
		return cntNew, cntOld, cntFailed, err
	}

	question := mapToModel(apiQuestion.Question)
	if !repo.Contains(questions.ByQuestionText(question.Question)) {
		_, err = repo.Save(question)
		if err != nil {
			return cntNew, cntOld, 1, err
		} else {
			return 1, cntOld, cntFailed, err
		}
	}
	return cntNew, 1, cntFailed, err
}

func mapToModel(question Question, tags ...string) *questions.Question {
	var answerIds []uuid.UUID
	options := make([]questions.Option, 0)

	answers := strings.Split(question.Answer, " ")

	if question.A != "" {
		id := uuid.New()
		options = append(options, questions.Option{Id: id, Option: string(question.A)})
		if utils.Contains(answers, "A") {
			answerIds = append(answerIds, id)
		}
	}

	if question.B != "" {
		id := uuid.New()
		options = append(options, questions.Option{Id: id, Option: string(question.B)})
		if utils.Contains(answers, "B") {
			answerIds = append(answerIds, id)
		}
	}

	if question.C != "" {
		id := uuid.New()
		options = append(options, questions.Option{Id: id, Option: string(question.C)})
		if utils.Contains(answers, "C") {
			answerIds = append(answerIds, id)
		}
	}

	if question.D != "" {
		id := uuid.New()
		options = append(options, questions.Option{Id: id, Option: string(question.D)})
		if utils.Contains(answers, "D") {
			answerIds = append(answerIds, id)
		}
	}

	if question.E != "" {
		id := uuid.New()
		options = append(options, questions.Option{Id: id, Option: string(question.E)})
		if utils.Contains(answers, "E") {
			answerIds = append(answerIds, id)
		}
	}

	if question.F != "" {
		id := uuid.New()
		options = append(options, questions.Option{Id: id, Option: string(question.F)})
		if utils.Contains(answers, "F") {
			answerIds = append(answerIds, id)
		}
	}

	if question.G != "" {
		id := uuid.New()
		options = append(options, questions.Option{Id: id, Option: string(question.G)})
		if utils.Contains(answers, "G") {
			answerIds = append(answerIds, id)
		}
	}

	media := make([]string, 0)

	if len(question.Media) > 0 {
		media = strings.Split(question.Media, ",")
	}

	return questions.CreateQuestion(
		question.Question,
		options,
		answerIds,
		media,
		tags)

}
