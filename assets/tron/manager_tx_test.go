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

	var txID string = "67b965cf2f47de253eb8cc2acf0cd64b19990836b361195c9fdcd3df489f8756"

	if r, isSuccess, err := tw.GetTransactionByID(txID); err != nil || isSuccess != true {
		t.Errorf("TestGetTransactionByID failed: %v\n", err)
	} else {
		t.Logf("TestGetTransactionByID return: \n\t%+v\n", r)
	}
}

func TestCreateTransaction(t *testing.T) {

	if r, err := tw.CreateTransaction(TOADDRESS, OWNERADDRESS, AMOUNT); err != nil {
		t.Errorf("TestCreateTransaction failed: %v\n", err)
	} else {
		t.Logf("TestCreateTransaction return: \n\t%+v\n", r)
	}

}

func TestGetTransactoinSign(t *testing.T) {

	var txRaw string = ""

	if r, err := tw.GetTransactionSign(txRaw, PRIVATEKEY); err != nil {
		t.Errorf("TestCreateTransaction failed: %v\n", err)
	} else {
		t.Logf("TestCreateTransaction return: \n\t%+v\n", r)
	}

}

func TestBroadcastTransaction(t *testing.T) {

	var raw string = "0a7e0a02fd77220882256bb5fe08d39240d0a7c98fe82c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d1241373bf54b04e287d902beff4c6bd7369395b7b65527513922ee3b61ac0c4c6e8d0061da08b1b2f361e53c933360c3e5783996339431d44469f8bd57ee8fdfd3d700"

	if err := tw.BroadcastTransaction(raw); err != nil {
		t.Errorf("BroadcastTransaction failed: %v\n", err)
	} else {
		t.Logf("BroadcastTransaction return: \n\t%+v\n", "Success!")
	}
}
