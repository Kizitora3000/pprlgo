package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"pprlgo/doublenc"
	"pprlgo/party"
	"pprlgo/qlearn"
	"strconv"
	"time"

	"github.com/tuneinsight/lattigo/v4/bfv"
	"github.com/tuneinsight/lattigo/v4/ckks"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

func main() {
	USE_BFV := true

	if USE_BFV {
		// BFVパラメータの設定 (128ビットセキュリティ、素数モジュラス)
		params, err := bfv.NewParametersFromLiteral(bfv.PN13QP218)
		if err != nil {
			panic(err)
		}

		// キージェネレータ、エンコーダ、暗号化器、評価器、復号化器の生成
		kgen := bfv.NewKeyGenerator(params)
		sk, pk := kgen.GenKeyPair()
		encoder := bfv.NewEncoder(params)
		encryptor := bfv.NewEncryptor(params, pk)
		rlk := kgen.GenRelinearizationKey(sk, 1)
		evaluator := bfv.NewEvaluator(params, rlwe.EvaluationKey{Rlk: rlk})
		decryptor := bfv.NewDecryptor(params, sk)
		privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		publicKey := &privateKey.PublicKey

		bfvKeyTools := party.BfvKeyTools{
			Params:     params,
			Encryptor:  encryptor,
			Decryptor:  decryptor,
			Encoder:    encoder,
			Evaluator:  evaluator,
			PrivateKey: privateKey,
			PublicKey:  publicKey,
		}

		// 整数値のエンコードと暗号化
		value1 := []uint64{5}
		value2 := []uint64{10}
		plaintext1 := bfv.NewPlaintext(params, params.MaxLevel())
		plaintext2 := bfv.NewPlaintext(params, params.MaxLevel())
		encoder.Encode(value1, plaintext1)
		encoder.Encode(value2, plaintext2)
		ciphertext1 := encryptor.EncryptNew(plaintext1)
		ciphertext2 := encryptor.EncryptNew(plaintext2)

		// 加算と乗算の実行
		addResult := evaluator.AddNew(ciphertext1, ciphertext2)
		mulResult := evaluator.MulNew(ciphertext1, ciphertext2)

		// 結果の復号化と表示
		decodedAddResult := encoder.DecodeUintNew(decryptor.DecryptNew(addResult))
		decodedMulResult := encoder.DecodeUintNew(decryptor.DecryptNew(mulResult))

		fmt.Println("加算結果:", decodedAddResult[0]) // 5 + 10 = 15
		fmt.Println("乗算結果:", decodedMulResult[0]) // 5 * 10 = 50

		Agt := qlearn.NewAgent()
		Agt.LenQ = 502

		dirname := "./preprocessed_diabetes_SRL_dataset"

		files, err := os.ReadDir(dirname)
		if err != nil {
			panic(err)
		}

		// クラウドのQ値を初期化
		encryptedQtable := make([]*rlwe.Ciphertext, Agt.LenQ)
		for i := 0; i < Agt.LenQ; i++ {
			plaintext := make([]uint64, Agt.Nact)
			for i := range plaintext {
				plaintext[i] = 0 // Agt.InitValQ
			}

			ciphertext := doublenc.BFVenc(params, encoder, encryptor, plaintext)
			encryptedQtable[i] = ciphertext
		}

		for _, file := range files {
			fmt.Println(file)
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
			var totalDuration time.Duration

			for i, record := range records {
				startTime := time.Now()

				status, _ := strconv.Atoi(record[1])
				action, _ := strconv.Atoi(record[2])
				rwd, _ := strconv.ParseFloat(record[3], 64)
				next_status, _ := strconv.Atoi(record[4])

				Agt.Learn_BFV(status, action, rwd, next_status, bfvKeyTools, encryptedQtable)

				duration := time.Since(startTime)
				totalDuration += duration
				// fmt.Println(duration) // 平均時間を計算
				fmt.Printf("file: %s\tindex:%d\ttime:%s\n", file.Name(), i, duration)

			}
		}
		return
	}

	/*
		params, err := ckks.NewParametersFromLiteral(
			ckks.ParametersLiteral{
				LogN:         7,
				LogQ:         []int{35, 60, 60},
				LogP:         []int{45, 45},
				LogSlots:     6,
				DefaultScale: 1 << 30,
			})
	*/
	///* security level 128
	params, err := ckks.NewParametersFromLiteral(
		ckks.ParametersLiteral{
			LogN:         13,                // 13
			LogQ:         []int{35, 60, 60}, // []int{55, 40, 40},
			LogP:         []int{45, 45},
			LogSlots:     6,
			DefaultScale: 1 << 30,
		})
	//*/
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

	keyTools := party.CkksKeyTools{
		Params:     params,
		Encryptor:  encryptor,
		Decryptor:  decryptor,
		Encoder:    encoder,
		Evaluator:  evaluator,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}

	Agt := qlearn.NewAgent()
	Agt.LenQ = 502

	dirname := "./preprocessed_diabetes_SRL_dataset"

	files, err := os.ReadDir(dirname)
	if err != nil {
		panic(err)
	}

	// クラウドのQ値を初期化
	encryptedQtable := make([]*rlwe.Ciphertext, Agt.LenQ)
	for i := 0; i < Agt.LenQ; i++ {
		plaintext := make([]float64, Agt.Nact)
		for i := range plaintext {
			plaintext[i] = Agt.InitValQ
		}

		ciphertext := doublenc.FHEenc(params, encoder, encryptor, plaintext)
		//hoge := doublenc.FHEdec(params, encoder, decryptor, ciphertext)
		//fmt.Println(len(hoge))
		encryptedQtable[i] = ciphertext
	}

	for _, file := range files {
		fmt.Println(file)
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
		var totalDuration time.Duration

		for i, record := range records {
			startTime := time.Now()

			status, _ := strconv.Atoi(record[1])
			action, _ := strconv.Atoi(record[2])
			rwd, _ := strconv.ParseFloat(record[3], 64)
			next_status, _ := strconv.Atoi(record[4])

			Agt.Learn(status, action, rwd, next_status, keyTools, encryptedQtable)

			duration := time.Since(startTime)
			totalDuration += duration
			// fmt.Println(duration) // 平均時間を計算
			fmt.Printf("file: %s\tindex:%d\ttime:%s\n", file.Name(), i, duration)

		}
	}

	Qtable := [][]float64{}
	for i := 0; i < Agt.LenQ; i++ {
		plaintext := doublenc.FHEdec(params, encoder, decryptor, encryptedQtable[i])

		plaintext_real := make([]float64, Agt.Nact)
		for j, v := range plaintext {
			if j == Agt.Nact {
				continue
			}
			plaintext_real[j] = real(v)
		}
		Qtable = append(Qtable, plaintext_real)
	}

	jsonData, err := json.Marshal(Qtable)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ioutil.WriteFile("pprl_data.json", jsonData, 0644)
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
