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

package bopo

import (
	"fmt"

	"github.com/blocktree/OpenWallet/crypto"
	"github.com/bytom/common"
)

//Wallet 钱包模型
type Wallet struct {
	WalletID string `json:"rootid"`
	Addr     string `json:"addr"`
	Balance  string `json:"balance"`
	Alias    string `json:"alias"`
	Password string `json:"password"`
	RootPub  string `json:"rootpub"`
	KeyFile  string
}

// ------------------------------------------------------------------
//BlockchainInfo 本地节点区块链信息
type BlockchainInfo struct {
	Chain                string `json:"chain"`
	Blocks               uint64 `json:"blocks"`
	Headers              uint64 `json:"headers"`
	Bestblockhash        string `json:"bestblockhash"`
	Difficulty           string `json:"difficulty"`
	Mediantime           uint64 `json:"mediantime"`
	Verificationprogress string `json:"verificationprogress"`
	Chainwork            string `json:"chainwork"`
	Pruned               bool   `json:"pruned"`
}

// Block Fabric区块结构
type Block struct {
	/*
		{
		'consensusMetadata': 'CAI=',
		'nonHashData': {'chaincodeEvents': [{}],
		                'localLedgerCommitTimestamp': {'nanos': 94090909,
		                                               'seconds': 1518419369}},
		'previousBlockHash': '+SuQi13GO8bGYik2SQOIeRcRCiH21NUVH7uSRtMTtq/td0EaKTw/jTQdPvICWq2Gn+klrJ2Hs0DqRYukPVUyow==',
		'stateHash': 'ZUVvUvCScFUAWRk3E3gwIWnsOaFNDUXT2C4x6Wl3GTMsnWTk1gaFny4G582NTpzqWGR755zCQ5kDA8YyMl3R2w==',
		'transactions': [{'cert': 'MIICtDCCAlugAwIBAgIRAPn/PTSQLUT/gj6C34hBopQwCgYIKoZIzj0EAwMwMzELMAkGA1UEBhMCQ04xFjAUBgNVBAoTDUdhbWVwYWkgRnVuZC4xDDAKBgNVBAMTA3RjYTAeFw0xODAyMTIwNzA2MDhaFw0xODA1MTMwNzA2MDhaMEcxCzAJBgNVBAYTAkNOMRYwFAYDVQQKEw1HYW1lcGFpIEZ1bmQuMSAwHgYDVQQDExdUcmFuc2FjdGlvbiBDZXJ0aWZpY2F0ZTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABPTKMJHudYvdS4B08CEwWTEF8Ll4SUjq93mq4x+8R2bMrdXGHwC6h85r4cwwcyNLJjEYYgqa9VanrQ/Kn9usCpijggE6MIIBNjAOBgNVHQ8BAf8EBAMCB4AwDAYDVR0TAQH/BAIwADANBgNVHQ4EBgQEAQIDBDAPBgNVHSMECDAGgAQBAgMEMBIGBioDBAUGCgQIRGVsZWdhdGUwEwYGKgMEBQYLBAlBbGlFQ1NIQjEwTQYGKgMEBQYHAQH/BEBduQTYRCWuHIAIZoMgUMl90wPUkGP5bSF4MQTz/XNxsaYZ/Gsi5b1P2zOHJJNA8IQzw/TKwF11yXToaKm/oOImMEoGBioDBAUGCARA4nZ1FCdDLT8w2vnhPClfkIk13XCFBir60qDE3Ic//AMoHt8IFwmndtOaDIzPrPCKajDlt7OsTertkTO9dLQPBjAyBgYqAwQFBgkEKDAwSEVBRFBhaUFkbWluUm9sZS0+MSNQYWlBZG1pblJlZ2lvbi0+MiMwCgYIKoZIzj0EAwMDRwAwRAIgCkc8m2RBcl9dYbmyrdyceb5g51yGuRySD9cM+4Cc5s0CIEdR6HhgnM8RODTFUo4GGjY6/xk285MihBgL83Mtv1Lx',
		                  'chaincodeID': 'Eg9nYW1lcGFpY29yZV92MDE=',
		                  'nonce': '67ba1CKLYHSfQQ0KASPt4GV5j6YtmRQJ',
		                  'payload': 'CtsCCAESERIPZ2FtZXBhaWNvcmVfdjAxGqUCChFVU0VSX1JFR1BVQkxJQ0tFWQpIQ2lJMVdFWlZOVjlHZWtSVGFFRnplR0Z6TVZSNlVXMW5NV3hvVVROeVJUUlFXVnBuRWhEUkEwL3VWaUE2K0gxdXErZDZOblQ5CmRDa2dJQVJKRUNpQlZnR3o3d2NqTHNIbHpSdGhRREVOcWo0NUtsQUJBTTJSTHA4bm4raHUrN2hJZzB5U0pYVDdtdTJ4elFQa0wrQkwwN3pjTDdEQ2VQWDlpUjFVVDR0eStLZnc9CmBDa1FLSUJobGF3bzlnSWlFUXpzYXVjcXNYK3ZyM2ptdUl2VkpPcVgzcEg4TE5tUDVFaUNnUVBDaVNDeHpRdVRKNnJod3dqazRYNlFqWHFXQ1F6YWdydktzTk96bFJnPT1CDFBhaUFkbWluUm9sZUIOUGFpQWRtaW5SZWdpb24=',
		                  'signature': 'MEQCIHvGzTPfWbNLL0Ad+AbNY9/WTsgDNxzHeaUWw0BPFx8NAiB+QBfB6HXOQ+pJyX1bDpbJXX+qV2KZ1c1cIuYl+W99Ug==',
		                  'timestamp': {'nanos': 67157084,
		                                'seconds': 1518419357},
		                  'txid': 'effe5d08-346b-4689-b105-dfdcbdfe790f',
		                  'type': 2}]
		}
	*/
	Hash              string
	Previousblockhash string     `json:"previousBlockHash,omitempty"`
	Statehash         string     `json:"stateHash,omitempty"`
	Transactions      []*BlockTX `json:"transactions,omitempty"`
	Height            uint64     `storm:"id"`
}

