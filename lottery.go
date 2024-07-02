package abc

import (
	"math"
	"math/rand"
	"time"
)

// Example ↓↓↓
// temp 1: 0.29, 0.09, 0.02, 0.2, 0.4
// tamp 2: 0.008,0.006,0.005,0.005,0.005,0.001,0.631,0.339
// TODO: Lotteried(0.008,0.006,0.005,0.005,0.005,0.001,0.631,0.339)
func Lotteried[T NumdFloatd](probs ...T) int {
	length := len(probs)
	l := &Lottery{
		Probability: make([]cal32, 0),
		ProbRounds:  make([]cal32, length),
		length:      length,
	}
	for _, prob := range probs {
		l.Probability = append(l.Probability, cal32(prob))
	}
	return l.Rounds([]cal32{}).Lucky()
}

type Lottery struct {
	Probability []cal32
	ProbRounds  []cal32
	length      int
	count       int
}

type cal32 float32

func (c32 cal32) Float64() float64 {
	return float64(c32)
}

func (l *Lottery) Rounds(negetive []cal32) *Lottery {
	if l.length <= 1 {
		return l
	}
	if l.count == l.length {
		return l
	}
	if l.count == 0 {
		negetive = append(negetive, 1-l.Probability[l.count])
		l.ProbRounds[l.count] = cal32(math.Round((1-
			negetive[l.count]).Float64()*1000) / 1000)
		l.count++
		l.Rounds(negetive)
		return l
	}
	negetive = append(negetive, 1-
		l.Probability[l.count]/
			(l.Probability[l.count]*A(negetive))*
			l.Probability[l.count])
	l.ProbRounds[l.count] = cal32(math.Round((1-
		negetive[l.count]).Float64()*1000) / 1000)
	l.count++
	l.Rounds(negetive)
	return l
}

func (l *Lottery) Lucky() int {
	for k, round := range l.ProbRounds {
		rand.Seed(time.Now().UnixNano())
		h := int(round * 1000)
		if rand.Intn(1000+1) <= h {
			return k
		}
	}
	return -1
}

func A[T NumdFloatd](a []T) T {
	var mul T
	for i, v := range a {
		if i == 0 {
			mul = v
			continue
		}
		mul *= v
	}

	return mul
}
