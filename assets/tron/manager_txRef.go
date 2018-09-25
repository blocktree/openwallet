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

	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder"
	"github.com/tronprotocol/grpc-gateway/core"
)

// public static Contract.TransferContract createTransferContract(byte[] to, byte[] owner, long amount) {
//     Contract.TransferContract.Builder builder = Contract.TransferContract.newBuilder();
//     ByteString bsTo = ByteString.copyFrom(to);
//     ByteString bsOwner = ByteString.copyFrom(owner);
//     builder.setToAddress(bsTo);
//     builder.setOwnerAddress(bsOwner);
//     builder.setAmount(amount);

//     return builder.build();
// }
func (wm *WalletManager) CreateTransactionRef(to_address, owner_address string, amount uint64) (raw string, err error) {
	to_address_bytes, err := addressEncoder.AddressDecode(to_address, addressEncoder.TRON_mainnetAddress)
	if err != nil {
		return "", err
	} else {
		fmt.Println("X2 = ", to_address_bytes)
		to_address_bytes = append([]byte{41}, to_address_bytes...)
	}
	fmt.Println("X2 = ", to_address_bytes)

	owner_address_bytes, err := addressEncoder.AddressDecode(owner_address, addressEncoder.TRON_mainnetAddress)
	if err != nil {
		return "", err
	} else {
		owner_address_bytes = append([]byte{41}, owner_address_bytes...)
	}

	tx := &core.TransferContract{
		OwnerAddress: to_address_bytes,
		ToAddress:    owner_address_bytes,
		Amount:       int64(amount),
	}

	// // core.Transaction
	// tx := &core.Transaction_Contract{
	// 	Type:         core.Transaction_Contract_TransferContract,
	// 	Parameter:    nil,
	// 	Provider:     nil,
	// 	ContractName: nil,
	// }
	fmt.Println("tx = ", tx)

	return raw, nil
}

func (wm *WalletManager) GetTransactionSignRef(transaction, privateKey string) (rawSinged []byte, err error) {

	return rawSinged, nil
}

func (wm *WalletManager) BroadcastTransactionRef(signature, txID, raw_data string) error {

	return nil
}
