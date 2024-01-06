package training

import "github.com/google/uuid"

type CreatedEvent struct {
	TrainingId uuid.UUID `json:"trainingId"`
}

type UpdatedEvent struct {
	TrainingId  uuid.UUID   `json:"trainingId"`
	ChallengeId uuid.UUID   `json:"challengeId"`
	AnswerIds   []uuid.UUID `json:"answerId"`
	Passed      bool        `json:"passed"`
}
