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

func TestCreateTransactionRef(t *testing.T) {

	// predictTxRaw := "0a7e0a021031220816b0c1a29ce3387c40e099ad83e02c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a1541b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7a12154199fee02e1ee01189bc41a68e9069b7919ef2ad8218c0843d"
	to_address, owner_address, amount := "TQ1TiUzStbSLdEtACUDmzfMDpWUyo8cyCf", "TSdXzXKSQ3RQzQ5Ge8TiYfMQEjofSVQ8ax", uint64(1)

	if r, err := tw.CreateTransactionRef(to_address, owner_address, amount); err != nil {
		t.Errorf("TestCreateTransaction failed: %v\n", err)
	} else {
		// if strings.Join(r[:], "") != RAW_expect {
		// 	t.Errorf("TestCreateTransaction return invalid RAW!")
		// }
		t.Logf("TestCreateTransaction return: \n\t%+v\n", r)
		// fmt.Println("Predict Tx Raw: ", predictTxRaw)
		// fmt.Println("Returns Tx Raw: ", r)
	}

}

func TestGetTransactoinSignRef(t *testing.T) {

}

func TestBroadcastTransactionRef(t *testing.T) {
}
