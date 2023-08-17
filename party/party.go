package party

type User struct {
	Actions []float64
	States  []float64
	Alpha   float64
	Gamma   float64
	Qtable  [][]float64
}

type CloudPlatform struct {
	Qtable [][]float64
}
