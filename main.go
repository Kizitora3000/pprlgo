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

func SecureQtableUpdating(params ckks.Parameters, encoder ckks.Encoder, encryptor rlwe.Encryptor, decryptor rlwe.Decryptor, evaluator ckks.Evaluator, publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey, v_t []float64, w_t []float64, Q_new []float64) *rlwe.Ciphertext {
	
}

func SecureActionSelection(params ckks.Parameters, encoder ckks.Encoder, encryptor rlwe.Encryptor, decryptor rlwe.Decryptor, evaluator ckks.Evaluator, publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey, v_t []float64) *rlwe.Ciphertext {
	VtName := "VtName"

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

	return result
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
	res1 := SecureActionSelection(params, encoder, encryptor, decryptor, evaluator, publicKey, privateKey, v_t)
	fmt.Println(doublenc.FHEdec(params, encoder, decryptor, res1))

	w_t := []float64{0, 0, 1, 0}
	Q_new := []float64{3.5}

	res2 := SecureQtableUpdating(params, encoder, encryptor, decryptor, evaluator, publicKey, privateKey, v_t, w_t, Q_new)
}
