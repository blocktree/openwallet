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
	"fmt"

	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
	"github.com/tronprotocol/grpc-gateway/core"
)

// Function：Count all transactions (number) on the network
// demo: curl -X POST http://127.0.0.1:8090/wallet/totaltransaction
// Parameters：None
// Return value：
// 	Total number of transactions.
func (wm *WalletManager) GetTotalTransaction() (num uint64, err error) {

	r, err := wm.WalletClient.Call2("/wallet/totaltransaction", nil)
	if err != nil {
		return 0, err
	}
	num = gjson.ParseBytes(r).Get("num").Uint()
	// fmt.Println("CCC = ", gjson.ParseBytes(r).Get("num").Uint())
	return num, nil
}

// Function：Query transaction by ID
// 	demo: curl -X POST http://127.0.0.1:8090/wallet/gettransactionbyid -d ‘
// 		{“value”: “d5ec749ecc2a615399d8a6c864ea4c74ff9f523c2be0e341ac9be5d47d7c2d62”}’
// Parameters：Transaction ID.
// Return value：Transaction information.
func (wm *WalletManager) GetTransactionByID(txID string) (tx *core.Transaction, err error) {

	params := req.Param{
		"value": txID,
	}
	r, err := wm.WalletClient.Call2("/wallet/gettransactionbyid", params)
	if err != nil {
		return nil, err
	}

	tx = &core.Transaction{}
	if err := gjson.Unmarshal(r, tx); err != nil {
		return nil, err
	}

	return tx, nil
}

// Function：Creates a transaction of transfer. If the recipient address does not exist, a corresponding account will be created on the blockchain.
// demo: curl -X POST http://127.0.0.1:8090/wallet/createtransaction -d ‘
// 	{“to_address”: “41e9d79cc47518930bc322d9bf7cddd260a0260a8d”,
// 	“owner_address”: “41D1E7A6BC354106CB410E65FF8B181C600FF14292”,
// 	“amount”: 1000 }’ P
// Parameters：
// 	To_address is the transfer address, converted to a hex string;
// 	owner_address is the transfer transfer address, converted to a hex string;
// 	amount is the transfer amount
// Return value：
// 	Transaction contract data
func (wm *WalletManager) CreateTransaction(to_address, owner_address string, amount uint64) (raw string, err error) {

	to_address_bytes, _ := addressEncoder.AddressDecode(to_address, addressEncoder.TRON_mainnetAddress)
	to_address = hex.EncodeToString(to_address_bytes)
	owner_address_bytes, _ := addressEncoder.AddressDecode(owner_address, addressEncoder.TRON_mainnetAddress)
	owner_address = hex.EncodeToString(owner_address_bytes)

	params := req.Param{
		"to_address":    to_address,
		"owner_address": owner_address,
		"amount":        1,
	}

	rawBytes, err := wm.WalletClient.Call2("/wallet/createtransaction", params)
	if err != nil {
		return "", err
	}
	fmt.Println("XXX = ", string(rawBytes))
	raw = hex.EncodeToString(rawBytes)

	return raw, nil
}

// Function：Sign the transaction, the api has the risk of leaking the private key, please make sure to call the api in a secure environment
// 	demo: curl -X POST http://127.0.0.1:8090/wallet/gettransactionsign -d ‘
// 		{ “transaction” : {“txID”:”454f156bf1256587ff6ccdbc56e64ad0c51e4f8efea5490dcbc720ee606bc7b8”,
// 					 ”raw_data”:{
// 						“contract”:[{“parameter”:{“value”:{“amount”:1000,
// 											     ”owner_address”:”41e552f6487585c2b58bc2c9bb4492bc1f17132cd0”,
// 											     ”to_address”:”41d1e7a6bc354106cb410e65ff8b181c600ff14292”},
// 										  ”type_url”:”type.googleapis.com/protocol.TransferContract”},
// 								 ”type”:”TransferContract”}],
// 						”ref_block_bytes”:”267e”,
// 						”ref_block_hash”:”9a447d222e8de9f2”,
// 						”expiration”:1530893064000,
// 						”timestamp”:1530893006233}}
// 		“privateKey” : “your private key”} }’
// Parameters：
// 	Transaction is a contract created by http api,
// 	privateKey is the user private key
// Return value：Signed Transaction contract data
func (wm *WalletManager) GetTransactionSign(transaction, privateKey string) (rawSinged []byte, err error) {

	params := req.Param{
		"transaction": transaction,
		"privateKey":  privateKey,
	}

	r, err := wm.WalletClient.Call2("/wallet/gettransactionsign", params)
	if err != nil {
		return nil, err
	}
	fmt.Println("EEEE = ", r)

	return rawSinged, nil
}

// Function：Broadcast the signed transaction
// 	demo：curl -X POST http://127.0.0.1:8090/wallet/broadcasttransaction -d ‘
// 		{“signature”:[“97c825b41c77de2a8bd65b3df55cd4c0df59c307c0187e42321dcc1cc455ddba583dd9502e17cfec5945b34cad0511985a6165999092a6dec84c2bdd97e649fc01”],
// 		 ”txID”:”454f156bf1256587ff6ccdbc56e64ad0c51e4f8efea5490dcbc720ee606bc7b8”,
// 		 ”raw_data”:{“contract”:[{
// 			 				“parameter”:{
// 								“value”:{“amount”:1000,
// 									   ”owner_address”:”41e552f6487585c2b58bc2c9bb4492bc1f17132cd0”,
// 									   ”to_address”:”41d1e7a6bc354106cb410e65ff8b181c600ff14292”},
// 								”type_url”:”type.googleapis.com/protocol.TransferContract”},
// 							”type”:”TransferContract”
// 						}],
// 				”ref_block_bytes”:”267e”,
// 				”ref_block_hash”:”9a447d222e8de9f2”,
// 				”expiration”:1530893064000,
// 				”timestamp”:1530893006233}
// 		}’
// Parameters：Signed Transaction contract data
// Return value：broadcast success or failure
func (wm *WalletManager) BroadcastTransaction(signature, txID, raw_data string) error {

	params := req.Param{
		"signature": signature,
		"txID":      txID,
		"raw_data":  raw_data,
	}

	r, err := wm.WalletClient.Call2("/wallet/broadcasttransaction", params)
	if err != nil {
		return err
	}
	fmt.Println("EEEE = ", r)

	return nil
}
