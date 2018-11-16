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

package bitcoin

import "github.com/tidwall/gjson"

type OmniTransaction struct {

	/*
		{
			"txid": "e1afc9cf4a07bb1566bdb87363e6009c7c6858ea6eab4ac730babec92f5ba712",
			"fee": "0.00000257",
			"sendingaddress": "mgiD6ySS59rjWWmKXbV7Tf2WDcQvfWA89j",
			"referenceaddress": "mmRthvVHit7rPvi1AL852Kct7K3C8vQjVz",
			"ismine": false,
			"version": 0,
			"type_int": 0,
			"type": "Simple Send",
			"propertyid": 2,
			"divisible": true,
			"amount": "0.10000000",
			"valid": true,
			"blockhash": "00000000000001491084285ed904135e7290a548373460e77a6894e07e5dfb45",
			"blocktime": 1542174369,
			"positioninblock": 11,
			"block": 1443395,
			"confirmations": 1
		}
	*/

	TxID             string `json:"txid"`
	Fees             string `json:"fee"`
	SendingAddress   string `json:"sendingaddress"`
	ReferenceAddress string `json:"referenceaddress"`
	IsMine           string `json:"ismine"`
	Version          uint64 `json:"version"`
	TypeInt          uint64  `json:"type_int"`
	TypeStr          string `json:"type"`
	PropertyId       uint64 `json:"propertyid"`
	Divisible        bool   `json:"divisible"`
	Amount           string `json:"amount"`
	Valid            bool   `json:"valid"`
	BlockTime        int64  `json:"blocktime"`
	PositionInBlock  uint64 `json:"positioninblock"`
	BlockHash        string `json:"blockhash"`
	Block            uint64 `json:"block"`
	Confirmations    uint64 `json:"confirmations"`
}

func NewOmniTx(json *gjson.Result) *OmniTransaction {
	obj := &OmniTransaction{}
	obj.TxID = gjson.Get(json.Raw, "txid").String()
	obj.Fees = gjson.Get(json.Raw, "fee").String()
	obj.SendingAddress = gjson.Get(json.Raw, "sendingaddress").String()
	obj.ReferenceAddress = gjson.Get(json.Raw, "referenceaddress").String()
	obj.IsMine = gjson.Get(json.Raw, "ismine").String()
	obj.Version = gjson.Get(json.Raw, "version").Uint()
	obj.TypeInt = gjson.Get(json.Raw, "type_int").Uint()
	obj.TypeStr = gjson.Get(json.Raw, "type").String()
	obj.PropertyId = gjson.Get(json.Raw, "propertyid").Uint()
	obj.Divisible = gjson.Get(json.Raw, "divisible").Bool()
	obj.Amount = gjson.Get(json.Raw, "amount").String()
	obj.Valid = gjson.Get(json.Raw, "valid").Bool()
	obj.BlockTime = gjson.Get(json.Raw, "blocktime").Int()
	obj.PositionInBlock = gjson.Get(json.Raw, "positioninblock").Uint()
	obj.BlockHash = gjson.Get(json.Raw, "blockhash").String()
	obj.Block = gjson.Get(json.Raw, "block").Uint()
	obj.Confirmations = gjson.Get(json.Raw, "confirmations").Uint()
	return obj
}
