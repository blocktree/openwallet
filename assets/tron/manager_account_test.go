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

import "testing"

func TestGetAccountNet(t *testing.T) {

	var addr string = "4189139CB1387AF85E3D24E212A008AC974967E561"
	addr = "TUZYMxXwwnX77rQvmP3RMwm3RhfBJ7bovV"

	if r, err := tw.GetAccountNet(addr); err != nil {
		t.Errorf("GetAccountNet failed: %v\n", err)
	} else {
		t.Logf("GetAccountNet return: \n\t%+v\n", r)
	}
}

func TestCreateAccount(t *testing.T) {

	var owner_address, account_address string = "TUZYMxXwwnX77rQvmP3RMwm3RhfBJ7bovV", "TUZYMxXwwnX77rQvmP3RMwm3RhfBJ7bovV"

	if r, err := tw.CreateAccount(owner_address, account_address); err != nil {
		t.Errorf("CreateAccount failed: %v\n", err)
	} else {
		t.Logf("CreateAccount return: \n\t%+v\n", r)
	}
}

func TestUpdateAccount(t *testing.T) {

	var account_name, owner_address string = "", "4189139CB1387AF85E3D24E212A008AC974967E561"
	// addr = "TUZYMxXwwnX77rQvmP3RMwm3RhfBJ7bovV"

	if r, err := tw.UpdateAccount(account_name, owner_address); err != nil {
		t.Errorf("UpdateAccount failed: %v\n", err)
	} else {
		t.Logf("UpdateAccount return: \n\t%+v\n", r)
	}
}
