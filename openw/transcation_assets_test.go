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
	"fmt"
	"testing"

	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-owcdrivers/ontologyTransaction"
)

func testGetAssetsAccountBalance(tm *WalletManager, walletID, accountID string) {
	balance, err := tm.GetAssetsAccountBalance(testApp, walletID, accountID)
	if err != nil {
		log.Error("GetAssetsAccountBalance failed, unexpected error:", err)
		return
	}
	log.Info("balance:", balance)
}

func testGetAssetsAccountTokenBalance(tm *WalletManager, walletID, accountID string, contract openwallet.SmartContract) {
	balance, err := tm.GetAssetsAccountTokenBalance(testApp, walletID, accountID, contract)
	if err != nil {
		log.Error("GetAssetsAccountTokenBalance failed, unexpected error:", err)
		return
	}
	log.Info("token balance:", balance.Balance)
}

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

	log.Infof("rawTx: %+v", rawTx)
	return rawTx, nil
}

func testVerifyTransactionStep(tm *WalletManager, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	//log.Info("rawTx.Signatures:", rawTx.Signatures)

	_, err := tm.VerifyTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx)
	if err != nil {
		log.Error("VerifyTransaction failed, unexpected error:", err)
		return nil, err
	}

	log.Infof("rawTx: %+v", rawTx)
	return rawTx, nil
}

func testSubmitTransactionStep(tm *WalletManager, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	tx, err := tm.SubmitTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx)
	if err != nil {
		log.Error("SubmitTransaction failed, unexpected error:", err)
		return nil, err
	}

	log.Std.Info("tx: %+v", tx)
	log.Info("wxID:", tx.WxID)
	log.Info("txID:", rawTx.TxID)

	return rawTx, nil
}

func TestTransfer_ETH(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "59t47qyjHUMZ6PGAdjkJopE9ffAPUkdUhSinJqcWRYZ1"
	to := "0xd35f9Ea14D063af9B3567064FAB567275b09f03D"

	testGetAssetsAccountBalance(tm, walletID, accountID)

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

	testGetAssetsAccountBalance(tm, walletID, accountID)

	testGetAssetsAccountTokenBalance(tm, walletID, accountID, contract)

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "2.3", "", &contract)
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
	to := "mkXhHAd6o3RnEXtrQJi952AKaH3B9WYSe4"

	testGetAssetsAccountBalance(tm, walletID, accountID)

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
	to := "QfY78pcvLTYrU8YLvCSpb2bKDXrW3Lk6g3"

	testGetAssetsAccountBalance(tm, walletID, accountID)

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "0.01", "", nil)
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
	to := "QfY78pcvLTYrU8YLvCSpb2bKDXrW3Lk6g3"

	contract := openwallet.SmartContract{
		Address:  "f2033ede578e17fa6231047265010445bca8cf1c",
		Symbol:   "QTUM",
		Name:     "QCASH",
		Token:    "QC",
		Decimals: 8,
	}

	testGetAssetsAccountBalance(tm, walletID, accountID)

	testGetAssetsAccountTokenBalance(tm, walletID, accountID, contract)

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
	//n1JrqddgEYqiZqZB3kRrC2zEHtVA7Y1VZts
	//n1VkvsbgRJ6Tjro1mTUUaevPM3wv69z5njB
	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "7VftKuNoDtwZ3mn3wDA4smTDMz4iqCg3fNna1fXicVDg"
	to := "n1H2ddXG2oDyn1oqLgNyw18y8GXqhegmUS5"

	testGetAssetsAccountBalance(tm, walletID, accountID)

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "0.05", "", nil)
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
	to := "n1ZurJRnQyoRwBrx6B7DMndjBWAxnRbxKJ"

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "0.1", "", nil)
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
	to := "mkXhHAd6o3RnEXtrQJi952AKaH3B9WYSe4"

	contract := openwallet.SmartContract{
		Address:  "2",
		Symbol:   "BTC",
		Name:     "Test Omni",
		Token:    "Omni",
		Decimals: 8,
	}

	testGetAssetsAccountBalance(tm, walletID, accountID)

	testGetAssetsAccountTokenBalance(tm, walletID, accountID, contract)

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "0.1", "", &contract)
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
	walletID := "W1eRr8nRrawkQ1Ayf1XKPCjmvKk8aLGExu"
	accountID := "8XPSHP5cR16D4b1V225xig3sgNa45e8Y3P5AbeCzR5gr"
	to := "TNQkiUv4qtDRKWDrKS628FTbDwxLMiqbAz" // t2

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "0.5", "", nil)
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

