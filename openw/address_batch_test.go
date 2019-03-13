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

package openw

import (
	"encoding/base64"
	"encoding/json"
	"github.com/blocktree/OpenWallet/assets/litecoin"
	"github.com/blocktree/OpenWallet/crypto"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"testing"
)

func TestBatchCreateAddressByAccount(t *testing.T) {
	account := openwallet.AssetsAccount{
		AccountID:      "Aa7Chh2MdaGDejHdCJZAaX7AwvGNmMEMry2kZZTq114a",
		Symbol:         "LTC",
		PublicKey:      "owpubeyoV6FsNaAN9wxUDsBAXyKr4EVt3ib6R1CRLpGJ7qUdGUgiwB6tLbUATvFYDm2JGTaB6i7PkeXBAxzo5RkNXmx64KE9vFN4qPtbK7ywZTd9unM2C1",
		OwnerKeys: []string{"owpubeyoV6FsNaAN9wxUDsBAXyKr4EVt3ib6R1CRLpGJ7qUdGUgiwB6tLbUATvFYDm2JGTaB6i7PkeXBAxzo5RkNXmx64KE9vFN4qPtbK7ywZTd9unM2C1"},
		HDPath:         "m/44'/88'/1'",
		AddressIndex:   0,
		Index:   1,
	}
	wm := litecoin.NewWalletManager()
	decoder := wm.Decoder

	addrArr, err := openwallet.BatchCreateAddressByAccount(&account, decoder, 5000, 20)
	if err != nil {
		t.Errorf("error: %v", err)
		return
	}
	addrs := make([]string, 0)
	for _, a := range addrArr {
		//t.Logf("address[%d]: %s", a.Index, a.Address)
		addrs = append(addrs, a.Address)
	}
	log.Infof("create address")
	j, err := json.Marshal(addrs)
	if err != nil {
		t.Errorf("json.Marshal: %v", err)
		return
	}
	log.Infof("json.Marshal")
	enc, err := crypto.AESEncrypt(j, crypto.SHA256([]byte("1234")))
	if err != nil {
		t.Errorf("crypto.AESEncrypt: %v", err)
		return
	}
	log.Infof("AESEncrypt")
	encB58 := base64.StdEncoding.EncodeToString(enc)
	log.Infof("base58.Encode: %s", encB58)
	decB58, _ := base64.StdEncoding.DecodeString(encB58)
	log.Infof("base58.Decode")
	dec, err := crypto.AESDecrypt(decB58, crypto.SHA256([]byte("1234")))
	if err != nil {
		t.Errorf("crypto.AESEncrypt: %v", err)
		return
	}
	log.Infof("AESDecrypt")
	log.Infof("dec: %s", string(dec))
}
