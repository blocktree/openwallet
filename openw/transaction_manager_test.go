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
	"github.com/blocktree/OpenWallet/openwallet"
)

func createTransaction(tm *WalletManager, walletID, accountID, to string) (*openwallet.RawTransaction, error) {

	err := tm.RefreshAssetsAccountBalance(testApp, accountID)
	if err != nil {
		log.Error("RefreshAssetsAccountBalance failed, unexpected error:", err)
		return nil, err
	}

	rawTx, err := tm.CreateTransaction(testApp, walletID, accountID, "0.1", to, "0.1", "", nil)

	if err != nil {
		log.Error("CreateTransaction failed, unexpected error:", err)
		return nil, err
	}

	return rawTx, nil
}

func TestWalletManager_CreateTransaction(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "59t47qyjHUMZ6PGAdjkJopE9ffAPUkdUhSinJqcWRYZ1"
	to := "d35f9Ea14D063af9B3567064FAB567275b09f03D"

	rawTx, err := createTransaction(tm, walletID, accountID, to)

	if err != nil {
		return
	}

	log.Info("rawTx:", rawTx)

}

func TestWalletManager_CreateQrc20TokenTransaction(t *testing.T) {
	walletID := "WEP6cD2YSV773QZw5UuSS5U74XKdw6oQE2"
	accountID := "HCkvzSiWd4CLvRbkwUMzsjvydgRmGEbohrPPJTDy3PQb"
	to := "qYHPRYDUNq6ScqbweP5Cawnyp566VWBfUi"
	feeRate := "0.00000040"
	contractAddr := "91a6081095ef860d28874c9db613e7a4107b0281"
	tokenName := "QRC ZB TEST"
	tokeSymbol := "QZTC"
	var tokenDecimal uint64 = 8
	tm := testInitWalletManager()
	err := tm.RefreshAssetsAccountBalance(testApp, accountID)
	if err != nil {
		log.Error("RefreshAssetsAccountBalance failed, unexpected error:", err)
	}

	rawTx, err := tm.CreateQrc20TokenTransaction(testApp, walletID, accountID, "0.4", to, feeRate, "", contractAddr, tokenName, tokeSymbol, tokenDecimal)
	if err != nil {
		log.Error("CreateQrc20TokenTransaction failed, unexpected error:", err)
	} else {
		log.Info("rawTx:", rawTx)
	}
}

func TestWalletManager_SendQrc20TokenTransaction(t *testing.T) {
	walletID := "WBiuSsYdPgzZhLSMkAA7XYQuKUgsb6DQgJ"
	accountID := "HYZs6qySNUCj5w1Px7Btm1VF1TLkFvhGcL1FTN9gTUz"
	to := "qVT4jAoQDJ6E4FbjW1HPcwgXuF2ZdM2CAP"
	feeRate := "0.01"
	contractAddr := "0x91a6081095ef860d28874c9db613e7a4107b0281"
	tokenName := "QRC ZB TEST"
	tokeSymbol := "QZTC"
	var tokenDecimal uint64 = 8
	tm := testInitWalletManager()
	err := tm.RefreshAssetsAccountBalance(testApp, accountID)
	if err != nil {
		log.Error("RefreshAssetsAccountBalance failed, unexpected error:", err)
	}

	rawTx, err := tm.CreateQrc20TokenTransaction(testApp, walletID, accountID, "1", to, feeRate, "", contractAddr, tokenName, tokeSymbol, tokenDecimal)
	if err != nil {
		log.Error("CreateQrc20TokenTransaction failed, unexpected error:", err)
		return
	} else {
		log.Info("rawTx:", rawTx)
	}

	log.Info("rawTx unsigned:", rawTx.RawHex)

	_, err = tm.SignTransaction(testApp, walletID, accountID, "12345678", rawTx)
	if err != nil {
		log.Error("SignTransaction failed, unexpected error:", err)
		return
	}

	log.Info("rawTx.Signatures:", rawTx.Signatures)

	_, err = tm.VerifyTransaction(testApp, walletID, accountID, rawTx)
	if err != nil {
		log.Error("VerifyTransaction failed, unexpected error:", err)
		return
	}

	log.Info("rawTx signed:", rawTx.RawHex)

	tx, err := tm.SubmitTransaction(testApp, walletID, accountID, rawTx)
	if err != nil {
		log.Error("SubmitTransaction failed, unexpected error:", err)
		return
	}
	log.Info("wxID:", tx.WxID)
	log.Info("txID:", rawTx.TxID)
}

