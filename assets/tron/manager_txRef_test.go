/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package tron

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/tronprotocol/grpc-gateway/core"
)

var (
	pTxRaw    = "0a7d0a02d0762208239cf236e19b41cf40e887c8a7e12c5a66080112620a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412310a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518904e"
	pTxSigned = "0a7d0a02d0762208239cf236e19b41cf40e887c8a7e12c5a66080112620a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412310a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518904e124123218536461fff784bd32e206a4c1a7adc05b455aa37cd624724cb3a2826119d434317d121bda9b5352bf0aaa61326471c5167d14376a96a466317d041696c6801"
)

func TestCreateTransactionRef(t *testing.T) {

	if r, err := tw.CreateTransactionRef(TOADDRESS, OWNERADDRESS, AMOUNT); err != nil {
		t.Errorf("TestCreateTransaction failed: %v\n", err)
	} else {
		t.Logf("TestCreateTransaction return: \n%+v\n", "Success!")

		fmt.Println("")
		fmt.Printf("APP Generated: %+v\n", "0a7e0a02adcd220873bf2dfd044f459440e0d4d2f8e12c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d")
		fmt.Printf("Ref Generated: %+v\n\n", r)
	}
}

func TestSignTransactoinRef(t *testing.T) {

	var txRaw string
	txRaw = "0a7e0a02055122080d595b1696607c7e40f8baf4b3e72c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d"
	if r, err := tw.SignTransactionRef(txRaw, PRIVATEKEY); err != nil {
		t.Errorf("SignTransactionRef failed: %v\n", err)
	} else {
		t.Logf("SignTransactionRef return: \n\t%+v\n", r)
		debugPrintTx(r)
	}
}

func TestValidSignedTransactionRef(t *testing.T) {
	var txSignedRaw string
	txSignedRaw = TXSIGNED
	txSignedRaw = "0a7e0a02adcd220873bf2dfd044f459440e0d4d2f8e12c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d1241f5b8eaac6034f590b54d7d1e2fcd588c56573c113ec7e98aac4a393747ae290e55f1bc2e861cc9dde18ac48e9594054632f4a1da2491bf091c2fe813f4e373d201"
	txSignedRaw = "0a7e0a02fd77220882256bb5fe08d39240d0a7c98fe82c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d1241373bf54b04e287d902beff4c6bd7369395b7b65527513922ee3b61ac0c4c6e8d0061da08b1b2f361e53c933360c3e5783996339431d44469f8bd57ee8fdfd3d700"

	if err := tw.ValidSignedTransactionRef(txSignedRaw); err != nil {
		t.Errorf("ValidSignedTransactionRef: %v\n", err)
	} else {
		t.Logf("GetTransactionSignRef return: \n\t%+v\n", "Success!")
	}
}

func TestSuiteTx(t *testing.T) {

	println("Start testsuit...\n")

	var (
		txRaw, txSignedRaw string
		err                error
	)

	txRaw, err = tw.CreateTransactionRef(TOADDRESS, OWNERADDRESS, AMOUNT)
	if err != nil {
		t.Errorf("TestCreateTransaction failed: %v\n", err)
		return
	}
	println("txRaw: ", txRaw)
	println("------------------------------------------------------------------- Create Done! \n")

	// txRaw = "0a7e0a02b2302208e384c56f2822541840e0f2fed1e82c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d"
	txSignedRaw, err = tw.SignTransactionRef(txRaw, PRIVATEKEY)
	if err != nil {
		t.Errorf("GetTransactionSignRef failed: %v\n", err)
		return
	}
	println("txSignedRaw: ", txSignedRaw)
	println("------------------------------------------------------------------- Sign Done! \n")

	// txSignedRaw = "0a7e0a02b20722088b88797132e6dc8540b09af7d1e82c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d12415407e6344baced1f933762ab31d1054906b4307a63fc5b3764a3ac803a3c3ceeabe1ac40d2947a8eaf98747de88e0e982cbb5f083d3e798c76280e170657303700"
	err = tw.ValidSignedTransactionRef(txSignedRaw)
	if err != nil {
		t.Errorf("ValidSignedTransactionRef: %v\n", err)
		return
	} else {
		println("Success!")
	}
	println("------------------------------------------------------------------- Valid Done! \n")

	// txSignedRaw := "0a7e0a02b73b2208d971acfd3452661c40f0d68cfce12c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d1241166ae365792c1918de963cc4121e47417252da11d54783dbeb248f913240f27ea02b1b42f807c4ffb5d7ebecf687f5294400281021e6fefd0f38c50765f9c87200"
	err = tw.BroadcastTransaction(txSignedRaw)
	if err != nil {
		t.Errorf("ValidSignedTransactionRef: %v\n", err)
	} else {
		println("Success!")
	}
	println("------------------------------------------------------------------- Boradcast! \n")

	tx := &core.Transaction{}
	txRawBytes, _ := hex.DecodeString(txRaw)
	proto.Unmarshal(txRawBytes, tx)
	txIDBytes, _ := getTxHash(tx)
	txID := hex.EncodeToString(txIDBytes)

	for i := 0; i < 1000; i++ {

		_, isSuccess, _ := tw.GetTransactionByID(txID)
		fmt.Println("Is Success: ", isSuccess)
		fmt.Println("")

		if isSuccess {
			return
		}

		time.Sleep(time.Second * 1)
	}

}
