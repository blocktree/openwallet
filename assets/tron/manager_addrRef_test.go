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

func TestCreateAddressRef(t *testing.T) {

	var privateKeyValue, predictedAddr string

	privateKeyValue = "9e9fa25c9d70fecc91c90d23b55daffa2f5f23ffa9eeca823260e50e544cf7be"
	predictedAddr = "TQ1TiUzStbSLdEtACUDmzfMDpWUyo8cyCf"

	if r, err := tw.CreateAddressRef(privateKeyValue); err != nil {
		t.Errorf("CreateAddressRef failed: %v\n", err)
	} else {
		if r != predictedAddr {
			t.Errorf("CreateAddressRef failed: not equal!\n")
		} else {
			fmt.Printf("CreateAddressRef return: \n\t%+v\n", r)
			fmt.Printf("Pred: %+v\n", predictedAddr)
		}
	}
}

func TestValidateAddressRef(t *testing.T) {

	var addr string
	addr = "TQ1TiUzStbSLdEtACUDmzfMDpWUyo8cyCf"
	addr = OWNERADDRESS

	if err := tw.ValidateAddressRef(addr); err != nil {
		t.Errorf("ValidateAddressRef failed: \n\t%+v\n", err)
	} else {
		fmt.Printf("CreateAddressRef return: \n\tSuccess!\n")
	}
}
