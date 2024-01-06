package history

import (
	"github.com/google/uuid"
)

type History struct {
	Id             uuid.UUID
	currentAnswers []uuid.UUID
	history        []Item
}

type Item struct {
	ChallengeId   uuid.UUID
	GivenAnswers  []uuid.UUID
	SolvingAnswer []uuid.UUID
}

func CreateHistory(id uuid.UUID) History {
	return History{
		Id:             id,
		currentAnswers: make([]uuid.UUID, 0),
		history:        make([]Item, 0),
	}
}

func (hist *History) HistoryItemAt(index int) (exists bool, item Item) {
	if len(hist.history) > index {
		idx := len(hist.history) - index
		return true, hist.history[idx-1]
	} else {
		return false, item
	}
}

func (hist *History) AddAnswer(answerIds []uuid.UUID) {
	hist.currentAnswers = append(hist.currentAnswers, answerIds...)
}

func (hist *History) Size() int {
	return len(hist.history)
}

func (hist *History) Finalize(challengeId uuid.UUID, solvingAnswerId []uuid.UUID) {
	hist.history = append(hist.history, Item{
		ChallengeId:   challengeId,
		GivenAnswers:  hist.currentAnswers,
		SolvingAnswer: solvingAnswerId,
	})
	hist.currentAnswers = make([]uuid.UUID, 0)
}
