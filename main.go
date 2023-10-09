package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"pprlgo/doublenc"
	"pprlgo/qlearn"
	"strconv"

	"github.com/tuneinsight/lattigo/v4/ckks"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

var EncryptedQtable []*rlwe.Ciphertext

func main() {

	/*
		params, err := ckks.NewParametersFromLiteral(
			ckks.ParametersLiteral{
				LogN:         4,
				LogQ:         []int{35, 60, 60},
				LogP:         []int{45, 45},
				LogSlots:     1,
				DefaultScale: 1 << 30,
			})
		/* security level 128
		LogN:         13,                // 13
		LogQ:         []int{35, 60, 60}, // []int{55, 40, 40},
		LogP:         []int{45, 45},
		LogSlots:     1,
		DefaultScale: 1 << 30,
		if err != nil {
			panic(err)
		}

		kgen := ckks.NewKeyGenerator(params)
		sk := kgen.GenSecretKey()
		encryptor := ckks.NewEncryptor(params, sk)
		decryptor := ckks.NewDecryptor(params, sk)
		encoder := ckks.NewEncoder(params)
		rlk := kgen.GenRelinearizationKey(sk, 1)
		evaluator := ckks.NewEvaluator(params, rlwe.EvaluationKey{Rlk: rlk})
		privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		publicKey := &privateKey.PublicKey

		keyTools := party.KeyTools{
			Params:     params,
			Encryptor:  encryptor,
			Decryptor:  decryptor,
			Encoder:    encoder,
			Evaluator:  evaluator,
			PrivateKey: privateKey,
			PublicKey:  publicKey,
		}
	*/

	Agt := qlearn.NewAgent()

	dirname := "./proprocessed_diabetes_dataset"

	files, err := os.ReadDir(dirname)
	if err != nil {
		panic(err)
	}

	mx_status := 0.0
	for _, file := range files {
		filename := filepath.Join(dirname, file.Name())
		file, err := os.Open(filename)

		// open csv
		if err != nil {
			fmt.Printf("Error opening file %s: %v\n", filename, err)
			return
		}
		defer file.Close()

		r := csv.NewReader(file)
		records, err := r.ReadAll()
		if err != nil {
			fmt.Printf("Error reading CSV %s: %v\n", filename, err)
			return
		}

		// Exclude the last row
		records = records[:len(records)-1]

		for _, record := range records {
			status, _ := strconv.Atoi(record[1])
			action, _ := strconv.Atoi(record[2])
			rwd, _ := strconv.ParseFloat(record[3], 64)
			next_status, _ := strconv.Atoi(record[4])
			mx_status = math.Max(mx_status, float64(status))

			Agt.Learn(status, action, rwd, next_status)
		}
	}

	/*
		for key, values := range Agt.Q {
			fmt.Printf("Key: %s, Values: %v\n", key, values)
		}
	*/

	// mx_status × Nact
	Qtable := make([][]float64, int(mx_status))
	for i := range Qtable {
		Qtable[i] = make([]float64, Agt.Nact)
		for j := range Qtable[i] {
			Qtable[i][j] = Agt.InitValQ
		}
		if _, isExist := Agt.Q[i]; isExist {
			Qtable[i] = Agt.Q[i]
		}
	}

	jsonData, err := json.Marshal(Qtable)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ioutil.WriteFile("data.json", jsonData, 0644)
	if err != nil {
		fmt.Println(err)
	}

	/*

		Nstep := 50000

		numInstances := 4
		CorridorEnvs := make([]*qlearn.Environment, numInstances)
		Agents := make([]*qlearn.Agent, numInstances)
		obs := make([][]int, numInstances)

		for i := 0; i < numInstances; i++ {
			CorridorEnvs[i] = qlearn.NewEnvironment()
			Agents[i] = qlearn.NewAgent()
			obs[i] = CorridorEnvs[i].Reset()
			Agents[i].Reset()
		}

		// クラウドのQ値は最初のエージェントで初期化(全エージェント共通)
		var encryptedQtable []*rlwe.Ciphertext
		for i := 0; i < Agents[0].LenQ; i++ {
			ciphertext := doublenc.FHEenc(params, encoder, encryptor, Agents[0].Q[i])
			encryptedQtable = append(encryptedQtable, ciphertext)
		}

		// ---
		all_trial := 0.0
		num_success := 0.0

		file, err := os.Create("result.txt")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer file.Close()
		// ---

		for i := 0; i < Nstep; i++ {
			start := time.Now()
			fmt.Printf("───── %d ─────\n", i)

			println("Q candidates:")

			for j := 0; j < 1; j++ {
				act := Agents[j].SelectAction(obs[j], keyTools, encryptedQtable)

				rwd, done, next_obs := CorridorEnvs[j].Step(act)

				println("true Qnew:")

				Agents[j].Learn(obs[j], act, rwd, done, next_obs, keyTools, encryptedQtable)

				println("Q table:")
				printEncryptedQtableForDebug(params, encoder, decryptor, encryptedQtable)

				obs[j] = next_obs

				if done {
					if rwd == 5 {
						num_success++
					}
					all_trial++
					//fmt.Printf("%f, %f, %f\n", num_success, all_trial, num_success/all_trial)
					_, err = fmt.Fprintf(file, "%f,%f\n", all_trial, num_success/all_trial)

					if err != nil {
						fmt.Println("Error:", err)
						return
					}
				}
			}

			elapsed := time.Since(start)
			fmt.Printf("The operation of %d took: %f[sec]\n", i, elapsed.Seconds())
		}

		for i := 0; i < 4; i++ {
			for key, _ := range Agents[i].QKey {
				fmt.Printf("%s: %f\n", key, Agents[i].Q[Agents[i].QKey[key]])
			}
		}
	*/
}

func printEncryptedQtableForDebug(params ckks.Parameters, encoder ckks.Encoder, decryptor rlwe.Decryptor, encryptedQtable []*rlwe.Ciphertext) {
	for i, row := range encryptedQtable {
		decryptedRow := doublenc.FHEdec(params, encoder, decryptor, row)
		fmt.Printf("Decrypted Qtable Row %d: ", i)
		for _, val := range decryptedRow {
			fmt.Printf("%f, ", real(val))
		}
		fmt.Println()
	}
}
