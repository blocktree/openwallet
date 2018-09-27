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
	"fmt"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"time"
)

// CreateTransaction
func (wm *WalletManager) CreateTransaction(appID, walletID, accountID, amount, address, feeRate, memo string) (*openwallet.RawTransaction, error) {

	wrapper, err := wm.newWalletWrapper(appID, "")
	if err != nil {
		return nil, err
	}

	account, err := wrapper.GetAssetsAccountInfo(accountID)
	if err != nil {
		return nil, err
	}

	assetsMgr, err := GetAssetsManager(account.Symbol)
	if err != nil {
		return nil, err
	}

	rawTx := openwallet.RawTransaction{
		Account:  account,
		FeeRate:  feeRate,
		To:       map[string]string{address: amount},
		Required: 1,
	}

	txdecoder := assetsMgr.GetTransactionDecoder()
	if txdecoder == nil {
		return nil, fmt.Errorf("[%s] is not support transaction. ", account.Symbol)
	}

	err = txdecoder.CreateRawTransaction(wrapper, &rawTx)
	if err != nil {
		return nil, err
	}

	log.Debug("transaction has been created successfully")

	return &rawTx, nil
}

// SignTransaction
func (wm *WalletManager) SignTransaction(appID, walletID, accountID, password string, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	account, err := wm.GetAssetsAccountInfo(appID, "", accountID)
	if err != nil {
		return nil, err
	}

	wrapper, err := wm.newWalletWrapper(appID, account.WalletID)
	if err != nil {
		return nil, err
	}

	assetsMgr, err := GetAssetsManager(account.Symbol)
	if err != nil {
		return nil, err
	}

	txdecoder := assetsMgr.GetTransactionDecoder()
	if txdecoder == nil {
		return nil, fmt.Errorf("[%s] is not support transaction. ", account.Symbol)
	}

	//解锁钱包
	err = wrapper.UnlockWallet(password, 5*time.Second)
	if err != nil {
		return nil, err
	}

	err = txdecoder.SignRawTransaction(wrapper, rawTx)
	if err != nil {
		return nil, err
	}

	log.Debug("transaction has been signed successfully")

	return rawTx, nil
}

// VerifyTransaction
func (wm *WalletManager) VerifyTransaction(appID, walletID, accountID string, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	wrapper, err := wm.newWalletWrapper(appID, "")
	if err != nil {
		return nil, err
	}

	account, err := wrapper.GetAssetsAccountInfo(accountID)
	if err != nil {
		return nil, err
	}

	assetsMgr, err := GetAssetsManager(account.Symbol)
	if err != nil {
		return nil, err
	}

	txdecoder := assetsMgr.GetTransactionDecoder()
	if txdecoder == nil {
		return nil, fmt.Errorf("[%s] is not support transaction. ", account.Symbol)
	}

	err = txdecoder.VerifyRawTransaction(wrapper, rawTx)
	if err != nil {
		return nil, err
	}

	log.Debug("transaction has been validated successfully")

	return rawTx, nil
}

// SubmitTransaction
func (wm *WalletManager) SubmitTransaction(appID, walletID, accountID string, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {

	wrapper, err := wm.newWalletWrapper(appID, "")
	if err != nil {
		return nil, err
	}

	account, err := wrapper.GetAssetsAccountInfo(accountID)
	if err != nil {
		return nil, err
	}

	assetsMgr, err := GetAssetsManager(account.Symbol)
	if err != nil {
		return nil, err
	}

	txdecoder := assetsMgr.GetTransactionDecoder()
	if txdecoder == nil {
		return nil, fmt.Errorf("[%s] is not support transaction. ", account.Symbol)
	}

	err = txdecoder.SubmitRawTransaction(wrapper, rawTx)
	if err != nil {
		return nil, err
	}

	log.Debug("transaction has been submitted successfully")

	des := make([]string, 0)

	for to, _ := range rawTx.To {
		des = append(des, to)
	}

	tx := &openwallet.Transaction{
		To:        des,
		Coin:      rawTx.Coin,
		TxID:      rawTx.TxID,
		Decimal:   assetsMgr.Decimal(),
		AccountID: rawTx.Account.AccountID,
	}

	tx.WxID = openwallet.GenTransactionWxID(tx)

	//提取交易单
	scanner := assetsMgr.GetBlockScanner()
	if scanner == nil {
		log.Std.Error("[%s] is not block scan", account.Symbol)
		return tx, nil
	}

	extractData, err := scanner.ExtractTransactionData(rawTx.TxID)
	if err != nil {
		log.Error("ExtractTransactionData failed, unexpected error:", err)
		return tx, nil
	}

	accountTxData, ok := extractData[wm.encodeSourceKey(appID, accountID)]
	if !ok {
		return tx, nil
	}

	txWrapper := openwallet.NewTransactionWrapper(wrapper)
	err = txWrapper.SaveBlockExtractData(accountID, accountTxData)
	if err != nil {
		log.Error("SaveBlockExtractData failed, unexpected error:", err)
		return tx, err
	}

	log.Error("Save new transaction data successfully")

	//更新账户余额
	err = wm.RefreshAssetsAccountBalance(appID, accountID)
	if err != nil {
		log.Error("RefreshAssetsAccountBalance error:", err)
	}

	perfectTx, err := wm.GetTransactionByWxID(appID, tx.WxID)
	if err != nil {
		log.Error("GetTransactionByTxID failed, unexpected error:", err)
		return tx, err
	}

	//log.Error("perfectTx:", perfectTx)

	return perfectTx, nil
}

//GetTransactions
func (wm *WalletManager) GetTransactions(appID string, offset, limit int, cols ...interface{}) ([]*openwallet.Transaction, error) {

	wrapper, err := wm.newWalletWrapper(appID, "")
	if err != nil {
		return nil, err
	}

	txWrapper := openwallet.NewTransactionWrapper(wrapper)
	trx, err := txWrapper.GetTransactions(offset, limit, cols...)
	if err != nil {
		return nil, err
	}

	return trx, nil
}

//GetTransactionByWxID 通过WxID获取交易单
func (wm *WalletManager) GetTransactionByWxID(appID, wxID string) (*openwallet.Transaction, error) {

	wrapper, err := wm.newWalletWrapper(appID, "")
	if err != nil {
		return nil, err
	}

	txWrapper := openwallet.NewTransactionWrapper(wrapper)
	trx, err := txWrapper.GetTransactions(0, -1,"WxID", wxID)
	if err != nil || len(trx) == 0 {
		return nil, err
	}

	return trx[0], nil
}

//GetTxUnspent
func (wm *WalletManager) GetTxUnspent(appID string, offset, limit int, cols ...interface{}) ([]*openwallet.TxOutPut, error) {

	wrapper, err := wm.newWalletWrapper(appID, "")
	if err != nil {
		return nil, err
	}

	txWrapper := openwallet.NewTransactionWrapper(wrapper)
	trx, err := txWrapper.GetTxOutputs(offset, limit, cols...)
	if err != nil {
		return nil, err
	}

	return trx, nil
}

//GetTxSpent
func (wm *WalletManager) GetTxSpent(appID string, offset, limit int, cols ...interface{}) ([]*openwallet.TxInput, error) {

	wrapper, err := wm.newWalletWrapper(appID, "")
	if err != nil {
		return nil, err
	}

	txWrapper := openwallet.NewTransactionWrapper(wrapper)
	trx, err := txWrapper.GetTxInputs(offset, limit, cols...)
	if err != nil {
		return nil, err
	}

	return trx, nil
}
