/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */
//Some common define of ontology-go-sdk
package common

var (
	VERSION_TRANSACTION  = byte(0)
	VERSION_CONTRACT_ONT = byte(0)
	VERSION_CONTRACT_ONG = byte(0)
)

const (
	NATIVE_TRANSFER      = "transfer"
	NATIVE_TRANSFER_FROM = "transferFrom"
	NATIVE_APPROVE       = "approve"
	NATIVE_ALLOWANCE     = "allowance"
)

//NeoVM invoke smart contract return type
type NeoVMReturnType byte

const (
	NEOVM_TYPE_BOOL       NeoVMReturnType = 1
	NEOVM_TYPE_INTEGER    NeoVMReturnType = 2
	NEOVM_TYPE_BYTE_ARRAY NeoVMReturnType = 3
	NEOVM_TYPE_STRING     NeoVMReturnType = 4
)

//Balance object for account
type Balance struct {
	Ont uint64
	Ong uint64
}

//BalanceRsp response object for balance request
type BalanceRsp struct {
	Ont string `json:"ont"`
	Ong string `json:"ong"`
}

//SmartContactEvent object for event of transaction
type SmartContactEvent struct {
	TxHash      string
	State       byte
	GasConsumed uint64
	Notify      []*NotifyEventInfo
}

type NotifyEventInfo struct {
	ContractAddress string
	States          interface{}
}

//MerkleProof return struct
type MerkleProof struct {
	Type             string
	TransactionsRoot string
	BlockHeight      uint32
	CurBlockRoot     string
	CurBlockHeight   uint32
	TargetHashes     []string
}
