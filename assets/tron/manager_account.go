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
	"encoding/hex"
	"errors"
	"log"

	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-owcdrivers/addressEncoder"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
)

func convertAddrToHex(address string) string {
	toAddressBytes, err := addressEncoder.AddressDecode(address, addressEncoder.TRON_mainnetAddress)
	if err != nil {
		log.Println(err)
	}
	toAddressBytes = append([]byte{0x41}, toAddressBytes...)
	return hex.EncodeToString(toAddressBytes)
}

// GetAccountNet Done!
// Function：Query bandwidth information.
// demo: curl -X POST http://127.0.0.1:8090/wallet/getaccountnet -d ‘
// 	{“address”: “4112E621D5577311998708F4D7B9F71F86DAE138B5”}’
// Parameters：
// 	Account address，converted to a hex string
// Return value：
// 	Bandwidth information for the account.
// 	If a field doesn’t appear, then the corresponding value is 0.
// 	{“freeNetUsed”: 557,”freeNetLimit”: 5000,”NetUsed”: 353,”NetLimit”: 5239157853,”TotalNetLimit”: 43200000000,”TotalNetWeight”: 41228}
func (wm *WalletManager) GetAccountNet(address string) (accountNet *AccountNet, err error) {

	address = convertAddrToHex(address)

	params := req.Param{"address": address}
	r, err := wm.WalletClient.Call("/wallet/getaccountnet", params)
	if err != nil {
		return nil, err
	}
	accountNet = NewAccountNet(r)
	return accountNet, nil
}

// GetAccount Done!
// Function：Query bandwidth information.
// Parameters：
// 	Account address，converted to a base64 string
// Return value：
func (wm *WalletManager) GetAccount(address string) (account *openwallet.AssetsAccount, err error) {
	address = convertAddrToHex(address)

	params := req.Param{"address": address}
	r, err := wm.WalletClient.Call("/wallet/getaccount", params)
	if err != nil {
		return nil, err
	}
	account = &openwallet.AssetsAccount{}

	// // type Account struct {
	// // 	AccountName []byte      `protobuf:"bytes,1,opt,name=account_name,json=accountName,proto3" json:"account_name,omitempty"`
	// // 	Type        AccountType `protobuf:"varint,2,opt,name=type,proto3,enum=protocol.AccountType" json:"type,omitempty"`
	// // 	// the create address
	// // 	Address []byte `protobuf:"bytes,3,opt,name=address,proto3" json:"address,omitempty"`
	// // 	// the trx balance
	// // 	Balance int64 `protobuf:"varint,4,opt,name=balance,proto3" json:"balance,omitempty"`
	// // 	// the votes
	// // 	Votes []*Vote `protobuf:"bytes,5,rep,name=votes,proto3" json:"votes,omitempty"`
	// // 	// the other asset owned by this account
	// // 	Asset map[string]int64 `protobuf:"bytes,6,rep,name=asset,proto3" json:"asset,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	// // 	// latest asset operation time
	// // 	// the frozen balance
	// // 	Frozen []*Account_Frozen `protobuf:"bytes,7,rep,name=frozen,proto3" json:"frozen,omitempty"`
	// // 	// bandwidth, get from frozen
	// // 	NetUsage int64 `protobuf:"varint,8,opt,name=net_usage,json=netUsage,proto3" json:"net_usage,omitempty"`
	// // 	// this account create time
	// // 	CreateTime int64 `protobuf:"varint,9,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty"`
	// // 	// this last operation time, including transfer, voting and so on. //FIXME fix grammar
	// // 	LatestOprationTime int64 `protobuf:"varint,10,opt,name=latest_opration_time,json=latestOprationTime,proto3" json:"latest_opration_time,omitempty"`
	// // 	// witness block producing allowance
	// // 	Allowance int64 `protobuf:"varint,11,opt,name=allowance,proto3" json:"allowance,omitempty"`
	// // 	// last withdraw time
	// // 	LatestWithdrawTime int64 `protobuf:"varint,12,opt,name=latest_withdraw_time,json=latestWithdrawTime,proto3" json:"latest_withdraw_time,omitempty"`
	// // 	// not used so far
	// // 	Code        []byte `protobuf:"bytes,13,opt,name=code,proto3" json:"code,omitempty"`
	// // 	IsWitness   bool   `protobuf:"varint,14,opt,name=is_witness,json=isWitness,proto3" json:"is_witness,omitempty"`
	// // 	IsCommittee bool   `protobuf:"varint,15,opt,name=is_committee,json=isCommittee,proto3" json:"is_committee,omitempty"`
	// // 	// frozen asset(for asset issuer)
	// // 	FrozenSupply []*Account_Frozen `protobuf:"bytes,16,rep,name=frozen_supply,json=frozenSupply,proto3" json:"frozen_supply,omitempty"`
	// // 	// asset_issued_name
	// // 	AssetIssuedName          []byte           `protobuf:"bytes,17,opt,name=asset_issued_name,json=assetIssuedName,proto3" json:"asset_issued_name,omitempty"`
	// // 	LatestAssetOperationTime map[string]int64 `protobuf:"bytes,18,rep,name=latest_asset_operation_time,json=latestAssetOperationTime,proto3" json:"latest_asset_operation_time,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	// // 	FreeNetUsage             int64            `protobuf:"varint,19,opt,name=free_net_usage,json=freeNetUsage,proto3" json:"free_net_usage,omitempty"`
	// // 	FreeAssetNetUsage        map[string]int64 `protobuf:"bytes,20,rep,name=free_asset_net_usage,json=freeAssetNetUsage,proto3" json:"free_asset_net_usage,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	// // 	LatestConsumeTime        int64            `protobuf:"varint,21,opt,name=latest_consume_time,json=latestConsumeTime,proto3" json:"latest_consume_time,omitempty"`
	// // 	LatestConsumeFreeTime    int64            `protobuf:"varint,22,opt,name=latest_consume_free_time,json=latestConsumeFreeTime,proto3" json:"latest_consume_free_time,omitempty"`
	// // 	XXX_NoUnkeyedLiteral     struct{}         `json:"-"`
	// // 	XXX_unrecognized         []byte           `json:"-"`
	// // 	XXX_sizecache            int32            `json:"-"`
	// // }
	// account = &core.Account{}
	// if err := gjson.Unmarshal([]byte(r.Raw), account); err != nil {
	// 	return nil, err
	// }
	account.Balance = r.Get("balance").String()
	account.Symbol = wm.Config.Symbol
	account.Required = 1

	return account, nil
}

