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
	"github.com/blocktree/OpenWallet/openwallet"
)

func createTransaction(walletID, accountID, to string) (*openwallet.RawTransaction, error) {

	err := tm.RefreshAssetsAccountBalance(testApp, accountID)
	if err != nil {
		log.Error("RefreshAssetsAccountBalance failed, unexpected error:", err)
		return nil, err
	}

	rawTx, err := tm.CreateTransaction(testApp, walletID, accountID, "0.1", to, "", "")
	if err != nil {
		log.Error("CreateTransaction failed, unexpected error:", err)
		return nil, err
	}

	return rawTx, nil
}

func TestWalletManager_CreateTransaction(t *testing.T) {

	walletID := "WFPHAs2uyeHcfBzKF4vN4NkMpArX8wkCxp"
	accountID := "Jq9LCP9AkDfi6zqsgwodfFHPQpo9VDxXCBaM6pxRQQYk1Ra1mH"
	to := "mySLanFVyyUL6P2fAsbiwfTuBBh9xf3vhT"

	rawTx, err := createTransaction(walletID, accountID, to)

	if err != nil {
		return
	}

	log.Info("rawTx:", rawTx)

}

func TestWalletManager_SignTransaction(t *testing.T) {

	walletID := "WFPHAs2uyeHcfBzKF4vN4NkMpArX8wkCxp"
	accountID := "Jq9LCP9AkDfi6zqsgwodfFHPQpo9VDxXCBaM6pxRQQYk1Ra1mH"
	to := "mySLanFVyyUL6P2fAsbiwfTuBBh9xf3vhT"

	rawTx, err := createTransaction(walletID, accountID, to)
	if err != nil {
		return
	}

	_, err = tm.SignTransaction(testApp, walletID, accountID, "12345678", rawTx)
	if err != nil {
		log.Error("CreateTransaction failed, unexpected error:", err)
		return
	}

	log.Info("rawTx:", rawTx)

}
