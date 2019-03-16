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

package obyte

import (
	"github.com/blocktree/openwallet/log"
	"testing"
)


func testNewWalletManager() *WalletManager {

	tw := NewWalletManager()

	tw.Config.ServerAPI = "http://47.106.102.2:10015"
	tw.WalletClient = NewClient(tw.Config.ServerAPI, "", true)

	return tw
}

func TestWalletManager_GetInfo(t *testing.T) {
	tw := testNewWalletManager()
	b, err := tw.GetInfo()
	if err != nil {
		t.Errorf("GetInfo failed unexpected error: %v\n", err)
	} else {
		log.Infof("GetInfo info: %v", b)
	}
}

func TestWalletManager_GetNewAddress(t *testing.T) {
	tw := testNewWalletManager()
	b, err := tw.GetNewAddress()
	if err != nil {
		t.Errorf("GetNewAddress failed unexpected error: %v\n", err)
	} else {
		log.Infof("GetNewAddress: %+v", b)
	}
}

func TestWalletManager_GetChangeAddress(t *testing.T) {
	tw := testNewWalletManager()
	b, err := tw.GetChangeAddress()
	if err != nil {
		t.Errorf("GetChangeAddress failed unexpected error: %v\n", err)
	} else {
		log.Infof("GetChangeAddress: %+v", b)
	}
}

func TestWalletManager_GetBalance(t *testing.T) {
	tw := testNewWalletManager()
	b, err := tw.GetBalance()
	if err != nil {
		t.Errorf("GetBalance failed unexpected error: %v\n", err)
	} else {
		log.Infof("GetBalance: %+v", b)
	}
}


func TestWalletManager_SendToAddress(t *testing.T) {
	tw := testNewWalletManager()
	b, err := tw.SendToAddress("I4CEUEFL7BOWXJUHE7XYKCDP4AA2QG3Z", 100)
	if err != nil {
		t.Errorf("SendToAddress failed unexpected error: %v\n", err)
	} else {
		log.Infof("SendToAddress: %+v", b)
	}
}