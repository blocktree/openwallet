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

	"github.com/blocktree/OpenWallet/log"
	"github.com/golang/protobuf/proto"
	"github.com/imroc/req"
	"github.com/tronprotocol/grpc-gateway/core"
)

// GetTotalTransaction Done!
// Function：Count all transactions (number) on the network
// demo: curl -X POST http://127.0.0.1:8090/wallet/totaltransaction
// Parameters：Nones
// Return value：
// 	Total number of transactions.
func (wm *WalletManager) GetTotalTransaction() (num uint64, err error) {

	r, err := wm.WalletClient.Call("/wallet/totaltransaction", nil)
	if err != nil {
		return 0, err
	}

	num = r.Get("num").Uint()
	return num, nil
}

// GetTransactionByID Done!
// Function：Query transaction by ID
// 	demo: curl -X POST http://127.0.0.1:8090/wallet/gettransactionbyid -d ‘
// 		{“value”: “d5ec749ecc2a615399d8a6c864ea4c74ff9f523c2be0e341ac9be5d47d7c2d62”}’
// Parameters：Transaction ID.
// Return value：Transaction information.
func (wm *WalletManager) GetTransactionByID(txID string) (tx *Transaction, err error) {

	params := req.Param{"value": txID}
	r, err := wm.WalletClient.Call("/wallet/gettransactionbyid", params)
	if err != nil {
		return nil, err
	}

	tx = NewTransaction(r)
	return tx, err
}

// CreateTransaction Writing!
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
func (wm *WalletManager) CreateTransaction(toAddress, ownerAddress string, amount float64) (raw string, err error) {

	params := req.Param{
		"to_address":    toAddress,
		"owner_address": ownerAddress,
		"amount":        amount * 1000000,
	}

	r, err := wm.WalletClient.Call("/wallet/createtransaction", params)
	if err != nil {
		return "", err
	}

	// // type Transaction_Contract struct {
	// // 	Type                 Transaction_Contract_ContractType `protobuf:"varint,1,opt,name=type,proto3,enum=protocol.Transaction_Contract_ContractType" json:"type,omitempty"`
	// // 	Parameter            *any.Any                          `protobuf:"bytes,2,opt,name=parameter,proto3" json:"parameter,omitempty"`
	// // 	Provider             []byte                            `protobuf:"bytes,3,opt,name=provider,proto3" json:"provider,omitempty"`
	// // 	ContractName         []byte                            `protobuf:"bytes,4,opt,name=ContractName,proto3" json:"ContractName,omitempty"`
	// // 	XXX_NoUnkeyedLiteral struct{}                          `json:"-"`
	// // 	XXX_unrecognized     []byte                            `json:"-"`
	// // 	XXX_sizecache        int32                             `json:"-"`
	// // }
	// tx := &core.Transaction_Contract{}
	// if err := gjson.Unmarshal([]byte(r.Raw), tx); err != nil {
	// 	log.Errorf("Proto Unmarshal: ", err)
	// 	return "", err
	// }
	raw = hex.EncodeToString([]byte(r.Raw))

	return raw, nil
}

// Writing! No used!
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

	r, err := wm.WalletClient.Call("/wallet/gettransactionsign", params)
	if err != nil {
		return nil, err
	}
	fmt.Println("Test = ", r)

	return rawSinged, nil
}

// BroadcastTransaction Done!
// Function：Broadcast the signed transaction
// 	demo：curl -X POST http://127.0.0.1:8090/wallet/broadcasttransaction -d ‘
// 		{“signature”:[“97c825b41c77de2a8bd65b3df55cd4c0df59c307c0187e42321dcc1cc455ddba583dd9502e17cfec5945b34cad0511985a6165999092a6dec84c2bdd97e649fc01”],
// 		 ”txID”:”454f156bf1256587ff6ccdbc56e64ad0c51e4f8efea5490dcbc720ee606bc7b8”,
// 		 ”raw_data”:{“contract”:[{
// 			 				“parameter”:{
// 								“value”:{“amount”:1000,
// 									   ”owner_address”:”41e552f6487585c2b58bc2c9bb4492bc1f17132cd0”,
// 									   ”to_address”:   ”41d1e7a6bc354106cb410e65ff8b181c600ff14292”},
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
func (wm *WalletManager) BroadcastTransaction1(raw string) error {

	tx := &core.Transaction{}
	if txBytes, err := hex.DecodeString(raw); err != nil {
		log.Errorf("Hex decode error: %+v", err)
		return err
	} else {
		if err := proto.Unmarshal(txBytes, tx); err != nil {
			log.Errorf("Hex decode error: %+v", err)
			return err
		}
	}

	/* Generate Params */

	var (
		signature []string
		txID      string
		contracts []map[string]interface{}
		raw_data  map[string]interface{}
	)

	for _, x := range tx.GetSignature() {
		signature = append(signature, hex.EncodeToString(x)) // base64
	}

	if txHash, err := getTxHash(tx); err != nil {
		log.Error(err)
		return err
	} else {
		txID = hex.EncodeToString(txHash)
	}

	rawData := tx.GetRawData()

	contracts = []map[string]interface{}{}
	for _, c := range rawData.GetContract() {
		any := c.GetParameter().GetValue()

		tc := &core.TransferContract{}
		if err := proto.Unmarshal(any, tc); err != nil {
			return err
		}

		contract := map[string]interface{}{
			"type": c.GetType().String(),
			"parameter": map[string]interface{}{
				"type_url": c.GetParameter().GetTypeUrl(),
				"value": map[string]interface{}{
					"amount":        tc.Amount,
					"owner_address": hex.EncodeToString(tc.GetOwnerAddress()),
					"to_address":    hex.EncodeToString(tc.GetToAddress()),
				},
			},
		}
		contracts = append(contracts, contract)
	}
	raw_data = map[string]interface{}{
		"ref_block_bytes": hex.EncodeToString(rawData.GetRefBlockBytes()),
		"ref_block_hash":  hex.EncodeToString(rawData.GetRefBlockHash()),
		"expiration":      rawData.GetExpiration(),
		"timestamp":       rawData.GetTimestamp(),
		"contract":        contracts,
	}
	params := req.Param{
		"signature": signature,
		"txID":      txID,
		"raw_data":  raw_data,
	}

	// Call api
	r, err := wm.WalletClient.Call("/wallet/broadcasttransaction", params)
	if err != nil {
		log.Error(err)
		return err
	} else {
		log.Debugf("Test = %+v\n", r)

		if r.Get("result").Bool() != true {

			var err error

			if r.Get("message").String() != "" {
				msg, _ := hex.DecodeString(r.Get("message").String())
				err = fmt.Errorf("BroadcastTransaction error message: %+v", string(msg))
			} else {
				err = fmt.Errorf("BroadcastTransaction return error: %+v", r)
			}
			log.Error(err)

			return err
		}
	}

	return nil
}
