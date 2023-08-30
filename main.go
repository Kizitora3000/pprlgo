package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"pprlgo/doublenc"
	"pprlgo/party"
	"pprlgo/qlearn"

	"github.com/tuneinsight/lattigo/v4/ckks"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

var EncryptedQtable []*rlwe.Ciphertext

func main() {
	params, err := ckks.NewParametersFromLiteral(
		ckks.ParametersLiteral{
			LogN:         12, // 14
			LogQ:         []int{55, 40, 40, 40, 40, 40, 40, 40},
			LogP:         []int{45, 45},
			LogSlots:     1,
			DefaultScale: 1 << 40,
		})
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

	Nstep := 5000
	CorridorEnv := qlearn.NewEnvironment()
	Agt := qlearn.NewAgent()

	obs := CorridorEnv.Reset()
	Agt.Reset()

	var encryptedQtable []*rlwe.Ciphertext
	for i := 0; i < Agt.LenQ; i++ {
		ciphertext := doublenc.FHEenc(params, encoder, encryptor, Agt.Q[i])
		encryptedQtable = append(encryptedQtable, ciphertext)
	}

	for i := 0; i < Nstep; i++ {
		fmt.Printf("───── %d ─────\n", i)

		println("Q candidates:")

		//start := time.Now()
		act := Agt.SelectAction(obs, keyTools, encryptedQtable)
		//elapsed := time.Since(start)
		//fmt.Printf("The function took %s to execute.\n", elapsed)

		rwd, done, next_obs := CorridorEnv.Step(act)

		println("true Qnew:")

		//start = time.Now()
		Agt.Learn(obs, act, rwd, done, next_obs, keyTools, encryptedQtable)
		//elapsed = time.Since(start)
		//fmt.Printf("The function took %s to execute.\n", elapsed)

		println("Q table:")
		printEncryptedQtableForDebug(params, encoder, decryptor, encryptedQtable)

		obs = next_obs

		//fmt.Printf("The operation of %d took: %f[sec]\n", i, elapsed.Seconds())
	}

	for key, _ := range Agt.QKey {
		fmt.Printf("%s: %f\n", key, Agt.Q[Agt.QKey[key]])
	}
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
