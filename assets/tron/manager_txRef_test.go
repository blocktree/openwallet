/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package tron

import (
	"encoding/hex"
	"fmt"
	"github.com/blocktree/openwallet/assets/tron/grpc-gateway/core"
	"testing"
)

var (
	pTxRaw    = "0a7d0a02d0762208239cf236e19b41cf40e887c8a7e12c5a66080112620a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412310a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518904e"
	pTxSigned = "0a7d0a02d0762208239cf236e19b41cf40e887c8a7e12c5a66080112620a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412310a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518904e124123218536461fff784bd32e206a4c1a7adc05b455aa37cd624724cb3a2826119d434317d121bda9b5352bf0aaa61326471c5167d14376a96a466317d041696c6801"
)

func TestHash(t *testing.T) {
	txRaw := "0a7e0a0241222208c38f37b66624de47409096f8e1fa2c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a1541887661d2e0215851756b1e7933216064526badcd121541b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7a18a08d06"
	hash, _ := getTxHash1(txRaw)
	fmt.Println("hash:=", hex.EncodeToString(hash))
}

func TestGetbalance(t *testing.T) {
	addr := "TAJTMJuzvAqB8wmdUjRBVJW8CozfgrhpX3"

	addrBalance, _ := tw.Getbalance(addr)
	fmt.Println(addrBalance.TronBalance)
}

func TestCreateTransactionRef(t *testing.T) {

	if r, err := tw.CreateTransactionRef(TOADDRESS, OWNERADDRESS, AMOUNT); err != nil {
		t.Errorf("TestCreateTransaction failed: %v\n", err)
	} else {
		/*
		 t.Logf("TestCreateTransaction return: \n%+v\n", "Success!")

		 fmt.Println("")
		 fmt.Printf("APP Generated: %+v\n", "0a7e0a02adcd220873bf2dfd044f459440e0d4d2f8e12c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d")
		 fmt.Printf("Ref Generated: %+v\n\n", r)
		*/
		fmt.Println("Tx:=", r)
	}
}

/*
 func TestSignTransactoinRef(t *testing.T) {

	 var txRaw string
	 txRaw = "0a7e0a0241222208c38f37b66624de47409096f8e1fa2c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a1541887661d2e0215851756b1e7933216064526badcd121541b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7a18a08d06"
	 if r, err := tw.SignTransactionRef(txRaw, PRIVATEKEY); err != nil {
		 t.Errorf("SignTransactionRef failed: %v\n", err)
	 } else {
		 //t.Logf("SignTransactionRef return: \n\t%+v\n", r)
		 //debugPrintTx(r)
		 fmt.Println("signature:=", r)
	 }
 }
*/

func TestSignTransactionRef1(t *testing.T) {
	var txRaw string
	txRaw = "0a7e0a023d00220855e3f0172c91725e40c6d9a7aafc2c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a1541887661d2e0215851756b1e7933216064526badcd121541b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7a18a08d06"
	hash, err := getTxHash1(txRaw)
	if err != nil {
		t.Errorf("Get transaction hash failed:%v\n", err)
	}
	txHash := hex.EncodeToString(hash)
	if r, err := tw.SignTransactionRef(txHash, PRIVATEKEY); err != nil {
		t.Errorf("SignTransactionRef failed: %v\n", err)
	} else {
		fmt.Println("signature:=", r)
	}

}

func TestValidSignedTransactionRef(t *testing.T) {
	var txSignedRaw string
	//txSignedRaw = TXSIGNED
	txSignedRaw = "0a7e0a0289432208607606a10d2e670340e189ccfdfa2c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a1541887661d2e0215851756b1e7933216064526badcd121541b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7a18a08d06124100c8965a5d450bbeb263e852e6578556ba266338c99f887540c6e4ddb73488ef136316d72b36ff88770614a68810de5e2f2924aea0da5c9c5af724c71a34cfe100"
	//txSignedRaw = "0a7e0a02fd77220882256bb5fe08d39240d0a7c98fe82c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d1241373bf54b04e287d902beff4c6bd7369395b7b65527513922ee3b61ac0c4c6e8d0061da08b1b2f361e53c933360c3e5783996339431d44469f8bd57ee8fdfd3d700"
	if err := tw.ValidSignedTransactionRef(txSignedRaw); err != nil {
		t.Errorf("ValidSignedTransactionRef: %v\n", err)
	} else {
		t.Logf("GetTransactionSignRef return: \n\t%+v\n", "Success!")
	}
}

func TestBroadcastTransaction(t *testing.T) {

	var raw = "0a7e0a02a71f2208f4735d707a8537a240b0c2f088fb2c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a1541887661d2e0215851756b1e7933216064526badcd121541b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7a18a08d061241cb5ba106fafe57e2e3e6fc801a297aa5cc78d5c0b5358f97ceb5d903ab1e487b213dbb66195c992de2c7d9da5b13ace9880bc32141f28430ab7893291f06056f01"

	if txid, err := tw.BroadcastTransaction(raw, core.Transaction_Contract_TransferContract); err != nil {
		t.Errorf("BroadcastTransaction failed: %v\n", err)
	} else {
		t.Logf("BroadcastTransaction return: \n\t%+v\n", "Success!")
		fmt.Println("txid:=", txid)
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
	}
	println("Success!")
	println("------------------------------------------------------------------- Valid Done! \n")

	// txSignedRaw := "0a7e0a02b73b2208d971acfd3452661c40f0d68cfce12c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d1241166ae365792c1918de963cc4121e47417252da11d54783dbeb248f913240f27ea02b1b42f807c4ffb5d7ebecf687f5294400281021e6fefd0f38c50765f9c87200"
	txid, err := tw.BroadcastTransaction(txSignedRaw, core.Transaction_Contract_TransferContract)
	if err != nil {
		t.Errorf("ValidSignedTransactionRef: %v\n", err)
	} else {
		println("Success!")
		fmt.Println("txid:=", txid)
	}
	println("------------------------------------------------------------------- Boradcast! \n")

	//tx := &core.Transaction{}
	//txRawBytes, _ := hex.DecodeString(txRaw)
	//proto.Unmarshal(txRawBytes, tx)
	//txIDBytes, _ := getTxHash(tx)
	//_ := hex.EncodeToString(txIDBytes)

	//for i := 0; i < 1000; i++ {
	//
	//	tx, _ := tw.GetTransactionByID(txID)
	//	fmt.Println("Is Success: ", tx.IsSuccess)
	//	fmt.Println("")
	//
	//	if tx.IsSuccess {
	//		return
	//	}
	//
	//	time.Sleep(time.Second * 1)
	//}

}
