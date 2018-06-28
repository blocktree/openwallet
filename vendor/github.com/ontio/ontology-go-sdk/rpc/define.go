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

package rpc

import (
	"encoding/json"
)

const (
	RPC_GET_VERSION                 = "getversion"
	RPC_GET_TRANSACTION             = "getrawtransaction"
	RPC_SEND_TRANSACTION            = "sendrawtransaction"
	RPC_GET_BLOCK                   = "getblock"
	RPC_GET_BLOCK_COUNT             = "getblockcount"
	RPC_GET_BLOCK_HASH              = "getblockhash"
	RPC_GET_CURRENT_BLOCK_HASH      = "getbestblockhash"
	RPC_GET_ONT_BALANCE             = "getbalance"
	RPC_GET_SMART_CONTRACT_EVENT    = "getsmartcodeevent"
	RPC_GET_STORAGE                 = "getstorage"
	RPC_GET_SMART_CONTRACT          = "getcontractstate"
	RPC_GET_GENERATE_BLOCK_TIME     = "getgenerateblocktime"
	RPC_GET_MERKLE_PROOF            = "getmerkleproof"
	SEND_EMERGENCY_GOV_REQ          = "sendemergencygovreq"
	GET_BLOCK_ROOT_WITH_NEW_TX_ROOT = "getblockrootwithnewtxroot"
)

//JsonRpc version
const JSON_RPC_VERSION = "2.0"

//JsonRpcRequest object in rpc
type JsonRpcRequest struct {
	Version string        `json:"jsonrpc"`
	Id      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

//JsonRpcResponse object response for JsonRpcRequest
type JsonRpcResponse struct {
	Id     string          `json:"id"`
	Error  int64           `json:"error"`
	Desc   string          `json:"desc"`
	Result json.RawMessage `json:"result"`
}
