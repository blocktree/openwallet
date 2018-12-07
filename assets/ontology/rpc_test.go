package ontology

import (
	"fmt"
	"testing"
)

func Test_getBalanceByRest(t *testing.T) {
	address := "ATfZt5HAHrx3Xmio3Ak9rr23SyvmgNVJqU"

	balance, err := tw.RPCClient.getBalance(address)

	if err != nil {
		t.Error("get balance by local failed!")
	} else {
		fmt.Println("ONT: ", balance.ONTBalance)
		fmt.Println("ONG: ", balance.ONGBalance)
		fmt.Println("ONGUnbound: ", balance.ONGUnbound)
	}
}

func Test_getBlockByRest(t *testing.T) {
	hash := "2e6954676d4ba75a5e9875536e419177dbd1df3ca3548f6537d26ed4ed43e55b"
	ret, err := tw.RPCClient.getBlock(hash)

	if err != nil {
		t.Error("get current block height failed!")
	} else {
		fmt.Println("current block height: ", ret)
	}
}

func Test_getBlockHeightFromTxIDByRest(t *testing.T) {
	txid := "dc55118ac9442af38a0ec85bcce54a8f8d68ba65de0120a8739d90b9d93b6ca2"

	height, err := tw.RPCClient.getBlockHeightFromTxID(txid)

	if err != nil {
		t.Error("get current block height failed!")
	} else {
		fmt.Println("current block height: ", height)
	}
}

func Test_getBlockHeightByRest(t *testing.T) {
	height, err := tw.RPCClient.getBlockHeight()

	if err != nil {
		t.Error("get current block height failed!")
	} else {
		fmt.Println("current block height: ", height)
	}
}

func Test_getBlockHashByRest(t *testing.T) {
	hash, err := tw.RPCClient.getBlockHash(41919)
	if err != nil {
		t.Error("get block hash failed!")
	} else {
		fmt.Println("block hash is  :", hash)
	}
}
