package training

import (
	"github.com/google/uuid"
	"github.com/mwildt/ceh-utils/pkg/utils"
	"testing"
	"time"
)

func TestDomainFindCandidateSingle(t *testing.T) {
	candidates := []*TrainingChallenge{
		{
			Id:        uuid.New(),
			Answer:    uuid.New(),
			Level:     1,
			Timestamp: time.Now(),
			Done:      false,
		},
	}
	candidate, found := filterCandidates(candidates)
	utils.Assert(t, found, "not found")
	utils.Assert(t, candidate == candidates[1], "wrong found")
}

func TestDomainFindCandidateCompareLevel(t *testing.T) {
	candidates := []*TrainingChallenge{
		{
			Id:        uuid.New(),
			Answer:    uuid.New(),
			Level:     1,
			Timestamp: time.Now(),
			Done:      false,
		},
		{
			Id:        uuid.New(),
			Answer:    uuid.New(),
			Level:     0,
			Timestamp: time.Now(),
			Done:      false,
		},
	}
	candidate, found := filterCandidates(candidates)
	utils.Assert(t, found, "not found")
	utils.Assert(t, candidate == candidates[1], "wrong found")
}

func TestDomainFindCandidateCutoffNoneFound(t *testing.T) {
	candidates := []*TrainingChallenge{
		{
			Id:        uuid.New(),
			Answer:    uuid.New(),
			Level:     1,
			Timestamp: time.Now().Add(time.Minute * 100),
			Done:      false,
		},
		{
			Id:        uuid.New(),
			Answer:    uuid.New(),
			Level:     0,
			Timestamp: time.Now().Add(time.Minute),
			Done:      false,
		},
	}
	_, found := filterCandidates(candidates)
	utils.Assert(t, !found, "found")
}

func TestDomainFindCandidateDoneffNoneFound(t *testing.T) {
	candidates := []*TrainingChallenge{
		{
			Id:        uuid.New(),
			Answer:    uuid.New(),
			Level:     4,
			Timestamp: time.Now(),
			Done:      true,
		},
		{
			Id:        uuid.New(),
			Answer:    uuid.New(),
			Level:     4,
			Timestamp: time.Now(),
			Done:      true,
		},
	}
	_, found := filterCandidates(candidates)
	utils.Assert(t, !found, "found")
}

func TestDomainFindCandidateByTime(t *testing.T) {
	candidates := []*TrainingChallenge{
		{
			Id:        uuid.New(),
			Answer:    uuid.New(),
			Level:     1,
			Timestamp: time.Now().Add(-1 * time.Minute),
			Done:      false,
		},
		{
			Id:        uuid.New(),
			Answer:    uuid.New(),
			Level:     1,
			Timestamp: time.Now().Add(-2 * time.Minute),
			Done:      false,
		},
		{
			Id:        uuid.New(),
			Answer:    uuid.New(),
			Level:     1,
			Timestamp: time.Now().Add(-3 * time.Minute),
			Done:      false,
		},
		{
			Id:        uuid.New(),
			Answer:    uuid.New(),
			Level:     1,
			Timestamp: time.Now().Add(-1 * time.Minute),
			Done:      false,
		},
	}
	candidate, found := filterCandidates(candidates)
	utils.Assert(t, found, "found")
	utils.Assert(t, candidate == candidates[2], "wrong found")
}
