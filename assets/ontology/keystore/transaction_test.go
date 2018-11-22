package keystore

import (
	"fmt"
	"testing"

	"github.com/blocktree/OpenWallet/log"
	"github.com/ontio/ontology/cmd/utils"
)

// type PrivateKeyAdapter PrivateKey

// func (this *PrivateKeyAdapter) Public() crypto.PublicKey {
// 	return &PublicKey{Algorithm: this.Algorithm, PublicKey: &this.PublicKey}
// }

func Test_Transfer(t *testing.T) {
	testWallet, err := NewClientImpl("/Users/peter/workspace/bitcoin/wallet/src/github.com/ontio/ontology/tree/node/wallet.dat")
	if err != nil {
		t.Errorf("open wallet failed, err=%v", err)
		return
	}

	from, err := GetAccountMulti(testWallet, []byte("new.1234"), "")
	if err != nil {
		t.Errorf("get from account failed, err=%v", err)
		return
	}

	// to, err := GetAccountMulti(testWallet, []byte("123456"), "ATk9xPuYpXcMgKoevxwt7Q8Jwnm4N7xfvV")
	// if err != nil{
	// 	t.Errorf("get to account failed, err=%v", err)
	// 	return
	// }
	var gasPrice uint64 = 500
	var gasLimit uint64 = 30000

	//p := PrivateKeyAdapter(*from.PrivateKey)
	a := utils.MakeAccountForBlockTree(byte(from.PrivateKey.Algorithm), from.PrivateKey.PrivateKey, from.PublicKey.PublicKey, from.Address, byte(from.SigScheme))
	// a := account.Account{
	// 	PrivateKey: &p,
	// 	PublicKey:  from.PublicKey,
	// 	Address:    common.Address(from.Address),
	// 	SigScheme:  s.SignatureScheme(from.SigScheme),
	// }

	log.Debugf("begin test, from:%v", from.Address.ToBase58())
	fmt.Println("begin test2")
	txHash, err := utils.Transfer(gasPrice, gasLimit, a, "ont", from.Address.ToBase58(), "ATk9xPuYpXcMgKoevxwt7Q8Jwnm4N7xfvV", 100)
	if err != nil {
		t.Errorf("transfer error:%s", err)
		return
	}

	log.Debugf("transaction:%v", txHash)
}
