package qlearn

type Agent struct {
	n_act      int
	init_val_Q int
	epsilon    float64
	alpha      float64
	gamma      float64
	max_memory int
	filepath   string
}
