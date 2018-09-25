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

package tron

import (
	"fmt"

	"github.com/tronprotocol/grpc-gateway/core"
)

func (wm *WalletManager) CreateTransactionRef(to_address, owner_address string, amount uint64) (raw string, err error) {

	// core.Transaction
	tx := &core.Transaction_Contract{
		Type:         core.Transaction_Contract_TransferContract,
		Parameter:    nil,
		Provider:     nil,
		ContractName: nil,
	}
	fmt.Println("tx = ", tx)

	return raw, nil
}

func (wm *WalletManager) GetTransactionSignRef(transaction, privateKey string) (rawSinged []byte, err error) {

	return rawSinged, nil
}

func (wm *WalletManager) BroadcastTransactionRef(signature, txID, raw_data string) error {

	return nil
}
