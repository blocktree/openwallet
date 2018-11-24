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

package bitcoin

import (
	"testing"
)

func TestWalletManager_GetOmniBalance(t *testing.T) {
	propertyID := uint64(2)
	address := "n1ZurJRnQyoRwBrx6B7DMndjBWAxnRbxKJ"
	balance, err := tw.GetOmniBalance(propertyID, address)
	if err != nil {
		t.Errorf("GetOmniBalance failed unexpected error: %v\n", err)
		return
	}
	t.Logf("balance: %v\n", balance)
}

func TestWalletManager_GetOmniTransaction(t *testing.T) {
	txid := "c0aad040a04bba0168a63da3d41509722c49005d2fd045d9d0f81ad551c56f1d"
	transaction, err := tw.GetOmniTransaction(txid)
	if err != nil {
		t.Errorf("GetOmniBalance failed unexpected error: %v\n", err)
		return
	}
	t.Logf("transaction: %+v", transaction)
}

func TestWalletManager_GetOmniInfo(t *testing.T) {
	result, err := tw.GetOmniInfo()
	if err != nil {
		t.Errorf("TestWalletManager_GetOmniInfo failed unexpected error: %v\n", err)
		return
	}
	t.Logf("OmniInfo: %+v", result)
}

func TestWalletManager_GetOmniProperty(t *testing.T) {
	propertyID := uint64(31)
	result, err := tw.GetOmniProperty(propertyID)
	if err != nil {
		t.Errorf("GetOmniProperty failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetOmniProperty: %+v", result)
}