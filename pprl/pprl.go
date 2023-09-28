package pprl

import (
	"crypto/rsa"
	"fmt"
	"pprlgo/doublenc"

	"github.com/tuneinsight/lattigo/v4/ckks"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

func SecureQtableUpdating(params ckks.Parameters, encoder ckks.Encoder, encryptor rlwe.Encryptor, decryptor rlwe.Decryptor, evaluator ckks.Evaluator, publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey, v_t []float64, w_t []float64, Q_new float64, Nv int, Na int, EncryptedQtable []*rlwe.Ciphertext) {
	WtName := "WtName"
	VtName := "VtName"

	//start := time.Now()
	doublenc.DEenc(params, encoder, encryptor, publicKey, w_t, WtName)

	for i := 0; i < Nv; i++ {
		filename := fmt.Sprintf(VtName+"_%d", i)
		if v_t[i] == 0 {
			zeros := make([]float64, Na)
			doublenc.DEenc(params, encoder, encryptor, publicKey, zeros, filename)
		} else if v_t[i] == 1 {
			ones := make([]float64, Na)
			for i := range ones {
				ones[i] = 1
			}
			doublenc.DEenc(params, encoder, encryptor, publicKey, ones, filename)
		}
	}

	Q_news := make([]float64, Na)
	Q_news_name := "Q_news_name"
	for i := range Q_news {
		Q_news[i] = Q_new
	}
	doublenc.DEenc(params, encoder, encryptor, publicKey, Q_news, Q_news_name)

	//elapsed := time.Since(start)
	//fmt.Printf("The function took %s to execute.\n", elapsed)

	//start = time.Now()

	// v_and_w = Vt[i] * Wt
	// Qtable[i] += Q_new * (v_and_w) - Qtable[i] * (v_and_w)
	fhe_Q_news := doublenc.RSAdec(privateKey, Q_news_name)
	for i := 0; i < Nv; i++ {
		filename := fmt.Sprintf(VtName+"_%d", i)
		fhe_v_t := doublenc.RSAdec(privateKey, filename)
		fhe_w_t := doublenc.RSAdec(privateKey, WtName)

		// make Qnew
		v_and_w_Qnew := make([]float64, Na)
		fhe_v_and_w_Qnew := doublenc.FHEenc(params, encoder, encryptor, v_and_w_Qnew)
		evaluator.Mul(fhe_v_t, fhe_w_t, fhe_v_and_w_Qnew)
		evaluator.Relinearize(fhe_v_and_w_Qnew, fhe_v_and_w_Qnew)
		evaluator.Mul(fhe_v_and_w_Qnew, fhe_Q_news, fhe_v_and_w_Qnew)

		// make Qold
		v_and_w_Qold := make([]float64, Na)
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

	//elapsed = time.Since(start)
	//fmt.Printf("The function took %s to execute.\n", elapsed)
}

func SecureActionSelection(params ckks.Parameters, encoder ckks.Encoder, encryptor rlwe.Encryptor, decryptor rlwe.Decryptor, evaluator ckks.Evaluator, publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey, v_t []float64, Nv int, Na int, filename string, EncryptedQtable []*rlwe.Ciphertext) {
	VtName := "VtName"

	// 準同型演算のために縦行列を横に拡張する
	//[0,		[0, 0, 0, 0]
	// 1, ->  [1, 1, 1, 1]
	// 0]		[0, 0, 0, 0]

	for i := 0; i < Nv; i++ {
		filename := fmt.Sprintf(VtName+"_%d", i)
		if v_t[i] == 0 {
			zeros := make([]float64, Na)
			doublenc.DEenc(params, encoder, encryptor, publicKey, zeros, filename)
		} else if v_t[i] == 1 {
			ones := make([]float64, Na)
			for i := range ones {
				ones[i] = 1
			}
			doublenc.DEenc(params, encoder, encryptor, publicKey, ones, filename)
		}
	}

	zeros := make([]float64, Na)
	result := doublenc.FHEenc(params, encoder, encryptor, zeros)
	for i := 0; i < Nv; i++ {
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
