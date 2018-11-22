package ontology

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/blocktree/OpenWallet/log"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
)

//私钥转私钥结构

//Algorithm: ECDSA
//curve type : elliptic.P256()
//schema:SHA256withECDSA

func MakeTxHash(gasPrice uint64, gasLimit uint64, asset string,
	from string, to string, amount uint64, payer string) ([]byte, error) {
	mutable, err := utils.TransferTx(gasPrice, gasLimit, asset, from, to, amount)
	if err != nil {
		log.Errorf("make transaction failed, err=%v", err)
		return nil, err
	}

	fromAddr, err := common.AddressFromBase58(from)
	if err != nil {
		log.Errorf("decode address from base58 failed, err=%v", err)
		return nil, err
	}

	if payer == "" {
		mutable.Payer = [20]byte(fromAddr)
	}

	txHash := mutable.Hash()
	hasher := crypto.SHA256.New()
	hasher.Write(txHash.ToArray())
	return hasher.Sum(nil), nil
}

func SignTxHash(hash []byte, priKey *ecdsa.PrivateKey) ([]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, priKey, hash)
	if err != nil {
		log.Errorf("sign ontology tx failed, err=%v", err)
		return nil, err
	}

	//var buf bytes.Buffer
	//buf.WriteByte(byte(keystore.SHA256withECDSA))

	size := (elliptic.P256().Params().BitSize + 7) >> 3
	res := make([]byte, size*2)
	copy(res[size-len(r.Bytes()):], r.Bytes())
	copy(res[size*2-len(s.Bytes()):], s.Bytes())
	//buf.Write(res)
	return res, nil
}

// func MakeRawTx(gasPrice uint64, gasLimit uint64, asset string,
// 	from string, to string, amount uint64, payer string, priKey *ecdsa.PrivateKe) ([]byte, error) {
// 	mutable, err := utils.TransferTx(gasPrice, gasLimit, asset, from, to, amount)
// 	if err != nil {
// 		log.Errorf("make transaction failed, err=%v", err)
// 		return nil, err
// 	}

// 	fromAddr, err := common.AddressFromBase58(from)
// 	if err != nil {
// 		log.Errorf("decode address from base58 failed, err=%v", err)
// 		return nil, err
// 	}

// 	if payer == "" {
// 		mutable.Payer = [20]byte(fromAddr)
// 	}

// 	txHash := mutable.Hash()
// 	hasher := crypto.SHA256.New()
// 	hasher.Write(txHash.ToArray())

// 	sig, err := SignTxHash(hasher.Sum(nil), priKey)
// 	if err != nil {
// 		log.Errorf("sign tx failed, err=%v", err)
// 		return nil, err
// 	}

// 	psig := types.MakeSigForBlockTree()
// 	mutable.Sigs = append(mutable.Sigs, *psig)
// }
