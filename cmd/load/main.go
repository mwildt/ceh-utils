package main

import (
	"fmt"
	"github.com/mwildt/ceh-utils/pkg/questions"
	"log"
)

func main() {

	repo, err := questions.CreateRepo("question.data")
	if err != nil {
		log.Fatal(err)
	}

	loader := Loader{BaseUrl: "https://cehtest.org/"}

	// load from cehtest.org
	cntNew, cntOld, cntFailed, err := loader.LoadAll(
		NewSessionRequestDTO{QuestionCount: 125, Versions: []int{12}},
		repo)

	//cntNew, cntOld, cntFailed, err := loader.LoadFile(repo, "custom.json")

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("new %d, old %d, failed: %d, total: %d", cntNew, cntOld, cntFailed, repo.CountAll())

}