func TestTransfer_BCH(t *testing.T) {
	tm := testInitWalletManager()
	//1AYKNVJKqfrCtiDP7iKscZA8HB7tbdzZmK
	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "3a9izmhbaDLRitpw811mjRM5JwuGV7i5LB6pb6Awe844"
	to := "1MfQh4Lvk4xmRPcV5VbWdFmQshErHgPCh2"

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "0.2", "", nil)
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

func TestTransfer_ONT(t *testing.T) {
	tm := testONTInitWalletManager()

	walletID := "W33vxQiNcgjJgMvowsNerXao6LZjwR61zp"
	accountID := "7NHXSXEaBViL5koJgexd3qHeGLggHjen99nyjEigdGnn"
	to := "ANeTozd4yxTa5nTyfc3mxzuu7RqabV1iow"

	//  ONT transaction
	contract := openwallet.SmartContract{
		Symbol:   "ONT",
		Address:  ontologyTransaction.ONTContractAddress,
		Decimals: 0,
	}

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "1", "", &contract)
	fmt.Println(rawTx.RawHex)
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

	// ONG withdraw
	contract.Address = ontologyTransaction.ONGContractAddress
	to = "ASyQp8fTtr7gCaCHuRDrtaTr2ZTqSzeE1J"

	rawTx, err = testCreateTransactionStep(tm, walletID, accountID, to, "0", "0", &contract)
	fmt.Println(rawTx.RawHex)
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

	// ONG transaction
	to = "ANeTozd4yxTa5nTyfc3mxzuu7RqabV1iow"

	rawTx, err = testCreateTransactionStep(tm, walletID, accountID, to, "1", "0", &contract)
	fmt.Println(rawTx.RawHex)
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

func TestTransfer_ONT2(t *testing.T) {
	tm := testONTInitWalletManager()

	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "B7kiHeCH1FkuqG9kwyWbqSU96oMBgU9DRJdLqH1jaguh"
	to := "AZnKwMiqcrD3eWiSNdugDb8t5Mw5mUuyCp"

	//  ONT transaction
	contract := openwallet.SmartContract{
		Symbol:  "ONT",
		Address: ontologyTransaction.ONGContractAddress,
	}

	testGetAssetsAccountBalance(tm, walletID, accountID)

	testGetAssetsAccountTokenBalance(tm, walletID, accountID, contract)

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "1", "", &contract)
	//fmt.Println(rawTx.RawHex)
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

func TestTransfer_ONG2(t *testing.T) {
	tm := testONTInitWalletManager()

	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "B7kiHeCH1FkuqG9kwyWbqSU96oMBgU9DRJdLqH1jaguh"
	to := "AHt5qGrP9HG8zqXAEH5gqUwv2FY9pWqRyE"

	//  ONT transaction
	contract := openwallet.SmartContract{
		Symbol:  "ONT",
		Token:   "ONG",
		Address: ontologyTransaction.ONGContractAddress,
	}

	testGetAssetsAccountBalance(tm, walletID, accountID)

	testGetAssetsAccountTokenBalance(tm, walletID, accountID, contract)

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "0", "", &contract)

	if err != nil {
		return
	}
	log.Infof("rawTX: %+v", rawTx)
	_, err = testSignTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testVerifyTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	//_, err = testSubmitTransactionStep(tm, rawTx)
	//if err != nil {
	//	return
	//}

}
