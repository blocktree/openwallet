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

func TestCreateAddress(t *testing.T) {

	var passValue string = "7465737470617373776f7264"

	if r, err := tw.CreateAddress(passValue); err != nil {
		t.Errorf("CreateAddress failed: %v\n", err)
	} else {
		t.Logf("CreateAddress return: \n\t%+v\n", r)
	}
}

func TestGenerateAddress(t *testing.T) {

	if r, err := tw.GenerateAddress(); err != nil {
		t.Errorf("GenerateAddress failed: %v\n", err)
	} else {
		t.Logf("GenerateAddress return: \n\t%+v\n", r)
	}
}

func TestValidateAddress(t *testing.T) {

	var addr string = "4189139CB1387AF85E3D24E212A008AC974967E561"

	if err := tw.ValidateAddress(addr); err != nil {
		t.Errorf("ValidateAddress failed: %v\n", err)
	} else {
		t.Logf("ValidateAddress return: \n\t%+v\n", "success!")
	}
}
