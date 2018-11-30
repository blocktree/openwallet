package ontology

import (
	"fmt"
	"testing"
)

func Test_01(t *testing.T) {
	path := "balance"
	method := "GET"

	result, err := tw.LocalClient.Call(path, nil, method)

	if err != nil {
		t.Error(err)
	} else {
		fmt.Println(result)
	}
}

func Test_getBalanceByLocal(t *testing.T) {
	address := "ATfZt5HAHrx3Xmio3Ak9rr23SyvmgNVJqU"

	balance, err := tw.getBalanceByLocal(address)

	if err != nil {
		t.Error("get balance by local failed!")
	} else {
		fmt.Println("ONT: ", balance.ONTBalance)
		fmt.Println("ONG: ", balance.ONGBalance)
		fmt.Println("ONGUnbound: ", balance.ONGUnbound)
	}
}

func Test_getBlockHeightByLocal(t *testing.T) {
	height, err := tw.getBlockHeightByLocal()

	if err != nil {
		t.Error("get current block height failed!")
	} else {
		fmt.Println("current block height: ", height)
	}
}

func Test_getBlockHashByLocal(t *testing.T) {
	hash, err := tw.getBlockHashByLocal(41919)
	if err != nil {
		t.Error("get block hash failed!")
	} else {
		fmt.Println("block hash is  :", hash)
	}
}
