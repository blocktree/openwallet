package keystore

import (
	"encoding/json"
	"testing"

	"github.com/blocktree/OpenWallet/log"
)

func TestAddressDecoder_NewAccount(t *testing.T) {
	testWallet, err := NewClientImpl("/Users/peter/workspace/bitcoin/wallet/src/github.com/ontio/ontology/tree/node/wallet_test.dat")

	acct, err := testWallet.NewAccount("peter1", PK_ECDSA, P256, SHA256withECDSA, []byte("123456"))
	if err != nil {
		t.Errorf("TestNewAccount error:%s\n", err)
		return
	}

	obj, _ := json.MarshalIndent(acct, "", " ")
	log.Debugf("new account:", string(obj))
}

func TestClientImpl_NewAccountWithHK(t *testing.T) {
	testWallet, err := NewClientImpl("/Users/peter/workspace/bitcoin/wallet/src/github.com/ontio/ontology/tree/node/wallet.dat")

	acct, err := testWallet.NewAccountWithHK("peter2", PK_ECDSA, P256, SHA256withECDSA, []byte("123456"))
	if err != nil {
		t.Errorf("NewAccountWithHK error:%s\n", err)
		return
	}

	obj, _ := json.MarshalIndent(acct, "", " ")
	log.Debugf("new account:", string(obj))
}
