package qlearn

import "fmt"

/*
真値について

	真値は「強化学習を学びたい人が最初に読む本」のサンプルコード
	(URL: https://info.nikkeibp.co.jp/media/NSW/atcl/books/111100051/)
	における iRL/main/agt_tableQ.py の実行結果を使用している

真値の見方

	行動: 0列目: 後ろへ進む, 1列目: 前へ進む
	状態: 0: なにもない, 1: ロボットの位置, 2: クリスタルの位置
*/
var (
	true_Qtable = []map[string][]int{
		{
			"1020": {-1, 3},
        	"0120": {-1, 4},
			"0010": {5, -2}
       		"0021": {-1, -1}
       		"1002": {-1, 2}
        	"0102": {-1, 3}
        	"0012": {-1, 4}
        	"0001": {5, -1}
		}
	}
	obss = []string{
        "1020",
        "0120",
        "0010",
        "0021",
        "1002",
        "0102",
        "0012",
        "0001",
	}
)

func printResult(agt Agt, obss []map[string]float64) {
	println("result")
	for i, obs := range(obss) {
		fmt.Printf("[%s] true: %f, %f - result: %f, %f", 
			obs,
			true_Qtable[obs][0], true_Qtable[obs][0],
			agt.Q[obs][0], agt.Q[obs][1])
	}
}

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
