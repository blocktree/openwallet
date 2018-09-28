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

	rawTx, err := tm.CreateTransaction(testApp, walletID, accountID, "0.01", to, "", "")
	if err != nil {
		log.Error("CreateTransaction failed, unexpected error:", err)
		return nil, err
	}

	return rawTx, nil
}

func TestWalletManager_CreateTransaction(t *testing.T) {

	walletID := "WJwzaG2G4LoyuEb7NWAYiDa6DbtARtbUGv"
	accountID := "JYCcXtC18vnd1jbcJX47msDFbQMBDNjsq3xbvvK6qCHKAAqoQq"
	to := "QjPGf82YKxygtr7H2Tm9A8P7ETBsWG2ih9"

	rawTx, err := createTransaction(walletID, accountID, to)

	if err != nil {
		return
	}

	log.Info("rawTx:", rawTx)

}

func TestWalletManager_SignTransaction(t *testing.T) {

	walletID := "WJwzaG2G4LoyuEb7NWAYiDa6DbtARtbUGv"
	accountID := "JYCcXtC18vnd1jbcJX47msDFbQMBDNjsq3xbvvK6qCHKAAqoQq"
	to := "QjPGf82YKxygtr7H2Tm9A8P7ETBsWG2ih9"

	rawTx, err := createTransaction(walletID, accountID, to)
	if err != nil {
		return
	}

	_, err = tm.SignTransaction(testApp, walletID, accountID, "12345678", rawTx)
	if err != nil {
		log.Error("SignTransaction failed, unexpected error:", err)
		return
	}

	log.Info("rawTx:", rawTx)

}

func TestWalletManager_VerifyTransaction(t *testing.T) {

	walletID := "WJwzaG2G4LoyuEb7NWAYiDa6DbtARtbUGv"
	accountID := "JYCcXtC18vnd1jbcJX47msDFbQMBDNjsq3xbvvK6qCHKAAqoQq"
	to := "QjPGf82YKxygtr7H2Tm9A8P7ETBsWG2ih9"

	rawTx, err := createTransaction(walletID, accountID, to)
	if err != nil {
		return
	}

	_, err = tm.SignTransaction(testApp, walletID, accountID, "12345678", rawTx)
	if err != nil {
		log.Error("SignTransaction failed, unexpected error:", err)
		return
	}

	//log.Info("rawTx.Signatures:", rawTx.Signatures)

	_, err = tm.VerifyTransaction(testApp, walletID, accountID, rawTx)
	if err != nil {
		log.Error("VerifyTransaction failed, unexpected error:", err)
		return
	}

	log.Info("rawTx:", rawTx)

}

func TestWalletManager_SubmitTransaction(t *testing.T) {

	walletID := "WJwzaG2G4LoyuEb7NWAYiDa6DbtARtbUGv"
	accountID := "JYCcXtC18vnd1jbcJX47msDFbQMBDNjsq3xbvvK6qCHKAAqoQq"
	to := "QjcVLQmNBX45icPZ1jQMRSfb4PbLNLws86"

	rawTx, err := createTransaction(walletID, accountID, to)
	if err != nil {
		return
	}

	log.Info("rawTx unsigned:", rawTx.RawHex)

	_, err = tm.SignTransaction(testApp, walletID, accountID, "12345678", rawTx)
	if err != nil {
		log.Error("SignTransaction failed, unexpected error:", err)
		return
	}

	//log.Info("rawTx.Signatures:", rawTx.Signatures)

	_, err = tm.VerifyTransaction(testApp, walletID, accountID, rawTx)
	if err != nil {
		log.Error("VerifyTransaction failed, unexpected error:", err)
		return
	}

	log.Info("rawTx signed:", rawTx.RawHex)

	_, err = tm.SubmitTransaction(testApp, walletID, accountID, rawTx)
	if err != nil {
		log.Error("SubmitTransaction failed, unexpected error:", err)
		return
	}

	log.Info("txID:", rawTx.TxID)
}

func TestWalletManager_GetTransactions(t *testing.T) {
	list, err := tm.GetTransactions(testApp, 0, -1, "Received", false)
	if err != nil {
		log.Error("GetTransactions failed, unexpected error:", err)
		return
	}
	for i, tx := range list {
		log.Info("trx[", i, "] :", tx)
	}
	log.Info("trx count:", len(list))
}

func TestWalletManager_GetTxUnspent(t *testing.T) {
	list, err := tm.GetTxUnspent(testApp, 0, -1, "Received", false)
	if err != nil {
		log.Error("GetTxUnspent failed, unexpected error:", err)
		return
	}
	for i, tx := range list {
		log.Info("Unspent[", i, "] :", tx)
	}
	log.Info("Unspent count:", len(list))
}

func TestWalletManager_GetTxSpent(t *testing.T) {
	list, err := tm.GetTxSpent(testApp, 0, -1, "Received", false)
	if err != nil {
		log.Error("GetTxSpent failed, unexpected error:", err)
		return
	}
	for i, tx := range list {
		log.Info("Spent[", i, "] :", tx)
	}
	log.Info("Spent count:", len(list))
}

func TestWalletManager_ExtractUTXO(t *testing.T) {

	unspent, err := tm.GetTxUnspent(testApp, 0, -1, "Received", false)
	if err != nil {
		log.Error("GetTxUnspent failed, unexpected error:", err)
		return
	}
	for i, tx := range unspent {

		_, err := tm.GetTxSpent(testApp, 0, -1, "SourceTxID", tx.TxID, "SourceIndex", tx.Index)
		if err == nil {
			continue
		}

		log.Info("ExtractUTXO[", i, "] :", tx)
	}

}