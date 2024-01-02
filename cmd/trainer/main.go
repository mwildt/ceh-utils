package main

import (
	"github.com/mwildt/ceh-utils/pkg/history"
	"github.com/mwildt/ceh-utils/pkg/questions"
	"github.com/mwildt/ceh-utils/pkg/training"
	"github.com/mwildt/ceh-utils/pkg/utils"
	"github.com/mwildt/go-http/routing"
	"log"
	"net/http"
)

func main() {

	questionRepo, err := questions.CreateRepo(utils.GetEnvOrDefault("QUESTION_STORAGE_DIR", "./question.data"))
	if err != nil {
		log.Fatal(err)
	}
	questionsController := questions.NewRestController(questionRepo)

	trainingRepo, err := training.CreateRepository()
	if err != nil {
		log.Fatal(err)
	}

	historyRepo, err := history.CreateRepo()
	if err != nil {
		log.Fatal(err)
	}

	if err = history.Subscribe(historyRepo); err != nil {
		log.Fatal(err)
	}
	trainingController := training.NewRestController(trainingRepo, func() (training.Challenge, error) {
		q, err := questionRepo.FindRandom(utils.True[questions.Question]())
		return training.Challenge{
			Id:     q.Id,
			Answer: q.AnswerId,
		}, err
	})

	router := routing.NewRouter(
		questionsController.Routing,
		trainingController.Routing,
		history.NewRestController(historyRepo).Routing,
		func(router routing.Routing) {
			router.HandleFunc(routing.Path("/"), utils.NotFound)
		},
	)

	err = http.ListenAndServe(
		utils.GetEnvOrDefault("LISTEN_ADDRESS", ":8080"),
		router)

	if err != nil {
		log.Fatal(err)
	}
}
