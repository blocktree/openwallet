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

package merchant

import (
	"path/filepath"
	"testing"
)

var (
	nodeConfig NodeConfig
)

func init() {
	nodeConfig = NodeConfig{
		NodeKey:         "",
		PublicKey:       "dajosidjaiosjdioajsdioajsdiowhefi",
		PrivateKey:      "",
		MerchantNodeURL: "ws://192.168.30.4:8084/websocket",
		NodeID:          1,
		CacheFile:       filepath.Join(merchantDir, cacheFile),
	}

}

func TestGetChargeAddressVersion(t *testing.T) {
	m, err := NewMerchantNode(nodeConfig)
	if err != nil {
		t.Errorf("GetChargeAddressVersion failed unexpected error: %v", err)
	}

	params := struct {
		Coin     string `json:"coin"`
		WalletID string `json:"walletID"`
	}{"BTM", "123456"}

	//获取订阅的地址版本
	GetChargeAddressVersion(m.Node, params,
		true,
		func(addressVer *AddressVersion) {

			t.Logf("AddressVersion = %v", addressVer)

		})
}

func TestGetChargeAddress(t *testing.T) {
	m, err := NewMerchantNode(nodeConfig)
	if err != nil {
		t.Errorf("GetChargeAddressVersion failed unexpected error: %v", err)
	}

	params := struct {
		Coin     string `json:"coin"`
		WalletID string `json:"walletID"`
		Offset   uint64 `json:"offset"`
		Limit    uint64 `json:"limit"`
	}{"BTM", "0F8AV1FP00A02", 0, 20}

	GetChargeAddress(m.Node, params,
		true,
		func(addrs []*Address) {
			t.Logf("addrs.count = %v", len(addrs))
		})
}
