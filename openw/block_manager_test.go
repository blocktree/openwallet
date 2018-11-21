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
	"testing"
	"github.com/blocktree/OpenWallet/log"
)

func init() {
	//tm.Init()
}

func TestWalletManager_RescanBlockHeight(t *testing.T) {
	tm := testInitWalletManager()
	sub := subscriber{}
	tm.AddObserver(&sub)
	tm.RescanBlockHeight("QTUM", 235000, 235002)
}


func TestWalletManager_GetNewBlockHeight(t *testing.T) {
	tm := testInitWalletManager()
	currentHeight, scannedHeight, err := tm.GetNewBlockHeight("BTC")
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}

	log.Debug("currentHeight:", currentHeight)
	log.Debug("scannedHeight:", scannedHeight)
}