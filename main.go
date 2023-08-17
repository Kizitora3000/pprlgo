package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"pprlgo/doublenc"
	"pprlgo/party"

	"github.com/tuneinsight/lattigo/v4/ckks"
)

var user party.User = party.User{
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

var cloud_platform party.CloudPlatform = party.CloudPlatform{
	Qtable: [][]float64{
		{0.1, 0.2, 0.3, 0.4},
		{0.5, 0.6, 0.7, 0.8},
		{0.9, 1.0, 1.1, 1.2},
	},
}

func SecureActionSelection(u party.User, cp party.CloudPlatform) {
	for j := 0; j < len(u.Actions); j++ {
		for i := 0; i < len(u.States); i++ {

		}
	}
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

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := &privateKey.PublicKey

	/*
		v_t := []float64{0, 1.0, 0}
		c_v_t := doublenc.FHEenc(params, encoder, encryptor, v_t)
		p_v_t := doublenc.FHEdec(params, encoder, decryptor, c_v_t)

		fmt.Println(p_v_t)

		doublenc.RSAenc(publicKey, c_v_t, "cvt")
		hoge := doublenc.RSAdec(privateKey, "cvt")
		valuesTest := encoder.Decode(decryptor.DecryptNew(hoge), params.LogSlots())
		fmt.Printf("ValuesTest: %6.10f %6.10f %6.10f %6.10f...\n", valuesTest[0], valuesTest[1], valuesTest[2], valuesTest[3])
	*/

	v_t := []float64{0, 1.0, 0}
	doublenc.DEenc(params, encoder, encryptor, publicKey, v_t, "cvt")
	valuesTest := doublenc.DEdec(params, encoder, decryptor, privateKey, "cvt")
	fmt.Printf("ValuesTest: %6.10f %6.10f %6.10f %6.10f...\n", valuesTest[0], valuesTest[1], valuesTest[2], valuesTest[3])

	abc := doublenc.FHEenc(params, encoder, encryptor, cloud_platform.Qtable[0])
	bc := doublenc.FHEdec(params, encoder, decryptor, abc)
	fmt.Println(bc)
}
