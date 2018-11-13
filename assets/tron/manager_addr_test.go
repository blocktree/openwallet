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

func TestCreateAddress(t *testing.T) {

	var (
		passValue = "7465737470617373776f7264"

		predictBase58checkAddress = "TWwv3YcHJ1NfMemQSmXCPY48RR1tsY3n9N"
		// predictAddressHex         = "41e61c1205ee029fb4e41f294afd448cc5d578c8ef"
	)

	if r, err := tw.CreateAddress(passValue); err != nil {
		t.Errorf("CreateAddress failed: %v\n", err)
	} else {
		fmt.Printf("CreateAddress return: \n\t%+v\n", r)

		if r != predictBase58checkAddress {
			t.Errorf("CreateAddress failed: %v\n", "Data Invalid!")
		}
	}
}

func TestGenerateAddress(t *testing.T) {

	if r, err := tw.GenerateAddress(); err != nil {
		t.Errorf("GenerateAddress failed: %v\n", err)
	} else {
		fmt.Printf("GenerateAddress return: \n\t%+v\n", r)
	}
}

func TestValidateAddress(t *testing.T) {

	var addr = OWNERADDRESS

	if err := tw.ValidateAddress(addr); err != nil {
		t.Errorf("ValidateAddress failed: %v\n", err)
	} else {
		fmt.Printf("ValidateAddress return: \n\t%+v\n", "Success!")
	}
}
