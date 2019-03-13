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

package merchant

import (
	"testing"
	"path/filepath"
	"log"
	"github.com/blocktree/openwallet/openwallet"
)


func init() {
	nodeConfig = NodeConfig{
		MerchantNodeURL: "ws://192.168.30.4:8084/websocket",
		CacheFile:       filepath.Join(merchantDir, cacheFile),
	}

}

func TestGetImportAddress(t *testing.T) {

	m, err := NewMerchantNode(nodeConfig)
	if err != nil {
		t.Errorf("GetChargeAddressVersion failed unexpected error: %v", err)
	}

	walletID := "adsdsd1231"
	wallet, err := m.GetMerchantWalletByID(walletID)
	if err != nil {
		log.Printf("unexpected error: %v", err)
		return
	}

	db, err := wallet.OpenDB()
	if err != nil {
		log.Printf("unexpected error: %v", err)
		return
	}
	defer db.Close()

	var addresses []*openwallet.Address
	err = db.All(&addresses)
	if err != nil {
		log.Printf("unexpected error: %v", err)
		return
	}

	for _, a := range addresses {
		log.Printf("address = %s\n", a.Address)
	}

}

func TestGetAllMerchantWallet(t *testing.T) {

	m, err := NewMerchantNode(nodeConfig)
	if err != nil {
		t.Errorf("GetChargeAddressVersion failed unexpected error: %v", err)
	}

	db, err := m.OpenDB()
	if err != nil {
		log.Printf("unexpected error: %v", err)
		return
	}
	defer db.Close()

	var wallets []*openwallet.Wallet
	db.All(&wallets)
	for _, w := range wallets {
		log.Printf("wallet = %v\n", *w)
	}

	var wallet openwallet.Wallet
	err = db.One("WalletID", "adsdsd1231", &wallet)
	if err != nil {
		log.Printf("unexpected error: %v", err)
		return
	}
}

func TestGetAllAddressVersion(t *testing.T) {

	m, err := NewMerchantNode(nodeConfig)
	if err != nil {
		t.Errorf("GetChargeAddressVersion failed unexpected error: %v", err)
	}

	db, err := m.OpenDB()
	if err != nil {
		log.Printf("unexpected error: %v", err)
		return
	}
	defer db.Close()

	db.Save(&AddressVersion{
		Key:"BTC_123123",
		Coin:"BTC",
		WalletID:"123123",
		Total:111,
		Version:2,
	})

	var addressVers []*AddressVersion
	db.All(&addressVers)
	for _, v := range addressVers {
		log.Printf("addressVersion = %v\n", *v)
	}

}