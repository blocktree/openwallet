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
	"encoding/base64"
	"encoding/hex"
	"fmt"

	// "github.com/blocktree/OpenWallet/assets/tron/protocol/api"
	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder"
	"github.com/shengdoushi/base58"
	"github.com/tidwall/gjson"

	"github.com/imroc/req"
	"github.com/tronprotocol/grpc-gateway/api"
)

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
	fmt.Println("address source = \t ", address)

	address_bytes, _ := addressEncoder.AddressDecode(address, addressEncoder.TRON_mainnetAddress)
	addr_owcryptdec := hex.EncodeToString(address_bytes)
	fmt.Println("address owcryptdec = \t ", addr_owcryptdec)
	r, err := wm.WalletClient.Call2("/wallet/getaccount", req.Param{"address": addr_owcryptdec})
	fmt.Printf("Results: %+v, \n\t Error: %+v\n\n", r, err)

	btcAlphabet := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	myAlphabet := base58.NewAlphabet(btcAlphabet)
	bts2, err := base58.Decode(address, myAlphabet)
	address_base58tohex := hex.EncodeToString(bts2)
	fmt.Println("address base58tohex = \t ", address_base58tohex)
	r, err = wm.WalletClient.Call2("/wallet/getaccount", req.Param{"address": address_base58tohex})
	fmt.Printf("Results: %+v, \n\t Error: %+v\n\n", r, err)

	unBase64Bytes, _ := base64.StdEncoding.DecodeString(address)
	addr_base64tohex := hex.EncodeToString(unBase64Bytes)
	fmt.Println("address base64tohex = \t ", addr_base64tohex)
	r, err = wm.WalletClient.Call2("/wallet/getaccount", req.Param{"address": addr_base64tohex})
	fmt.Printf("Results: %+v, \n\t Error: %+v\n\n", r, err)

	address_srctohex := hex.EncodeToString([]byte(address))
	fmt.Println("address srctohex = \t ", address_srctohex)
	r, err = wm.WalletClient.Call2("/wallet/getaccount", req.Param{"address": address_srctohex})
	fmt.Printf("Results: %+v, \n\t Error: %+v\n\n", r, err)

	// address source =          TSdXzXKSQ3RQzQ5Ge8TiYfMQEjofSVQ8ax
	// address owcryptdec =        b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7a   // B6C1ABF9FB31C9077DFB3C25469E6E943FFBFA7A
	// address base58tohex =     41b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7af078d025
	// address base64tohex =     4d2757cd7292437450cd0e467bc4e261f310123a1f49543c
	// address srctohex =        545364587a584b53513352517a5135476538546959664d51456a6f66535651386178

	// 6fa73569b7fd7dbdf573dd3bedd7dbddcdb9e3af5ee9ef78ddf7db7daeda // base64tohex

	params := req.Param{
		"address": address_srctohex,
	}

	r, err = wm.WalletClient.Call2("/wallet/getaccount", params)
	// r, err = wm.WalletClient.Call2("/wallet/getbalance", params)
	// r, err = wm.WalletClient.Call2("/wallet/listaccounts", params)

	if err != nil {
		return nil, err
	}

	fmt.Println("Result = ", string(r))

	account = &api.AccountNetMessage{}
	if err := gjson.Unmarshal(r, account); err != nil {
		return nil, err
	}

	return account, nil
}

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
func (wm *WalletManager) CreateAccount(owner_address, account_address string) (txRaw string, err error) {

	owner_address_bytes, _ := addressEncoder.AddressDecode(owner_address, addressEncoder.TRON_mainnetAddress)
	owner_address = hex.EncodeToString(owner_address_bytes)
	account_address_bytes, _ := addressEncoder.AddressDecode(account_address, addressEncoder.TRON_mainnetAddress)
	account_address = hex.EncodeToString(account_address_bytes)

	params := req.Param{
		"owner_address":   owner_address,
		"account_address": account_address,
	}

	r, err := wm.WalletClient.Call2("/wallet/createaccount", params)
	if err != nil {
		return "", err
	}
	fmt.Println("Result = ", string(r))

	return "", nil
}

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
func (wm *WalletManager) UpdateAccount(account_name, owner_address string) (tx string, err error) {

	params := req.Param{
		"account_name":  account_name,
		"owner_address": owner_address,
	}

	r, err := wm.WalletClient.Call2("/wallet/updateaccount", params)
	if err != nil {
		return "", err
	}
	fmt.Println("Result = ", r)

	return "", nil
}
