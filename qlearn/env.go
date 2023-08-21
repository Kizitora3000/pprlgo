package qlearn

type Environment struct {
	ID_blank   int
	ID_robot   int
	ID_crystal int

	n_act             int
	done              bool
	field_length      int
	crystal_candidate []int
	rwd_fail          float64
	rwd_move          float64
	rwd_crystal       float64
}

func NewEnvironment() *Environment {
	env := &Environment{
		n_act:             2,
		done:              false,
		field_length:      4,
		crystal_candidate: []int{2, 3},
		rwd_fail:          -1.0,
		rwd_move:          -1.0,
		rwd_crystal:       5.0,
	}
	return env
}

func (e Environment) reset() []int {
	return []int{0, 0, 0, 0}
}
