package qlearn

import "math/rand"

type Environment struct {
	// constant
	IDBlank   int
	IDRobot   int
	IDCrystal int

	// variable
	NAct             int
	Done             bool
	FieldLength      int
	CrystalCandidate []int
	RwdFail          float64
	RwdMove          float64
	RwdCrystal       float64
	RobotPos         int
	CrystalPos       int
	RobotState       string
}

func NewEnvironment() Environment {
	return Environment{
		// constant
		IDBlank:   0,
		IDRobot:   1,
		IDCrystal: 2,

		// variable
		NAct:             2,
		Done:             false,
		FieldLength:      4,
		CrystalCandidate: []int{2, 3},
		RwdFail:          -1.0,
		RwdMove:          -1.0,
		RwdCrystal:       5.0,
		RobotPos:         -1, // -1 instead of nil
		CrystalPos:       -1,
		RobotState:       "",
	}
}

func (e Environment) reset() []int {
	e.Done = false
	e.RobotState = "normal"
	e.RobotPos = 0
	idx := rand.Intn(e.FieldLength)
}
