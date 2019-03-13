/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package bopo

import (
	"github.com/imroc/req"
)

// Transfer
func (wm *WalletManager) toTransfer(wid, toaddr, amount, message string) (*Wallet, error) {
	var wallet *Wallet

	// Register
	if _, err := wm.fullnodeClient.Call("rpc/registar", "POST", req.Param{"id": wid}); err != nil {
		return nil, err
	}

	// To transfer
	request := req.Param{"id": wid, "to": toaddr, "amount": amount, "message": message}
	if _, err := wm.fullnodeClient.Call("rpc/fund", "POST", request); err != nil {
		return nil, err
	} else {
		if w, err := wm.getWalletInfo(wid); err != nil {
			wallet = &Wallet{}
		} else {
			wallet = w
		}
	}

	return wallet, nil
}

// Get detail of a transzation
func (wm *WalletManager) GetTransaction(txid string) (txr *BlockTX, err error) {

	return &BlockTX{}, nil
}