// CreateAccount Done!
// Function：Create an account. Uses an already activated account to create a new account
// demo：curl -X POST http://127.0.0.1:8090/wallet/createaccount -d ‘
// 	{
// 		“owner_address”:”41d1e7a6bc354106cb410e65ff8b181c600ff14292”,
// 		“account_address”: “41e552f6487585c2b58bc2c9bb4492bc1f17132cd0”
// 	}’
// Parameters：
// 	Owner_address is an activated account，converted to a hex String;
// 	account_address is the address of the new account, converted to a hex string, this address needs to be calculated in advance
// Return value：Create account Transaction raw data
func (wm *WalletManager) CreateAccount(ownerAddress, accountAddress string) (txRaw *gjson.Result, err error) {

	ownerAddress = convertAddrToHex(ownerAddress)
	accountAddress = convertAddrToHex(accountAddress)

	params := req.Param{"owner_address": ownerAddress, "account_address": accountAddress}
	r, err := wm.WalletClient.Call("/wallet/createaccount", params)
	if err != nil {
		return nil, err
	}

	if r.Get("Error").String() != "" {
		return nil, errors.New(r.Get("Error").String())
	}
	return r, nil
}

// UpdateAccount Done!
// Function：Modify account name
// demo：curl -X POSThttp://127.0.0.1:8090/wallet/updateaccount -d ‘
// 	{
// 		“account_name”: “0x7570646174654e616d6531353330383933343635353139” ,
// 		”owner_address”:”41d1e7a6bc354106cb410e65ff8b181c600ff14292”
// 	}’
// Parameters：
// 	account_name is the name of the account, converted into a hex string；
// 	owner_address is the account address of the name to be modified, converted to a hex string.
// Return value：modified Transaction Object
func (wm *WalletManager) UpdateAccount(accountName, ownerAddress string) (tx *gjson.Result, err error) {

	params := req.Param{
		"account_name":  hex.EncodeToString([]byte(accountName)),
		"owner_address": convertAddrToHex(ownerAddress),
	}

	r, err := wm.WalletClient.Call("/wallet/updateaccount", params)
	if err != nil {
		return nil, err
	}

	if r.Get("raw_data").Map()["contract"].Array()[0].Map()["parameter"].Map()["value"].Map()["owner_address"].String() != convertAddrToHex(ownerAddress) {
		return nil, errors.New("UpdateAccount: Update failed")
	}
	return r, nil
}
