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

func TestGetTotalTransaction(t *testing.T) {
	if r, err := tw.GetTotalTransaction(); err != nil {
		t.Errorf("TestGetTotalTransaction failed: %v\n", err)
	} else {
		t.Logf("TestGetTotalTransaction return: \n\t%+v\n", r)
	}
}

func TestGetTransactionByID(t *testing.T) {

	var txID string = "d5ec749ecc2a615399d8a6c864ea4c74ff9f523c2be0e341ac9be5d47d7c2d62"

	if r, err := tw.GetTransactionByID(txID); err != nil {
		t.Errorf("TestGetTransactionByID failed: %v\n", err)
	} else {
		t.Logf("TestGetTransactionByID return: \n\t%+v\n", r)
	}
}

func TestCreateTransaction(t *testing.T) {
	// RAW_expect := "0a7e0a0231d422084246e99b0394a3da40b0b4d2b0df2c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a1541b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7a121541e6992304ae03e5c6bba7334432b7345bef031c1418c0843d"

	if r, err := tw.CreateTransaction(TOADDRESS, OWNERADDRESS, AMOUNT); err != nil {
		t.Errorf("TestCreateTransaction failed: %v\n", err)
	} else {
		// if strings.Join(r[:], "") != RAW_expect {
		// 	t.Errorf("TestCreateTransaction return invalid RAW!")
		// }
		t.Logf("TestCreateTransaction return: \n\t%+v\n", r)
	}

}

func TestGetTransactoinSign(t *testing.T) {
	var transaction string = ""
	if r, err := tw.GetTransactionSign(transaction, PRIVATEKEY); err != nil {
		t.Errorf("TestCreateTransaction failed: %v\n", err)
	} else {
		t.Logf("TestCreateTransaction return: \n\t%+v\n", r)
	}

}

func TestBroadcastTransaction(t *testing.T) {
	var raw string
	raw = "0a7e0a0265912208acca59b1293334d640d88cee90e62c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d12414f7008a14665d7acbb65da8aca558a12f3c7bb668bca700536fa46adf49501d7dd312207bca0788940379f7eaa5ced8131b8945a3aa664fe0b8fa86e5ae0eab501"
	// raw = "0a85010a0222c32208d584f987611b830a40cb89a490e62c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d70abf0a190e62c1241f520824e57f002ba1b99931b0338a923c305bd6a68ae1ab2b6223a6ec33a27d00713be1525aeccad532c61a736089abc635244395b0e6e8a1983192588cd649101"
	if err := tw.BroadcastTransaction(raw); err != nil {
		t.Errorf("BroadcastTransaction failed: %v\n", err)
	} else {
		t.Logf("BroadcastTransaction return: \n\t%+v\n", "Success!")
	}
}
