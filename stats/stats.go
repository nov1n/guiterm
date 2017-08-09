package stats

import (
	"math"
)

const (
	Half          = 50
	Full          = 100
	multiplierInc = 10
	maxStreak     = 9
)

type Stats struct {
	TotalNotes    int
	CorrectNotes  int
	MistakenNotes int
	Score         int
	LastNote      int
	Streak        int
}

func New() *Stats {
	return &Stats{
		TotalNotes:    0,
		CorrectNotes:  0,
		Streak:        0,
		MistakenNotes: 0,
		Score:         0,
		LastNote:      0,
	}
}

func (s *Stats) Add(n int) {
	score := n * s.Multiplier()
	s.LastNote = score
	s.Score += score

	if s.Score < 0 {
		s.Score = 0
	}
}

func (s *Stats) Multiplier() int {
	return int(math.Min(float64(1+(s.Streak/multiplierInc)), float64(maxStreak)))
}

func (s *Stats) Accuracy() int {
	if s.TotalNotes == 0 {
		return 0
	}
	return int(math.Max(float64(s.CorrectNotes)/(float64(s.TotalNotes)+float64(s.MistakenNotes))*100, 0))
}

func (s *Stats) TotalNotesAdd(n int) {
	s.TotalNotes += n
}

func (s *Stats) Correct() {
	s.CorrectNotes += 1
	s.TotalNotes += 1
	s.Streak += 1
}

func (s *Stats) Incorrect() {
	s.Streak = 0
	s.MistakenNotes += 1
	s.Add(-Half)
}
