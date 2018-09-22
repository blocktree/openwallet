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

	// "github.com/blocktree/OpenWallet/assets/tron/protocol/api"
	"github.com/tidwall/gjson"
	"github.com/tronprotocol/grpc-gateway/api"

	"github.com/imroc/req"
)

// Function: Create address from a specified password string (NOT PRIVATE KEY)
// Demo: curl -X POST http://127.0.0.1:8090/wallet/createaddress -d ‘
// 	{“value”: “7465737470617373776f7264” }’
// Parameters:
// 	value is the password, converted from ascii to hex
// Return value：
// 	value is the corresponding address for the password, encoded in hex.
// 	Convert it to base58 to use as the address.
// Warning:
// 	Please control risks when using this API. To ensure environmental security, please do not invoke APIs provided by other or invoke this very API on a public network.
func (wm *WalletManager) CreateAddress(passValue string) (addr string, err error) {

	params := req.Param{
		"value": passValue,
	}
	r, err := wm.WalletClient.Call2("/wallet/createaddress", params)
	if err != nil {
		return "nil", err
	}
	fmt.Println("Result = ", r)

	return "", nil
}

// Function: Generates a random private key and address pair
// Demo：curl -X POST -k http://127.0.0.1:8090/wallet/generateaddress
// Parameters:
// 	no parameters.
// Return value：
// 	value is the corresponding address for the password, encoded in hex.
// 	Convert it to base58 to use as the address.
// Warning:
// 	Please control risks when using this API.
// 	To ensure environmental security, please do not invoke APIs provided by other or invoke this very API on a public network.
func (wm *WalletManager) GenerateAddress() (addr *api.AddressPrKeyPairMessage, err error) {

	// res = map[string]string{"addr": "", "privatekey": ""}

	r, err := wm.WalletClient.Call2("/wallet/generateaddress", nil)
	if err != nil {
		return nil, err
	}

	addr = &api.AddressPrKeyPairMessage{}
	if err := gjson.Unmarshal(r, addr); err != nil {
		return nil, err
	}

	return addr, nil
}

// Function：validate address
// Demo: curl -X POST http://127.0.0.1:8090/wallet/validateaddress -d ‘
// 	{“address”: “4189139CB1387AF85E3D24E212A008AC974967E561”}’
// Parameters：
// 	The address, should be in base58checksum, hexString or base64 format.
// Return value: ture or false
func (wm *WalletManager) ValidateAddress(address string) (err error) {

	params := req.Param{
		"address": address,
	}
	r, err := wm.WalletClient.Call2("/wallet/validateaddress", params)
	if err != nil {
		return err
	}
	fmt.Println("Result = ", r)

	return nil
}
