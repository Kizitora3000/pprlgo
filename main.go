package main

import (
	"fmt"
	"pprlgo/qlearn"
)

func main() {
	CorridorEnv := qlearn.NewEnvironment()
	fmt.Println(CorridorEnv)

	obs := CorridorEnv.Reset()
	fmt.Println(obs)
	fmt.Println(CorridorEnv)
	fmt.Println(CorridorEnv.Step(1))
	fmt.Println(CorridorEnv.Step(1))
	fmt.Println(CorridorEnv.Step(1))
	fmt.Println(CorridorEnv.Step(0))
}

/*
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"pprlgo/doublenc"
	"pprlgo/party"

	"github.com/tuneinsight/lattigo/v4/ckks"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

var (
	user party.User = party.User{
		Actions: []float64{0.1, 0.2, 0.3, 0.4},
		States:  []float64{1.0, 2.0, 3.0},
		Alpha:   0.5,
		Gamma:   0.9,
		Qtable: [][]float64{
			{0.1, 0.2, 0.3, 0.4},
			{0.5, 0.6, 0.7, 0.8},
			{0.9, 1.0, 1.1, 1.2},
		},
	}
	cloud_platform party.CloudPlatform = party.CloudPlatform{
		Qtable: [][]float64{
			{0.1, 0.2, 0.3, 0.4},
			{0.5, 0.6, 0.7, 0.8},
			{0.9, 1.0, 1.1, 1.2},
		},
	}
	EncryptedQtable []*rlwe.Ciphertext
)

func SecureQtableUpdating(params ckks.Parameters, encoder ckks.Encoder, encryptor rlwe.Encryptor, decryptor rlwe.Decryptor, evaluator ckks.Evaluator, publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey, v_t []float64, w_t []float64, Q_new float64) {
	WtName := "WtName"
	VtName := "VtName"

	doublenc.DEenc(params, encoder, encryptor, publicKey, w_t, WtName)

	for i := 0; i < len(cloud_platform.Qtable); i++ {
		filename := fmt.Sprintf(VtName+"_%d", i)
		if v_t[i] == 0 {
			zeros := make([]float64, len(cloud_platform.Qtable[i]))
			doublenc.DEenc(params, encoder, encryptor, publicKey, zeros, filename)
		} else if v_t[i] == 1 {
			ones := make([]float64, len(cloud_platform.Qtable[i]))
			for i := range ones {
				ones[i] = 1
			}
			doublenc.DEenc(params, encoder, encryptor, publicKey, ones, filename)
		}
	}

	Q_news := make([]float64, len(cloud_platform.Qtable[0]))
	Q_news_name := "Q_news_name"
	for i := range Q_news {
		Q_news[i] = Q_new
	}
	doublenc.DEenc(params, encoder, encryptor, publicKey, Q_news, Q_news_name)

	// v_and_w = Vt[i] * Wt
	// Qtable[i] += Q_new * (v_and_w) - Qtable[i] * (v_and_w)
	fhe_Q_news := doublenc.RSAdec(privateKey, Q_news_name)
	for i := 0; i < len(cloud_platform.Qtable); i++ {
		filename := fmt.Sprintf(VtName+"_%d", i)
		fhe_v_t := doublenc.RSAdec(privateKey, filename)
		fhe_w_t := doublenc.RSAdec(privateKey, WtName)

		// make Qnew
		v_and_w_Qnew := make([]float64, len(cloud_platform.Qtable[i]))
		fhe_v_and_w_Qnew := doublenc.FHEenc(params, encoder, encryptor, v_and_w_Qnew)
		evaluator.Mul(fhe_v_t, fhe_w_t, fhe_v_and_w_Qnew)
		evaluator.Relinearize(fhe_v_and_w_Qnew, fhe_v_and_w_Qnew)
		evaluator.Mul(fhe_v_and_w_Qnew, fhe_Q_news, fhe_v_and_w_Qnew)

		// make Qold
		v_and_w_Qold := make([]float64, len(cloud_platform.Qtable[i]))
		fhe_v_and_w_Qold := doublenc.FHEenc(params, encoder, encryptor, v_and_w_Qold)
		evaluator.Mul(fhe_v_t, fhe_w_t, fhe_v_and_w_Qold)
		evaluator.Relinearize(fhe_v_and_w_Qold, fhe_v_and_w_Qold)
		evaluator.Mul(fhe_v_and_w_Qold, EncryptedQtable[i], fhe_v_and_w_Qold)

		evaluator.Relinearize(fhe_v_and_w_Qnew, fhe_v_and_w_Qnew)
		evaluator.Relinearize(fhe_v_and_w_Qold, fhe_v_and_w_Qold)

		decrypt_fhe_v_and_w_Qnew := doublenc.FHEdec(params, encoder, decryptor, fhe_v_and_w_Qnew)
		realValues1 := make([]float64, len(decrypt_fhe_v_and_w_Qnew))
		for i, v := range decrypt_fhe_v_and_w_Qnew {
			realValues1[i] = real(v)
		}
		re_fhe_v_and_w_Qnew := doublenc.FHEenc(params, encoder, encryptor, realValues1)

		decrypt_fhe_v_and_w_Qold := doublenc.FHEdec(params, encoder, decryptor, fhe_v_and_w_Qold)
		realValues2 := make([]float64, len(decrypt_fhe_v_and_w_Qold))
		for i, v := range decrypt_fhe_v_and_w_Qold {
			realValues2[i] = real(v)
		}
		re_fhe_v_and_w_Qold := doublenc.FHEenc(params, encoder, encryptor, realValues2)

		// EncryptedQtalbe[i]がノイズで爆発する
		evaluator.Add(EncryptedQtable[i], re_fhe_v_and_w_Qnew, EncryptedQtable[i])
		evaluator.Sub(EncryptedQtable[i], re_fhe_v_and_w_Qold, EncryptedQtable[i])
	}
}

func SecureActionSelection(params ckks.Parameters, encoder ckks.Encoder, encryptor rlwe.Encryptor, decryptor rlwe.Decryptor, evaluator ckks.Evaluator, publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey, v_t []float64, filename string) {
	VtName := "VtName"

	// 準同型演算のために縦行列を横に拡張する
	//[0,		[0, 0, 0, 0]
	// 1, ->  [1, 1, 1, 1]
	// 0]		[0, 0, 0, 0]

	for i := 0; i < len(cloud_platform.Qtable); i++ {
		filename := fmt.Sprintf(VtName+"_%d", i)
		if v_t[i] == 0 {
			zeros := make([]float64, len(cloud_platform.Qtable[i]))
			doublenc.DEenc(params, encoder, encryptor, publicKey, zeros, filename)
		} else if v_t[i] == 1 {
			ones := make([]float64, len(cloud_platform.Qtable[i]))
			for i := range ones {
				ones[i] = 1
			}
			doublenc.DEenc(params, encoder, encryptor, publicKey, ones, filename)
		}
	}

	zeros := make([]float64, len(cloud_platform.Qtable[0]))
	result := doublenc.FHEenc(params, encoder, encryptor, zeros)
	for i := 0; i < len(cloud_platform.Qtable); i++ {
		filename := fmt.Sprintf(VtName+"_%d", i)
		vt := doublenc.RSAdec(privateKey, filename)
		evaluator.Mul(vt, EncryptedQtable[i], vt)

		// The multiplicable depth is one, so Relinearize is used to reset depth.
		evaluator.Relinearize(vt, vt)
		evaluator.Add(result, vt, result)
	}

	doublenc.RSAenc(publicKey, result, filename)
	return
}

func main() {

	fmt.Println(user)
	fmt.Println(cloud_platform)

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
	for _, row := range cloud_platform.Qtable {
		ciphertext := doublenc.FHEenc(params, encoder, encryptor, row)
		encrypted_Qtable = append(encrypted_Qtable, ciphertext)
	}
	EncryptedQtable = encrypted_Qtable

	v_t := []float64{0, 1, 0}
	v_t_name := "v_t"
	SecureActionSelection(params, encoder, encryptor, decryptor, evaluator, publicKey, privateKey, v_t, v_t_name)
	fmt.Println(doublenc.DEdec(params, encoder, decryptor, privateKey, v_t_name))

	w_t := []float64{0, 0, 1, 0}
	Q_new := float64(3.5)

	println("--- previous ---")
	printEncryptedQtableForDebug(params, encoder, decryptor)
	SecureQtableUpdating(params, encoder, encryptor, decryptor, evaluator, publicKey, privateKey, v_t, w_t, Q_new)
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
*/
