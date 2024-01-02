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

type Stats struct {
	totalChallenges          int
	passedChallenges         int
	failedChallenges         int
	currentChallengeAttempts int
}

func (stats *Stats) pass() {
	stats.totalChallenges++
	if stats.currentChallengeAttempts == 0 {
		stats.passedChallenges++
	}
	stats.currentChallengeAttempts = 0
}

func (stats *Stats) fail() {
	if stats.currentChallengeAttempts == 0 {
		stats.failedChallenges++
	}
	stats.currentChallengeAttempts++
}

type Training struct {
	Id        uuid.UUID
	Challenge Challenge
	Updated   time.Time
	Created   time.Time
	events    []event
	stats     *Stats
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
		stats: &Stats{
			totalChallenges:          1,
			passedChallenges:         0,
			failedChallenges:         0,
			currentChallengeAttempts: 0,
		},
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
		training.stats.pass()
		challenge, err := nextChallenge()
		if err != nil {
			return success, err
		}
		training.Challenge = challenge
		training.Updated = time.Now()
		return success, err
	} else {
		training.stats.fail()
		return success, nil
	}
}
