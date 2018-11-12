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
	"encoding/hex"
	"fmt"
	"testing"
)

func TestGetPrivateKeyRef(t *testing.T) {
	var (
		walletID          = "W4Hv5qiUb3R7GVQ9wgmX8MfhZ1GVR6dqL7"
		password          = "1234qwer"
		index      uint64 = 1539142990
		serializes uint32
	)

	if r, err := tw.GetPrivateKeyRef(walletID, password, index, serializes); err != nil {
		t.Errorf("CreateAddressRef failed: %v\n", err)
	} else {

		h, _ := hex.DecodeString(r)

		if addr, err := tw.CreateAddressRef(h, true); err != nil {
			t.Errorf("CreateAddressRef failed: %v\n", err)
		} else {
			fmt.Println("Address = ", addr)
		}

		fmt.Printf("GetPrivateKeyRef return: \n\t%+v\n\n", r)
	}
}
