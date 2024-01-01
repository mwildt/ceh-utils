package main

import (
	"github.com/mwildt/ceh-utils/pkg/questions"
	"github.com/mwildt/ceh-utils/pkg/utils"
	"github.com/mwildt/go-http/routing"
	"log"
	"net/http"
)

func main() {

	questionRepo, err := questions.CreateRepo("question.data")
	if err != nil {
		log.Fatal(err)
	}
	questionsController := questions.NewRestController(questionRepo)
	router := routing.NewRouter(questionsController.Routing)

	err = http.ListenAndServe(
		utils.GetEnvOrDefault("LISTEN_ADDRESS", ":8080"),
		router)

	if err != nil {
		log.Fatal(err)
	}
}
