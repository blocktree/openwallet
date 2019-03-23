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

package eosio

import (
	"github.com/blocktree/openwallet/log"
	"testing"
)

func testNewWalletManager() *WalletManager {
	wm := NewWalletManager()
	wm.Config.ServerAPI = "http://127.0.01:8888/v1"
	wm.WalletClient = NewClient(wm.Config.ServerAPI, true)
	return wm
}

func TestWalletManager_GetInfo(t *testing.T) {
	wm := testNewWalletManager()
	r, err := wm.GetInfo()
	if err != nil {
		log.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("%+v", r)
}

func TestWalletManager_GetAccount(t *testing.T) {
	wm := testNewWalletManager()
	r, err := wm.GetAccount("bob")
	if err != nil {
		log.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("%+v", r)
}