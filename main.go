package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"pprlgo/doublenc"
	"pprlgo/pprl"
	"pprlgo/qlearn"

	"github.com/tuneinsight/lattigo/v4/ckks"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

var EncryptedQtable []*rlwe.Ciphertext

func main() {
	Nstep := 5000
	CorridorEnv := qlearn.NewEnvironment()
	Agt := qlearn.NewAgent()

	obs := CorridorEnv.Reset()
	Agt.Reset()
	for i := 0; i < Nstep; i++ {
		act := Agt.SelectAction(obs)

		rwd, done, next_obs := CorridorEnv.Step(act)

		Agt.Learn(obs, act, rwd, done, next_obs)

		obs = next_obs
	}

	for key, _ := range Agt.QKey {
		fmt.Printf("%s: %f\n", key, Agt.Q[Agt.QKey[key]])
	}

	params, err := ckks.NewParametersFromLiteral(
		ckks.ParametersLiteral{
			LogN:         14,
			LogQ:         []int{55, 40, 40, 40, 40, 40, 40, 40},
			LogP:         []int{45, 45},
			LogSlots:     2,
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

	// encrypt Qtalbe in CP
	var encrypted_Qtable []*rlwe.Ciphertext
	for _, row := range Agt.Q {
		ciphertext := doublenc.FHEenc(params, encoder, encryptor, row)
		encrypted_Qtable = append(encrypted_Qtable, ciphertext)
	}
	EncryptedQtable = encrypted_Qtable

	v_t := []float64{0, 1, 0}
	v_t_name := "v_t"
	pprl.SecureActionSelection(params, encoder, encryptor, decryptor, evaluator, publicKey, privateKey, v_t, Agt.LenQ, Agt.Nact, v_t_name)
	fmt.Println(doublenc.DEdec(params, encoder, decryptor, privateKey, v_t_name))

	w_t := []float64{0, 0, 1, 0}
	Q_new := float64(3.5)

	println("--- previous ---")
	printEncryptedQtableForDebug(params, encoder, decryptor)
	pprl.SecureQtableUpdating(params, encoder, encryptor, decryptor, evaluator, publicKey, privateKey, v_t, w_t, Q_new, Agt.LenQ, Agt.Nact)
	println("--- present ---")
	printEncryptedQtableForDebug(params, encoder, decryptor)
}

func printEncryptedQtableForDebug(params ckks.Parameters, encoder ckks.Encoder, decryptor rlwe.Decryptor) {
	for i, row := range EncryptedQtable {
		decryptedRow := doublenc.FHEdec(params, encoder, decryptor, row)
		fmt.Printf("Decrypted Qtable Row %d: ", i)
		for _, val := range decryptedRow {
			fmt.Printf("%f, ", real(val))
		}
		fmt.Println()
	}
}
