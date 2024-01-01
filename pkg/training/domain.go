package training

import (
	"github.com/google/uuid"
	"time"
)

type Challenge struct {
	Id     uuid.UUID
	Answer uuid.UUID
}

type event struct {
	Type  string
	event interface{}
}

type Training struct {
	Id        uuid.UUID
	Challenge Challenge
	Updated   time.Time
	Created   time.Time
	events    []event
}

type ChallengeProvider func() (Challenge, error)

func CreateTraining(nextChallenge ChallengeProvider) (training *Training, err error) {
	challenge, err := nextChallenge()
	if err != nil {
		return training, err
	}
	id := uuid.New()
	return &Training{
		Id:        id,
		Challenge: challenge,
		Updated:   time.Now(),
		Created:   time.Now(),
		events:    []event{{"training.created", CreatedEvent{id}}},
	}, nil
}

func (training *Training) Next(answerId uuid.UUID, nextChallenge ChallengeProvider) (success bool, err error) {
	success = training.Challenge.Answer == answerId

	training.events = append(training.events, event{"training.updated", UpdatedEvent{
		TrainingId:  training.Id,
		ChallengeId: training.Challenge.Id,
		AnswerId:    answerId,
		Passed:      success,
	}})

	if success {
		//session.stats.pass()
		challenge, err := nextChallenge()
		if err != nil {
			return success, err
		}
		training.Challenge = challenge
		training.Updated = time.Now()
		return success, err
	} else {
		//session.stats.fail()
		return success, nil
	}
}
