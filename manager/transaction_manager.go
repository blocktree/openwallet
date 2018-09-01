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

import "github.com/blocktree/OpenWallet/openwallet"

// CreateTransaction
func (wm *WalletManager) CreateTransaction(appID, accountID, amount, address, feeRate, memo string) (*openwallet.RawTransaction, error) {
	return nil, nil
}

// SignTransaction
func (wm *WalletManager) SignTransaction(appID, accountID string, rawTx openwallet.RawTransaction) (*openwallet.RawTransaction, error) {
	return nil, nil
}

// VerifyTransaction
func (wm *WalletManager) VerifyTransaction(appID, accountID string, rawTx openwallet.RawTransaction) (*openwallet.RawTransaction, error) {
	return nil, nil
}

// SubmitTransaction
func (wm *WalletManager) SubmitTransaction(appID, accountID string, rawTx openwallet.RawTransaction) (*openwallet.Transaction, error) {
	return nil, nil
}