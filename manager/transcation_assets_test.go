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
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"testing"
)

func testCreateTransactionStep(walletID, accountID, to, amount, feeRate string, contract *openwallet.SmartContract) (*openwallet.RawTransaction, error) {

	err := tm.RefreshAssetsAccountBalance(testApp, accountID)
	if err != nil {
		log.Error("RefreshAssetsAccountBalance failed, unexpected error:", err)
		return nil, err
	}

	rawTx, err := tm.CreateTransaction(testApp, walletID, accountID, amount, to, feeRate, "", contract)

	if err != nil {
		log.Error("CreateTransaction failed, unexpected error:", err)
		return nil, err
	}

	return rawTx, nil
}

func testSignTransactionStep(rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	_, err := tm.SignTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, "12345678", rawTx)
	if err != nil {
		log.Error("SignTransaction failed, unexpected error:", err)
		return nil, err
	}

	log.Info("rawTx:", rawTx)
	return rawTx, nil
}

func testVerifyTransactionStep(rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	//log.Info("rawTx.Signatures:", rawTx.Signatures)

	_, err := tm.VerifyTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx)
	if err != nil {
		log.Error("VerifyTransaction failed, unexpected error:", err)
		return nil, err
	}

	log.Info("rawTx:", rawTx)
	return rawTx, nil
}

func testSubmitTransactionStep(rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

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

	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "59t47qyjHUMZ6PGAdjkJopE9ffAPUkdUhSinJqcWRYZ1"
	to := "d35f9Ea14D063af9B3567064FAB567275b09f03D"

	rawTx, err := testCreateTransactionStep(walletID, accountID, to, "0.0003", "", nil)
	if err != nil {
		return
	}

	_, err = testSignTransactionStep(rawTx)
	if err != nil {
		return
	}

	_, err = testVerifyTransactionStep(rawTx)
	if err != nil {
		return
	}

	_, err = testSubmitTransactionStep(rawTx)
	if err != nil {
		return
	}

}

func TestTransfer_ERC20(t *testing.T) {

	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "59t47qyjHUMZ6PGAdjkJopE9ffAPUkdUhSinJqcWRYZ1"
	to := "0xd35f9Ea14D063af9B3567064FAB567275b09f03D"

	contract := openwallet.SmartContract{
		Address:  "0x4092678e4E78230F46A1534C0fbc8fA39780892B",
		Symbol:   "ETH",
		Name:     "OCoin",
		Token:    "OCN",
		Decimals: 18,
	}

	rawTx, err := testCreateTransactionStep(walletID, accountID, to, "1.1", "", &contract)
	if err != nil {
		return
	}

	_, err = testSignTransactionStep(rawTx)
	if err != nil {
		return
	}

	_, err = testVerifyTransactionStep(rawTx)
	if err != nil {
		return
	}

	_, err = testSubmitTransactionStep(rawTx)
	if err != nil {
		return
	}

}

func TestTransfer_LTC(t *testing.T) {

	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "EbUsW3YaHQ61eNt3f4hDXJAFh9LGmLZWH1VTTSnQmhnL"
	to := "n3ctLfAj8ksiXbVRCSQwEiyh2ZV8REAcUC"

	rawTx, err := testCreateTransactionStep(walletID, accountID, to, "1.1", "0.001", nil)
	if err != nil {
		return
	}

	_, err = testSignTransactionStep(rawTx)
	if err != nil {
		return
	}

	_, err = testVerifyTransactionStep(rawTx)
	if err != nil {
		return
	}

	_, err = testSubmitTransactionStep(rawTx)
	if err != nil {
		return
	}

}

func TestTransfer_QTUM(t *testing.T) {
	//Qf6t5Ww14ZWVbG3kpXKoTt4gXeKNVxM9QJ
	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "2by6wzbzw7cnWkxiA31xMHpFmE99bqL3BnjkUJnJtEN6"
	to := "Qe2M15jwosabZeKNZ6xDEGsLx9SV95AEkw"

	rawTx, err := testCreateTransactionStep(walletID, accountID, to, "0.03", "", nil)
	if err != nil {
		return
	}

	_, err = testSignTransactionStep(rawTx)
	if err != nil {
		return
	}

	_, err = testVerifyTransactionStep(rawTx)
	if err != nil {
		return
	}

	_, err = testSubmitTransactionStep(rawTx)
	if err != nil {
		return
	}

}

func TestWalletManager_GetQRC20TokenBalance(t *testing.T) {
	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "2by6wzbzw7cnWkxiA31xMHpFmE99bqL3BnjkUJnJtEN6"

	contract := openwallet.SmartContract{
		Address:  "f2033ede578e17fa6231047265010445bca8cf1c",
		Symbol:   "QTUM",
		Name:     "QCASH",
		Token:    "QC",
		Decimals: 8,
	}

	balance, err := tm.GetAssetsAccountTokenBalance(testApp, walletID, accountID, contract)
	if err != nil {
		log.Error("GetAssetsAccountTokenBalance failed, unexpected error:", err)
		return
	}
	log.Info("balance:", balance.Balance)
}

func TestTransfer_QRC20(t *testing.T) {

	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "2by6wzbzw7cnWkxiA31xMHpFmE99bqL3BnjkUJnJtEN6"
	to := "QfY78pcvLTYrU8YLvCSpb2bKDXrW3Lk6g3"

	contract := openwallet.SmartContract{
		Address:  "f2033ede578e17fa6231047265010445bca8cf1c",
		Symbol:   "QTUM",
		Name:     "QCASH",
		Token:    "QC",
		Decimals: 8,
	}

	rawTx, err := testCreateTransactionStep(walletID, accountID, to, "2.345", "", &contract)
	if err != nil {
		return
	}

	_, err = testSignTransactionStep(rawTx)
	if err != nil {
		return
	}

	_, err = testVerifyTransactionStep(rawTx)
	if err != nil {
		return
	}

	_, err = testSubmitTransactionStep(rawTx)
	if err != nil {
		return
	}

}