// BlockTX 区块交易记录 Struct
type BlockTX struct {
	Cert        string `json:"cert,omitempty"`
	ChaincodeID string `json:"chaincodeID,omitempty"`
	Nonce       string `json:"nonce,omitempty"`
	Payload     []byte `json:"payload"`
	Signature   string `json:"signature,omitempty"`
	Timestrap   struct {
		nanos   uint64
		seconds uint64
	}
	Txid string `json:"txid,omitempty"`
	Type string `json:"type,omitempty"`
}

// PayloadSpec 每笔交易的具体数据
type BlockTxPayload struct {
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
	Comment string `json:"comment,omitempty"`
	Amount  uint64 `json:"amount,omitempty"`
}

// ------------------------------------------------------------------
//UnscanRecords 扫描失败的区块及交易
type UnscanRecord struct {
	ID          string `storm:"id"` // primary key
	BlockHeight uint64
	TxID        string
	Reason      string
	RescanCount int
}

func NewUnscanRecord(height uint64, txID, reason string) *UnscanRecord {
	obj := UnscanRecord{}
	obj.BlockHeight = height
	obj.TxID = txID
	obj.Reason = reason
	obj.ID = common.Bytes2Hex(crypto.SHA256([]byte(fmt.Sprintf("%d_%s", height, txID))))
	return &obj
}

// 以 Fabric 原生结构解析区块数据
// type BlockTXPayload struct {
// 	/*
// 		   chaincodeSpec:< type:GOLANG
// 		                   chaincodeID:<name:"gamepaicore_v01" >
// 		                   ctorMsg:<args:"USER_FUND"
// 		                            args:"CiI1WmFQWGZKYUxOckduWHV5WHVuRkU0eEt4YWtFemdUSVpREhxGcmkgQXVnIDE3IDExOjA5OjI1IENTVCAyMDE4IkgIARJECiDii7hQ6FHfEscWWJxkV3gcKZphiXnt3+2vbJ3hUqkL5hIg7t8wD3SjhMoFOiIb54UUP0aLpNwrEXiflLDgxdWjS9s="
// 		                            args:"CiYIARIiNVpGVlZQNDdSZjVqLWs3TG9pUmNOb3psYzhkeW5iUFluZw=="
// 		                            args:"CkQKIGnwE1MpqraCAvg2mxnxsnX5T+iNbC0THejYY1QXB/eXEiAD4FZQ25G3OjxRCPPJDDdcLQXA/180ykUHW4TB0Fs+Vg==" > // Signature
// 		                   attributes:"PaiAdminRole"
// 					 attributes:"PaiAdminRegion" >

// 		   ctorMsg_args := []slice{
// 		       Fund_Type
// 		       &pb.UserTxHeader{FundId: userid, Nounce: m.Nounce}
// 		       &pb.Fund{InvokeChaincode: uint32, D: {Pai: uint32, ToUserId: string}}
// 		       &pb.Signature{P: &pb.ECPoint{rx.Bytes(), ry.Bytes()}}
// 		   }
// 	*/
// 	Type        string
// 	ChaincodeID string
// 	Funddata    struct {
// 		TxHeader struct {
// 			FundID string
// 			Nounce string
// 		}
// 		Fund struct {
// 			InvokeChaincode uint32
// 			D_Pai           uint32
// 			D_ToUserId      string
// 		}
// 		Signature struct {
// 		}
// 	}
// 	attributes []string
// }
