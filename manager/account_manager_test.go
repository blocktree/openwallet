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

package manager

import (
	"testing"
	"github.com/blocktree/OpenWallet/log"
)

func TestWalletManager_RefreshAssetsAccountBalance(t *testing.T) {

	walletID := "WJwzaG2G4LoyuEb7NWAYiDa6DbtARtbUGv"
	accountID := "JYCcXtC18vnd1jbcJX47msDFbQMBDNjsq3xbvvK6qCHKAAqoQq"
	err := tm.RefreshAssetsAccountBalance(testApp, accountID)
	if err != nil {
		log.Error("RefreshAssetsAccountBalance failed, unexpected error:", err)
		return
	}

	account, err := tm.GetAssetsAccountInfo(testApp, walletID, accountID)
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}

	log.Info("account:", account)
}

func TestWalletManager_ImportWatchOnlyAddress(t *testing.T) {

	walletID := "WJwzaG2G4LoyuEb7NWAYiDa6DbtARtbUGv"
	accountID := "JYCcXtC18vnd1jbcJX47msDFbQMBDNjsq3xbvvK6qCHKAAqoQq"

	address, err := tm.GetAddressList(testApp, walletID, accountID, 0, -1, true)
	if err != nil {
		log.Error("GetAddressList failed, unexpected error:", err)
		return
	}

	err = tm.ImportWatchOnlyAddress(testApp, walletID, accountID, address)
	if err != nil {
		log.Error("RefreshAssetsAccountBalance failed, unexpected error:", err)
		return
	}

}