func TestWalletManager_SignTransaction(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "WEP6cD2YSV773QZw5UuSS5U74XKdw6oQE2"
	accountID := "HCkvzSiWd4CLvRbkwUMzsjvydgRmGEbohrPPJTDy3PQb"
	to := "qYHPRYDUNq6ScqbweP5Cawnyp566VWBfUi"

	rawTx, err := createTransaction(tm, walletID, accountID, to)
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
	tm := testInitWalletManager()
	walletID := "WEP6cD2YSV773QZw5UuSS5U74XKdw6oQE2"
	accountID := "HCkvzSiWd4CLvRbkwUMzsjvydgRmGEbohrPPJTDy3PQb"
	to := "qVT4jAoQDJ6E4FbjW1HPcwgXuF2ZdM2CAP"

	rawTx, err := createTransaction(tm, walletID, accountID, to)
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
	tm := testInitWalletManager()
	walletID := "WBiuSsYdPgzZhLSMkAA7XYQuKUgsb6DQgJ"
	accountID := "HYZs6qySNUCj5w1Px7Btm1VF1TLkFvhGcL1FTN9gTUz"
	to := "qVT4jAoQDJ6E4FbjW1HPcwgXuF2ZdM2CAP"

	rawTx, err := createTransaction(tm, walletID, accountID, to)
	if err != nil {
		return
	}

	log.Info("rawTx unsigned:", rawTx.RawHex)

	_, err = tm.SignTransaction(testApp, walletID, accountID, "12345678", rawTx)
	if err != nil {
		log.Error("SignTransaction failed, unexpected error:", err)
		return
	}

	log.Info("rawTx.Signatures:", rawTx.Signatures)

	_, err = tm.VerifyTransaction(testApp, walletID, accountID, rawTx)
	if err != nil {
		log.Error("VerifyTransaction failed, unexpected error:", err)
		return
	}

	log.Info("rawTx signed:", rawTx.RawHex)

	tx, err := tm.SubmitTransaction(testApp, walletID, accountID, rawTx)
	if err != nil {
		log.Error("SubmitTransaction failed, unexpected error:", err)
		return
	}
	log.Info("wxID:", tx.WxID)
	log.Info("txID:", rawTx.TxID)
}

func TestWalletManager_GetTransactions(t *testing.T) {
	tm := testInitWalletManager()
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
	tm := testInitWalletManager()
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
	tm := testInitWalletManager()
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
	tm := testInitWalletManager()
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

func TestWalletManager_GetTransactionByWxID(t *testing.T) {
	tm := testInitWalletManager()
	wxID := openwallet.GenTransactionWxID(&openwallet.Transaction{
		TxID: "bfa6febb33c8ddde9f7f7b4d93043956cce7e0f4e95da259a78dc9068d178fee",
		Coin: openwallet.Coin{
			Symbol:     "LTC",
			IsContract: false,
			ContractID: "",
		},
	})
	log.Info("wxID:", wxID)
	//"D0+rxcKSqEsFMfGesVzBdf6RloM="
	tx, err := tm.GetTransactionByWxID(testApp, wxID)
	if err != nil {
		log.Error("GetTransactionByTxID failed, unexpected error:", err)
		return
	}
	log.Info("tx:", tx)
}

func TestWalletManager_GetAssetsAccountBalance(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "2m27uVj2xx645dDCEcGD1whPQGcB4fZv16TzBoLGCyKB"

	balance, err := tm.GetAssetsAccountBalance(testApp, walletID, accountID)
	if err != nil {
		log.Error("GetAssetsAccountBalance failed, unexpected error:", err)
		return
	}
	log.Info("balance:", balance)
}

func TestWalletManager_GetAssetsAccountTokenBalance(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "59t47qyjHUMZ6PGAdjkJopE9ffAPUkdUhSinJqcWRYZ1"

	contract := openwallet.SmartContract{
		Address:  "0x4092678e4E78230F46A1534C0fbc8fA39780892B",
		Symbol:   "ETH",
		Name:     "OCoin",
		Token:    "OCN",
		Decimals: 18,
	}

	balance, err := tm.GetAssetsAccountTokenBalance(testApp, walletID, accountID, contract)
	if err != nil {
		log.Error("GetAssetsAccountTokenBalance failed, unexpected error:", err)
		return
	}
	log.Info("balance:", balance.Balance)
}

func TestWalletManager_GetAssetsAccountOnmiBalance(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "2m27uVj2xx645dDCEcGD1whPQGcB4fZv16TzBoLGCyKB"

	contract := openwallet.SmartContract{
		Address:  "2",
		Symbol:   "BTC",
		Name:     "TetherUSD",
		Token:    "USDT",
		Decimals: 8,
	}

	balance, err := tm.GetAssetsAccountTokenBalance(testApp, walletID, accountID, contract)
	if err != nil {
		log.Error("GetAssetsAccountTokenBalance failed, unexpected error:", err)
		return
	}
	log.Info("balance:", balance.Balance)
}

func TestWalletManager_GetEstimateFeeRate(t *testing.T) {
	tm := testInitWalletManager()
	coin := openwallet.Coin{
		Symbol: "ETH",
	}
	feeRate, unit, err := tm.GetEstimateFeeRate(coin)
	if err != nil {
		log.Error("GetEstimateFeeRate failed, unexpected error:", err)
		return
	}
	log.Std.Info("feeRate: %s %s/%s", feeRate, coin.Symbol, unit)
}
