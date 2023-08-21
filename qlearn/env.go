package qlearn

type Environment struct {
}

func (e Environment) reset() []int {
	return []int{0, 0, 0, 0}
}
