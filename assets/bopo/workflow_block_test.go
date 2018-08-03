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

package bopo

import (
	"fmt"
	"testing"
)

var (
	testUser    = "SimonLuo"
	testPass    = "1234qwer"
	testDataDir = "/openwallet/data/bch/testnet3/"
	testKey     = "" // Generating within testing CreateNewWallet
)

func init() {
	serverAPI = "http://192.168.2.194:10061"
	isTestNet = false
	client = &Client{
		BaseURL: serverAPI,
		Debug:   true,
	}
}

func TestGetBlockChainInfo(t *testing.T) {
	b, err := GetBlockChainInfo()
	if err != nil {
		t.Errorf("GetBlockChainInfo failed unexpected error: %v\n", err)
	} else {
		t.Logf("GetBlockChainInfo info: %v\n", b)
	}

	fmt.Printf("TestGetBlockChainInfo: \n\t%+v\n", b)
}
