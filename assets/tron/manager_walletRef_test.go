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

func TestGetWallets(t *testing.T) {

	if r, err := tw.GetWallets(); err != nil {
		t.Errorf("GetWallets failed: %v\n", err)
	} else {
		tw.printWalletList(r)
		t.Logf("GetWallets return: \n%+v\n", r)
	}
}

func TestCreateNewWallet(t *testing.T) {

	var (
		wallet_name string = "simon"
		wallet_pass string = "1234qwer"
	)
	if r, s, err := tw.CreateNewWallet(wallet_name, wallet_pass); err != nil {
		t.Errorf("CreateNewWallet failed: %v\n", err)
	} else {
		fmt.Printf("keyfile = %+v\n", s)
		t.Logf("TestCreateTrCreateNewWalletansaction return: \n%+v\n", r)
	}
}
