package party

import (
	"crypto/rsa"

	"github.com/tuneinsight/lattigo/v4/ckks"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

type User struct {
	Actions []float64
	States  []float64
	Alpha   float64
	Gamma   float64
	Qtable  [][]float64
}

type CloudPlatform struct {
	Qtable [][]float64
}

type KeyTools struct {
	Params     ckks.Parameters
	Encryptor  rlwe.Encryptor
	Decryptor  rlwe.Decryptor
	Encoder    ckks.Encoder
	Evaluator  ckks.Evaluator
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}
