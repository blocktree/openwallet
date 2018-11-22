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

func testCreateTransactionStep(tm *WalletManager, walletID, accountID, to, amount, feeRate string, contract *openwallet.SmartContract) (*openwallet.RawTransaction, error) {

	//err := tm.RefreshAssetsAccountBalance(testApp, accountID)
	//if err != nil {
	//	log.Error("RefreshAssetsAccountBalance failed, unexpected error:", err)
	//	return nil, err
	//}

	rawTx, err := tm.CreateTransaction(testApp, walletID, accountID, amount, to, feeRate, "", contract)

	if err != nil {
		log.Error("CreateTransaction failed, unexpected error:", err)
		return nil, err
	}

	return rawTx, nil
}

func testSignTransactionStep(tm *WalletManager, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	_, err := tm.SignTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, "12345678", rawTx)
	if err != nil {
		log.Error("SignTransaction failed, unexpected error:", err)
		return nil, err
	}

	log.Info("rawTx:", rawTx)
	return rawTx, nil
}

func testVerifyTransactionStep(tm *WalletManager, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	//log.Info("rawTx.Signatures:", rawTx.Signatures)

	_, err := tm.VerifyTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx)
	if err != nil {
		log.Error("VerifyTransaction failed, unexpected error:", err)
		return nil, err
	}

	log.Info("rawTx:", rawTx)
	return rawTx, nil
}

func testSubmitTransactionStep(tm *WalletManager, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	tx, err := tm.SubmitTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx)
	if err != nil {
		log.Error("SubmitTransaction failed, unexpected error:", err)
		return nil, err
	}
	log.Info("wxID:", tx.WxID)
	log.Info("txID:", rawTx.TxID)

	return rawTx, nil
}

func TestTransfer_ETH(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "59t47qyjHUMZ6PGAdjkJopE9ffAPUkdUhSinJqcWRYZ1"
	to := "0xd35f9Ea14D063af9B3567064FAB567275b09f03D"

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "0.0003", "", nil)
	if err != nil {
		return
	}

	log.Std.Info("rawTx: %+v", rawTx)

	_, err = testSignTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testVerifyTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testSubmitTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

}

func TestTransfer_ERC20(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "59t47qyjHUMZ6PGAdjkJopE9ffAPUkdUhSinJqcWRYZ1"
	to := "0xd35f9Ea14D063af9B3567064FAB567275b09f03D"

	contract := openwallet.SmartContract{
		Address:  "4092678e4E78230F46A1534C0fbc8fA39780892B",
		Symbol:   "ETH",
		Name:     "OCoin",
		Token:    "OCN",
		Decimals: 18,
	}

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "1", "", &contract)
	if err != nil {
		return
	}

	_, err = testSignTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testVerifyTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testSubmitTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

}

func TestTransfer_LTC(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "EbUsW3YaHQ61eNt3f4hDXJAFh9LGmLZWH1VTTSnQmhnL"
	to := "my2S5LBREZ8YCcuAHZz1YChoZpGPZN28uw"

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "0.01", "0.001", nil)
	if err != nil {
		return
	}

	_, err = testSignTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testVerifyTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testSubmitTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

}

func TestTransfer_QTUM(t *testing.T) {
	tm := testInitWalletManager()
	//Qf6t5Ww14ZWVbG3kpXKoTt4gXeKNVxM9QJ
	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "2by6wzbzw7cnWkxiA31xMHpFmE99bqL3BnjkUJnJtEN6"
	to := "QTdtEdduBTybwnRpDWc2A44oUiLTpp227k"

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "0.4", "", nil)
	if err != nil {
		return
	}

	_, err = testSignTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testVerifyTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testSubmitTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

}

func TestTransfer_QRC20(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "2by6wzbzw7cnWkxiA31xMHpFmE99bqL3BnjkUJnJtEN6"
	to := "Qf6t5Ww14ZWVbG3kpXKoTt4gXeKNVxM9QJ"

	contract := openwallet.SmartContract{
		Address:  "f2033ede578e17fa6231047265010445bca8cf1c",
		Symbol:   "QTUM",
		Name:     "QCASH",
		Token:    "QC",
		Decimals: 8,
	}

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "0.01", "", &contract)
	if err != nil {
		return
	}

	_, err = testSignTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testVerifyTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testSubmitTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

}

func TestTransfer_NAS(t *testing.T) {
	tm := testInitWalletManager()
	//walletID := "VzQTLspxvbXSmfRGcN6LJVB8otYhJwAGWc"
	//accountID := "BjLtC1YN4sWQKzYHtNPdvx3D8yVfXmbyeCQTMHv4JUGG"
	//to := "n1Prn7ZbZtd5CTN8Yrj4K9c3gD4u8tjFQzX"

	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "7VftKuNoDtwZ3mn3wDA4smTDMz4iqCg3fNna1fXicVDg"
	to := "n1LyK9R1jZuMWut28kUiQn8dEoMQUezt9GC"

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "0.005", "", nil)
	if err != nil {
		return
	}

	_, err = testSignTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testVerifyTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testSubmitTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

}

func TestTransfer_BTC(t *testing.T) {
	tm := testInitWalletManager()
	//ms9NeTGFtaMcjrqRyRogkHqRoR8b1sQwu3
	//mp1JDsi7Dr2PkcWu1j4SUSTXJqXjFMaeVx
	//n1ZurJRnQyoRwBrx6B7DMndjBWAxnRbxKJ
	//mxoCkSBmiLQ86N73kXNLHEUgcUBoKdFawH
	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "2m27uVj2xx645dDCEcGD1whPQGcB4fZv16TzBoLGCyKB"
	to := "ms9NeTGFtaMcjrqRyRogkHqRoR8b1sQwu3"

	contract := openwallet.SmartContract{
		Address:  "2",
		Symbol:   "BTC",
		Name:     "TetherUSD",
		Token:    "USDT",
		Decimals: 8,
	}

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "0.0489", "", &contract)
	if err != nil {
		return
	}

	_, err = testSignTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testVerifyTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testSubmitTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

}



func TestTransfer_OMNI(t *testing.T) {
	tm := testInitWalletManager()
	//ms9NeTGFtaMcjrqRyRogkHqRoR8b1sQwu3
	//mp1JDsi7Dr2PkcWu1j4SUSTXJqXjFMaeVx
	//n1ZurJRnQyoRwBrx6B7DMndjBWAxnRbxKJ
	//mxoCkSBmiLQ86N73kXNLHEUgcUBoKdFawH
	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "2m27uVj2xx645dDCEcGD1whPQGcB4fZv16TzBoLGCyKB"
	to := "n1ZurJRnQyoRwBrx6B7DMndjBWAxnRbxKJ"

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "0.0489", "", nil)
	if err != nil {
		return
	}

	_, err = testSignTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testVerifyTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testSubmitTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

}

func TestTransfer_TRON(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "W33vxQiNcgjJgMvowsNerXao6LZjwR61zp"
	accountID := "GEGdASep1uA7RBarNNZuJjgnE8T3DyJGTRGz4JfNE4Me"
	to := "TWVRXXN5tsggjUCDmqbJ4KxPdJKQiynaG6" // t2

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "1", "", nil)
	if err != nil {
		return
	}
	_ = rawTx

	_, err = testSignTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testVerifyTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testSubmitTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

}
