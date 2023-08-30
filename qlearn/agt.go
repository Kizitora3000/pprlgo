package qlearn

import (
	"fmt"
	"math/rand"
	"pprlgo/doublenc"
	"pprlgo/party"
	"pprlgo/pprl"
	"strconv"
	"strings"

	"github.com/tuneinsight/lattigo/v4/rlwe"
)

type Agent struct {
	Nact     int
	InitValQ float64
	Epsilon  float64
	Alpha    float64
	Gamma    float64
	Q        [][]float64
	QKey     map[string]int
	LenQ     int
}

func NewAgent() *Agent {
	return &Agent{
		Nact:     2,
		InitValQ: 0,
		Epsilon:  1.0,
		Alpha:    0.1,
		Gamma:    0.9,
		Q:        [][]float64{},
		QKey:     map[string]int{},
		LenQ:     0,
	}
}

func (e *Agent) Reset() {
	e.LenQ = 9

	e.Q = make([][]float64, e.LenQ)
	for i := 0; i < e.LenQ; i++ {
		e.Q[i] = make([]float64, e.Nact)
		for j := 0; j < e.Nact; j++ {
			e.Q[i][j] = e.InitValQ
		}
	}

	e.QKey["1,0,2,0"] = 0
	e.QKey["0,1,2,0"] = 1
	e.QKey["0,0,1,0"] = 2
	e.QKey["0,0,2,1"] = 3
	e.QKey["1,0,0,2"] = 4
	e.QKey["0,1,0,2"] = 5
	e.QKey["0,0,1,2"] = 6
	e.QKey["0,0,0,1"] = 7
	e.QKey["9,9,9,9"] = 8
}

func (e *Agent) SelectAction(obs []int, keyTools party.KeyTools, encryptedQtable []*rlwe.Ciphertext) int {
	act := -1
	strObs := e.toStr(obs)
	obsIdx := e.QKey[strObs]

	v_t := make([]float64, e.LenQ)
	v_t[obsIdx] = 1
	v_t_name := "v_t"

	pprl.SecureActionSelection(keyTools.Params, keyTools.Encoder, keyTools.Encryptor, keyTools.Decryptor, keyTools.Evaluator, keyTools.PublicKey, keyTools.PrivateKey, v_t, e.LenQ, e.Nact, v_t_name, encryptedQtable)
	Qs := doublenc.DEdec(keyTools.Params, keyTools.Encoder, keyTools.Decryptor, keyTools.PrivateKey, v_t_name)
	QsFloat64 := make([]float64, len(Qs))

	fmt.Println(QsFloat64)

	for i, v := range Qs {
		QsFloat64[i] = real(v)
	}

	if rand.Float64() < e.Epsilon {
		act = rand.Intn(e.Nact)
	} else {
		act = maxIdx(QsFloat64)
	}

	return act
}

func (e *Agent) Learn(obs []int, act int, rwd float64, done bool, next_obs []int, keyTools party.KeyTools, encryptedQtable []*rlwe.Ciphertext) {
	strObs := e.toStr(obs)
	next_strObs := e.toStr(next_obs)

	//e.checkAndAddObservation(next_strObs)

	target := float64(0)
	if done == true {
		target = rwd
	} else {
		target = rwd + e.Gamma*maxValue(e.Q[e.QKey[next_strObs]])
	}
	/*
		fmt.Print(" ")
		fmt.Print(e.QKey[next_strObs])
		fmt.Print(" ")
		fmt.Println(target)
	*/
	e.Q[e.QKey[strObs]][act] = (1-e.Alpha)*e.Q[e.QKey[strObs]][act] + e.Alpha*target

	obsIdx := e.QKey[strObs]
	Qnew := (1-e.Alpha)*e.Q[e.QKey[strObs]][act] + e.Alpha*target
	v_t := make([]float64, e.LenQ)
	w_t := make([]float64, e.Nact)
	v_t[obsIdx] = 1
	w_t[act] = 1
	fmt.Println(Qnew)
	pprl.SecureQtableUpdating(keyTools.Params, keyTools.Encoder, keyTools.Encryptor, keyTools.Decryptor, keyTools.Evaluator, keyTools.PublicKey, keyTools.PrivateKey, v_t, w_t, Qnew, e.LenQ, e.Nact, encryptedQtable)
}

func (e *Agent) GetQ(obs []int) []float64 {
	strObs := e.toStr(obs)

	Q := []float64{}
	Q = e.Q[e.QKey[strObs]]

	return Q
}

func (e *Agent) toStr(obs []int) string {
	strs := make([]string, len(obs))

	for i, o := range obs {
		strs[i] = strconv.Itoa(o)
	}
	return strings.Join(strs, ",")
}

/*
func (e *Agent) checkAndAddObservation(strObs string) {
	if _, isExist := e.Q[strObs]; !isExist {
		e.Q[strObs] = make([]float64, e.Nact)
		for i := 0; i < e.Nact; i++ {
			e.Q[strObs][i] = e.InitValQ
		}
		e.LenQ++
	}
}
*/

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
