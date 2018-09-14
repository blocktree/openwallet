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

	return &rawTx, nil
}

// SignTransaction
func (wm *WalletManager) SignTransaction(appID, walletID, accountID, password string, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	wrapper, err := wm.newWalletWrapper(appID, walletID)
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

	//解锁钱包
	err = wrapper.UnlockWallet(password, 5 * time.Second)
	if err != nil {
		return nil, err
	}

	err = txdecoder.SignRawTransaction(wrapper, rawTx)
	if err != nil {
		return nil, err
	}

	return rawTx, nil
}

// VerifyTransaction
func (wm *WalletManager) VerifyTransaction(appID, walletID, accountID string, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	wrapper, err := wm.newWalletWrapper(appID, walletID)
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

	return rawTx, nil
}

// SubmitTransaction
func (wm *WalletManager) SubmitTransaction(appID, walletID, accountID string, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {
	return nil, nil
}
