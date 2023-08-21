package qlearn

func learn() {
	n_step := 5000
	agt := Agt{}
	env := Env{}

	obs := env.reset()
	for i := 0; i < n_step; i++ {
		act := agt.select_action(obs)
		rwd, done, next_obs := env.step(act)

		agt.learn(obs, act, rwd, done, next_obs)

		obs = next_obs
	}

	printResult(agt, obs)
}
