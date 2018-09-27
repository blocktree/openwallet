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
	"testing"
)

var (
	pTxRaw    = "0a7d0a02d0762208239cf236e19b41cf40e887c8a7e12c5a66080112620a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412310a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518904e"
	pTxSigned = "0a7d0a02d0762208239cf236e19b41cf40e887c8a7e12c5a66080112620a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412310a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518904e124123218536461fff784bd32e206a4c1a7adc05b455aa37cd624724cb3a2826119d434317d121bda9b5352bf0aaa61326471c5167d14376a96a466317d041696c6801"
)

func TestCreateTransactionRef(t *testing.T) {

	var to_address, owner_address string
	var amount int64

	// predictTxRaw := "0a7e0a021031220816b0c1a29ce3387c40e099ad83e02c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a1541b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7a12154199fee02e1ee01189bc41a68e9069b7919ef2ad8218c0843d"
	// to_address, owner_address, amount = "TQ1TiUzStbSLdEtACUDmzfMDpWUyo8cyCf", "TSdXzXKSQ3RQzQ5Ge8TiYfMQEjofSVQ8ax", uint64(1)
	to_address, owner_address, amount = TOADDRESS, OWNERADDRESS, AMOUNT

	if r, err := tw.CreateTransactionRef(to_address, owner_address, amount); err != nil {
		t.Errorf("TestCreateTransaction failed: %v\n", err)
	} else {
		// if strings.Join(r[:], "") != RAW_expect {
		// 	t.Errorf("TestCreateTransaction return invalid RAW!")
		// }
		t.Logf("TestCreateTransaction return: \n\t%+v\n", r)
		// fmt.Println("P_Raw: ", predictTxRaw)
		// fmt.Println("R_Raw: ", r)
	}
}

func TestGetTransactoinSignRef(t *testing.T) {

	var txRaw string
	txRaw = "0a7e0a02efac220835a56bead7df0fca40a8b8ffb2e12c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d"
	txRaw = ""

	if r, err := tw.GetTransactionSignRef(txRaw, PRIVATEKEY); err != nil {
		t.Errorf("GetTransactionSignRef failed: %v\n", err)
	} else {
		// if strings.Join(r[:], "") != RAW_expect {
		// 	t.Errorf("TestCreateTransaction return invalid RAW!")
		// }
		t.Logf("GetTransactionSignRef return: \n\t%+v\n", r)
	}
}

func TestValidSignedTransactionRef(t *testing.T) {
	var txSignedRaw string
	txSignedRaw = TXSIGNED
	txSignedRaw = "0a7e0a02eabc2208ffa4851bca447d8340a8ff97b1e12c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d1241ee734149bcbc09fd23f109cfbad7d301898ae04c51633894f83c8c8e1796fdcc995df9dfc0cf1018f10f95e40cd4109dd5e5228f0ba2ed51dee7069b75b87e8e00"
	txSignedRaw = "0a5a0a02b82822082c12208785a174144096bc8dc2f6b0feab155a360802123212300a1541e11973395042ba3c0b52b4cdf4e15ea77818f27512154199fee02e1ee01189bc41a68e9069b7919ef2ad82180170969af8b0f6b0feab151241bc96c7233f368cbaf3e8a3bf3315d815dd84d8bc518bc1e85b753a44d836ae335f7dcf4c11cae14b295cd8cd091b66209d6a3a6e256142aaaa38662bdd55fab100"

	if err := tw.ValidSignedTransactionRef(txSignedRaw); err != nil {
		t.Errorf("ValidSignedTransactionRef: %v\n", err)
	} else {
		t.Logf("GetTransactionSignRef return: \n\t%+v\n", "Success!")
	}
}

func TestSuiteTx(t *testing.T) {

	txRaw, err := tw.CreateTransactionRef(TOADDRESS, OWNERADDRESS, AMOUNT)
	if err != nil {
		t.Errorf("TestCreateTransaction failed: %v\n", err)
	}

	txSignedRaw, err := tw.GetTransactionSignRef(txRaw, PRIVATEKEY)
	if err != nil {
		t.Errorf("GetTransactionSignRef failed: %v\n", err)
	}

	err = tw.ValidSignedTransactionRef(txSignedRaw)
	if err != nil {
		t.Errorf("ValidSignedTransactionRef: %v\n", err)
	}
}
