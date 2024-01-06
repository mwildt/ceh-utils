package main

import (
	"fmt"
	"github.com/mwildt/ceh-utils/pkg/questions"
	"log"
)

func main() {

	repo, err := questions.CreateRepo("data/question.data")
	if err != nil {
		log.Fatal(err)
	}

	loader := Loader{BaseUrl: "https://cehtest.org/"}

	cntNew := 1
	cntOld := 0
	cntFailed := 0

	for cntNew > 0 {
		// load from cehtest.org
		cntNew, cntOld, cntFailed, err = loader.LoadAll(
			NewSessionRequestDTO{QuestionCount: 125, Versions: []int{12}},
			repo)
	}

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("new %d, old %d, failed: %d, total: %d", cntNew, cntOld, cntFailed, repo.CountAll())

}
