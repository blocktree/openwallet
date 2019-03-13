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
		walletName = "simon"
		walletPass = "1234qwer"
	)

	if r, s, err := tw.CreateNewWallet(walletName, walletPass); err != nil {
		t.Errorf("CreateNewWallet failed: %v\n", err)
	} else {
		t.Logf("TestCreateTrCreateNewWalletansaction return: \n%+v\n", r)

		fmt.Printf("keyfile = %+v\n", s)
	}
}
