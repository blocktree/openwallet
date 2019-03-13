/*
* Copyright 2018 The openwallet Authors
* This file is part of the openwallet library.
*
* The openwallet library is free software: you can redistribute it and/or modify
* it under the terms of the GNU Lesser General Public License as published by
* the Free Software Foundation, either version 3 of the License, or
* (at your option) any later version.
*
* The openwallet library is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
* GNU Lesser General Public License for more details.
 */

package tron

// ListWitnesses Writing!
// Function：
// 	Query the list of Super Representatives
// demo:
// 	curl -X POSThttp://127.0.0.1:8090/wallet/listwitnesses
// Parameters：None
// Return value：List of all Super Representatives
func (wm *WalletManager) ListWitnesses() (witnesses string, err error) {

	r, err := wm.WalletClient.Call("/wallet/listwitnesses", nil)
	if err != nil {
		return "", err
	}
	_ = r

	// // type WitnessList struct {
	// // 	Witnesses            []*core.Witness `protobuf:"bytes,1,rep,name=witnesses,proto3" json:"witnesses,omitempty"`
	// // 	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	// // 	XXX_unrecognized     []byte          `json:"-"`
	// // 	XXX_sizecache        int32           `json:"-"`
	// // }
	// // type Witness struct {
	// // 	Address              []byte   `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// // 	VoteCount            int64    `protobuf:"varint,2,opt,name=voteCount,proto3" json:"voteCount,omitempty"`
	// // 	PubKey               []byte   `protobuf:"bytes,3,opt,name=pubKey,proto3" json:"pubKey,omitempty"`
	// // 	Url                  string   `protobuf:"bytes,4,opt,name=url,proto3" json:"url,omitempty"`
	// // 	TotalProduced        int64    `protobuf:"varint,5,opt,name=totalProduced,proto3" json:"totalProduced,omitempty"`
	// // 	TotalMissed          int64    `protobuf:"varint,6,opt,name=totalMissed,proto3" json:"totalMissed,omitempty"`
	// // 	LatestBlockNum       int64    `protobuf:"varint,7,opt,name=latestBlockNum,proto3" json:"latestBlockNum,omitempty"`
	// // 	LatestSlotNum        int64    `protobuf:"varint,8,opt,name=latestSlotNum,proto3" json:"latestSlotNum,omitempty"`
	// // 	IsJobs               bool     `protobuf:"varint,9,opt,name=isJobs,proto3" json:"isJobs,omitempty"`
	// // 	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	// // 	XXX_unrecognized     []byte   `json:"-"`
	// // 	XXX_sizecache        int32    `json:"-"`
	// // }
	// witnesses = &api.WitnessList{}
	// if err := gjson.Unmarshal([]byte(r.Raw), witnesses); err != nil {
	// 	return nil, err
	// }

	return witnesses, nil
}

// ListNodes Writing!
// Function：
// 	List the nodes which the api fullnode is connecting on the network
// demo:
// 	curl -X POST http://127.0.0.1:8090/wallet/listnodes
// Parameters：None
// Return value：List of nodes
func (wm *WalletManager) ListNodes() (nodes []string, err error) {

	r, err := wm.WalletClient.Call("/wallet/listnodes", nil)
	if err != nil {
		return nil, err
	}
	_ = r

	// // Gossip node list
	// // type NodeList struct {
	// // 	Nodes                []*Node  `protobuf:"bytes,1,rep,name=nodes,proto3" json:"nodes,omitempty"`
	// // 	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	// // 	XXX_unrecognized     []byte   `json:"-"`
	// // 	XXX_sizecache        int32    `json:"-"`
	// // }
	// // Gossip node
	// // type Node struct {
	// // 	Address              *Address `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// // 	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	// // 	XXX_unrecognized     []byte   `json:"-"`
	// // 	XXX_sizecache        int32    `json:"-"`
	// // }
	// // type Address struct {
	// // 	Host                 []byte   `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	// // 	Port                 int32    `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	// // 	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	// // 	XXX_unrecognized     []byte   `json:"-"`
	// // 	XXX_sizecache        int32    `json:"-"`
	// // }
	// nodes = &api.NodeList{}
	// if err := gjson.Unmarshal([]byte(r.Raw), nodes); err != nil {
	// 	return nil, err
	// }

	return nodes, nil
}
