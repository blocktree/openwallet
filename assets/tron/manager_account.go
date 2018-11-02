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

	"github.com/blocktree/go-owcdrivers/addressEncoder"
	"github.com/tidwall/gjson"

	"github.com/imroc/req"
	"github.com/tronprotocol/grpc-gateway/api"
	"github.com/tronprotocol/grpc-gateway/core"
)

func convertAddrToHex(address string) string {
	to_address_bytes, err := addressEncoder.AddressDecode(address, addressEncoder.TRON_mainnetAddress)
	if err != nil {
		log.Println(err)
	}
	to_address_bytes = append([]byte{0x41}, to_address_bytes...)
	return hex.EncodeToString(to_address_bytes)
}

// Done
// Function：Query bandwidth information.
// demo: curl -X POST http://127.0.0.1:8090/wallet/getaccountnet -d ‘
// 	{“address”: “4112E621D5577311998708F4D7B9F71F86DAE138B5”}’
// Parameters：
// 	Account address，converted to a hex string
// Return value：
// 	Bandwidth information for the account.
// 	If a field doesn’t appear, then the corresponding value is 0.
// 	{“freeNetUsed”: 557,”freeNetLimit”: 5000,”NetUsed”: 353,”NetLimit”: 5239157853,”TotalNetLimit”: 43200000000,”TotalNetWeight”: 41228}
func (wm *WalletManager) GetAccountNet(address string) (account *api.AccountNetMessage, err error) {

	address = convertAddrToHex(address)

	params := req.Param{"address": address}
	r, err := wm.WalletClient.Call("/wallet/getaccountnet", params)
	if err != nil {
		return nil, err
	}

	account = &api.AccountNetMessage{}
	if err := gjson.Unmarshal([]byte(r.Raw), account); err != nil {
		return nil, err
	}

	return account, nil
}

// Done
// Function：Query bandwidth information.
// Parameters：
// 	Account address，converted to a base64 string
// Return value：
func (wm *WalletManager) GetAccount(address string) (account *core.Account, err error) {
	address = convertAddrToHex(address)

	params := req.Param{"address": address}
	r, err := wm.WalletClient.Call("/wallet/getaccount", params)
	if err != nil {
		return nil, err
	}

	account = &core.Account{}
	if err := gjson.Unmarshal([]byte(r.Raw), account); err != nil {
		return nil, err
	}

	return account, nil
}

// Done!
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
func (wm *WalletManager) CreateAccount(owner_address, account_address string) (txRaw *gjson.Result, err error) {

	owner_address = convertAddrToHex(owner_address)
	account_address = convertAddrToHex(account_address)

	params := req.Param{"owner_address": owner_address, "account_address": account_address}
	r, err := wm.WalletClient.Call("/wallet/createaccount", params)
	if err != nil {
		return nil, err
	}

	if r.Get("Error").String() != "" {
		return nil, errors.New(r.Get("Error").String())
	}

	return r, nil
}

// Done
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
func (wm *WalletManager) UpdateAccount(account_name, owner_address string) (tx *gjson.Result, err error) {

	params := req.Param{
		"account_name":  hex.EncodeToString([]byte(account_name)),
		"owner_address": convertAddrToHex(owner_address),
	}

	r, err := wm.WalletClient.Call("/wallet/updateaccount", params)
	if err != nil {
		return nil, err
	}

	return r, nil
}
