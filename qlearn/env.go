package qlearn

type Env struct {
}

func (e Env) reset() []int {
	return []int{0, 0, 0, 0}
}
