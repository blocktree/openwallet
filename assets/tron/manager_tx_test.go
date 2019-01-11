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
	"fmt"
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

	var txID = "2924a364b7bfc2c1f62796377d83008eb9c310a0f07305030d0b7ac9127d1848"

	if r, err := tw.GetTransactionByID(txID); err != nil || r.IsSuccess != true {
		t.Logf("TestGetTransactionByID return: \n\t%+v\n", r)
		t.Errorf("TestGetTransactionByID failed: %v\n", err)
	} else {
		t.Logf("TestGetTransactionByID return: \n\t%+v\n", r)
	}
}

func TestCreateTransaction(t *testing.T) {

	if r, err := tw.CreateTransaction(TOADDRESS, OWNERADDRESS, AMOUNT); err != nil {
		t.Errorf("TestCreateTransaction failed: %v\n", err)
	} else {
		//t.Logf("TestCreateTransaction return: \n\t%+v\n", r)
		fmt.Println("Tx:=", r)
	}

}

func TestGetTransactoinSign(t *testing.T) {

	var txRaw = "0a7e0a023c462208c84cf406d3b89d2640ffbd85e0fa2c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a1541887661d2e0215851756b1e7933216064526badcd121541b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7a18a08d06"

	if r, err := tw.GetTransactionSign(txRaw, PRIVATEKEY); err != nil {
		t.Errorf("TestCreateTransaction failed: %v\n", err)
	} else {
		//t.Logf("TestCreateTransaction return: \n\t%+v\n", r)
		fmt.Println("signature:=", r)
	}

}

func TestBroadcastTransaction1(t *testing.T) {

	var raw = "0a7e0a02a1192208d3216b8e04dc955240e0c2c786fb2c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a1541887661d2e0215851756b1e7933216064526badcd121541b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7a18a08d061241fb21995f0edd84ecc68e05026acc85c695dce9ff50ec5a4864f6178caa9465aa16784e787b01d56954d628794138bb0c20416ec3fa96ba1d132329d678ef1a9c00"

	if err := tw.BroadcastTransaction1(raw); err != nil {
		t.Errorf("BroadcastTransaction failed: %v\n", err)
	} else {
		t.Logf("BroadcastTransaction return: \n\t%+v\n", "Success!")
	}
}
