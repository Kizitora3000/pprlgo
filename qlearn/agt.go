package qlearn

import (
	"math/rand"
	"strconv"
	"strings"
)

type Agent struct {
	Nact     int
	InitValQ float64
	Epsilon  float64
	Alpha    float64
	Gamma    float64
	Q        map[string][]float64
	LenQ     int
}

func NewAgent() *Agent {
	return &Agent{
		Nact:     2,
		InitValQ: 0,
		Epsilon:  0.1,
		Alpha:    0.1,
		Gamma:    0.9,
		Q:        map[string][]float64{},
		LenQ:     0,
	}
}

func (e *Agent) SelectAction(obs []int) int {
	act := -1
	strObs := e.toStr(obs)

	e.checkAndAddObservation(strObs)

	if rand.Float64() < e.Epsilon {
		act = rand.Intn(e.Nact)
	} else {
		act = maxIdx(e.Q[strObs])
	}

	return act
}

func (e *Agent) Learn(obs []int, act int, rwd float64, done bool, next_obs []int) {
	strObs := e.toStr(obs)
	next_strObs := e.toStr(next_obs)

	e.checkAndAddObservation(next_strObs)

	target := float64(0)
	if done == true {
		target = rwd
	} else {
		target = rwd + e.Gamma*maxValue(e.Q[next_strObs])
	}

	e.Q[strObs][act] = (1-e.Alpha)*e.Q[strObs][act] + e.Alpha*target
}

func (e *Agent) GetQ(obs []int) []float64 {
	strObs := e.toStr(obs)

	Q := []float64{}

	if _, isExist := e.Q[strObs]; isExist {
		Q = e.Q[strObs]
	} else {
		Q = nil
	}

	return Q
}

func (e *Agent) toStr(obs []int) string {
	strs := make([]string, len(obs))

	for i, o := range obs {
		strs[i] = strconv.Itoa(o)
	}
	return strings.Join(strs, ",")
}

func (e *Agent) checkAndAddObservation(strObs string) {
	if _, isExist := e.Q[strObs]; !isExist {
		e.Q[strObs] = make([]float64, e.Nact)
		for i := 0; i < e.Nact; i++ {
			e.Q[strObs][i] = e.InitValQ
		}
		e.LenQ++
	}
}

func maxIdx(slice []float64) int {
	maxIndex := 0
	maxValue := slice[0]
	for i, v := range slice {
		if v > maxValue {
			maxValue = v
			maxIndex = i
		}
	}
	return maxIndex
}

func maxValue(slice []float64) float64 {
	maxValue := slice[0]
	for _, v := range slice {
		if v > maxValue {
			maxValue = v
		}
	}
	return maxValue
}
