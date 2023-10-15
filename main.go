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

	"github.com/tuneinsight/lattigo/v4/ckks"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

func main() {
	params, err := ckks.NewParametersFromLiteral(
		ckks.ParametersLiteral{
			LogN:         7,
			LogQ:         []int{35, 60, 60},
			LogP:         []int{45, 45},
			LogSlots:     6,
			DefaultScale: 1 << 30,
		})
	/* security level 128
	params, err := ckks.NewParametersFromLiteral(
		ckks.ParametersLiteral{
	LogN:         13,                // 13
	LogQ:         []int{35, 60, 60}, // []int{55, 40, 40},
	LogP:         []int{45, 45},
	LogSlots:     1,
	DefaultScale: 1 << 30,
	})
	*/
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

	Agt := qlearn.NewAgent()
	Agt.LenQ = 502

	dirname := "./preprocessed_diabetes_RL_dataset"

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
		if _, isExist := Agt.Q[i]; !isExist {
			Agt.Q[i] = make([]float64, Agt.Nact)
			for j := 0; j < Agt.Nact; j++ {
				Agt.Q[i][j] = Agt.InitValQ
			}
		}

		Qtable = append(Qtable, Agt.Q[i])
	}

	jsonData, err := json.Marshal(Qtable)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ioutil.WriteFile("rl_data.json", jsonData, 0644)
	if err != nil {
		fmt.Println(err)
	}
}
