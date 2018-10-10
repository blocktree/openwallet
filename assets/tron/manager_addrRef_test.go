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

func TestCreateBatchAddress(t *testing.T) {

	var (
		walletID string = "W4Hv5qiUb3R7GVQ9wgmX8MfhZ1GVR6dqL7"
		password string = "1234qwer"
		count    uint64 = 10000
	)
	if s, r, err := tw.CreateBatchAddress(walletID, password, count); err != nil {
		t.Errorf("CreateBatchAddress failed: \n\t%+v\n", err)
	} else {
		_, _ = s, r
		// fmt.Printf("CreateBatchAddress return: \n\t%+v\n", r)
		tw.printAddressList(r)
	}
}

func TestGetAddressesFromLocalDB(t *testing.T) {

	var (
		walletID string = "W4Hv5qiUb3R7GVQ9wgmX8MfhZ1GVR6dqL7"
		offset   int    = 0
		limit    int    = -1
	)
	if r, err := tw.GetAddressesFromLocalDB(walletID, offset, limit); err != nil {
		t.Errorf("GetAddressesFromLocalDB failed: \n\t%+v\n", err)
	} else {
		fmt.Printf("GetAddressesFromLocalDB return: \n\t%+v\n", r)

		// for i, a := range r {
		// 	t.Logf("GetAddressesFromLocalDB address[%d] = %v\n", i, a)
		// }

		tw.printAddressList(r)
	}
}
