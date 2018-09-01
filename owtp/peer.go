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

package owtp

import (
	"fmt"
	"net"
	"sync"
	"github.com/tidwall/gjson"
)

//DataPacket 数据包
type DataPacket struct {
	/*

		本协议传输数据，格式编码采用json。消息接收与发送，都遵循数据包规范定义字段内容。

		| 参数名 | 类型   | 示例             | 描述                                                                              |
		|--------|--------|------------------|-----------------------------------------------------------------------------------|
		| r      | uint8  | 1                | 传输类型，1：请求，2：响应                                                            |
		| m      | string | subscribe        | 方法名，对应接口方法定义                                                           |
		| n      | uint32 | 123              | 请求序号。为了保证请求对应响应按序执行，并防御重放攻击，序号可以为随机数，但不可重复。 |
		| t      | uint32 | 1528520843       | 时间戳。限制请求在特定时间范围内有效，如10分钟。                                     |
		| d      | Object | {"foo": "hello"} | 数据主体，请求内容或响应内容。接口方法说明中，主要说明这部分。                        |
		| s      | string | Qwse==           | 合并[r+m+n+t+d]进行sha256两次签名并base64编码，用于校验数据的一致性和合法性       |

	*/

	Req       uint64      `json:"r"`
	Method    string      `json:"m"`
	Nonce     uint64      `json:"n" storm:"id"`
	Timestamp int64       `json:"t"`
	Data      interface{} `json:"d"`
	Signature string      `json:"s"`
}

//NewDataPacket 通过 gjson转为DataPacket
func NewDataPacket(json gjson.Result) *DataPacket {
	dp := &DataPacket{}
	dp.Req = json.Get("r").Uint()
	dp.Method = json.Get("m").String()
	dp.Nonce = json.Get("n").Uint()
	dp.Timestamp = json.Get("t").Int()
	dp.Data = json.Get("d").Value()
	dp.Signature = json.Get("s").String()
	return dp
}

type PeerInfo struct {
	ID     string
	Config interface{}
}

type PeerAttribute map[string]interface{}

// Peer 节点
type Peer interface {
	Auth() Authorization
	PID() string                          //节点ID
	OpenPipe() error                      //OpenPipe 打开通道
	Send(data DataPacket) error           //发送请求
	IsHost() bool                         //是否主机，我方主动连接的节点
	IsConnected() bool                    //是否已经连接
	SetHandler(handler PeerHandler) error //设置节点的服务者
	Close() error                         //关闭节点

	LocalAddr() net.Addr  //本地节点地址
	RemoteAddr() net.Addr //远程节点地址

	GetConfig() interface{} // 返回配置信息
}

// PeerHandler 节点监听器
type PeerHandler interface {
	OnPeerOpen(peer Peer)                                      //节点连接成功
	OnPeerClose(peer Peer, reason string)                      //节点关闭
	OnPeerNewDataPacketReceived(peer Peer, packet *DataPacket) //节点获取新数据包
}

// Peerstore 节点存储器
type Peerstore interface {
	// SaveAddr 保存节点地址
	SavePeer(id string, peer Peer)

	// GetAddr 获取节点地址
	GetPeer(id string) Peer

	// DeleteAddr 删除节点的地址
	DeletePeer(id string)

	//PeerInfo 节点信息
	PeerInfo(id string) PeerInfo

	// Get 获取节点属性
	Get(id string, key string) (interface{}, error)

	// Put 设置节点属性
	Put(id string, key string, val interface{}) error

	//// Peers 节点列表
	//OnlinePeers() []Peer
	//
	//// GetOnlinePeer 获取当前在线的Peer
	//GetOnlinePeer(id string) Peer
	//
	//// AddOnlinePeer 添加在线节点
	//AddOnlinePeer(peer Peer)
	//
	////RemoveOfflinePeer 移除不在线的节点
	//RemoveOfflinePeer(id string)
}

type owtpPeerstore struct {
	onlinePeers map[string]Peer
	peers       map[string]Peer
	peerInfos   map[string]PeerAttribute
	mu          sync.RWMutex
}

// NewPeerstore 创建支持OWTP协议的Peerstore
func NewPeerstore() *owtpPeerstore {
	store := owtpPeerstore{
		onlinePeers: make(map[string]Peer),
		peers:       make(map[string]Peer),
		peerInfos:   make(map[string]PeerAttribute),
	}
	return &store
}

// SaveAddr 保存节点
func (store *owtpPeerstore) SavePeer(id string, peer Peer) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if store.peers == nil {
		store.peers = make(map[string]Peer)
	}
	store.peers[id] = peer
}

// GetAddr 获取节点
func (store *owtpPeerstore) GetPeer(id string) Peer {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.peers == nil {
		return nil
	}
	return store.peers[id]
}

// DeletePeer 删除节点的地址
func (store *owtpPeerstore) DeletePeer(id string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.peers == nil {
		return
	}
	delete(store.peers, id)
}

//PeerInfo 节点信息
func (store *owtpPeerstore) PeerInfo(id string) PeerInfo {
	if store.peers == nil {
		return PeerInfo{}
	}
	peer := store.peers[id]
	return PeerInfo{
		ID:     id,
		Config: peer.GetConfig(),
	}
}

// Get 获取节点属性
func (store *owtpPeerstore) Get(id string, key string) (interface{}, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.peerInfos == nil {
		return nil, fmt.Errorf("no peer for this peer.id")
	}
	peerAttribute := store.peerInfos[id]
	if peerAttribute == nil {
		return nil, fmt.Errorf("no peer for this peer.id")
	}
	return peerAttribute[key], nil
}

// Put 设置节点属性
func (store *owtpPeerstore) Put(id string, key string, val interface{}) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.peerInfos == nil {
		store.peerInfos = make(map[string]PeerAttribute)
	}
	peerAttribute := store.peerInfos[id]
	if peerAttribute == nil {
		peerAttribute = make(map[string]interface{})
		store.peerInfos[id] = peerAttribute
	}

	peerAttribute[key] = val
	return nil
}

// Peers 节点列表
func (store *owtpPeerstore) OnlinePeers() []Peer {
	store.mu.RLock()
	defer store.mu.RUnlock()
	peers := make([]Peer, 0)
	for _, peer := range store.onlinePeers {
		peers = append(peers, peer)
	}
	return peers
}

// GetOnlinePeer 获取当前在线的Peer
func (store *owtpPeerstore) GetOnlinePeer(id string) Peer {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.onlinePeers == nil {
		return nil
	}
	return store.onlinePeers[id]
}

// AddOnlinePeer 添加在线节点
func (store *owtpPeerstore) AddOnlinePeer(peer Peer) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.onlinePeers == nil {
		store.onlinePeers = make(map[string]Peer)
	}
	store.onlinePeers[peer.PID()] = peer
}

//RemoveOfflinePeer 移除不在线的节点
func (store *owtpPeerstore) RemoveOfflinePeer(id string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.onlinePeers == nil {
		return
	}
	delete(store.onlinePeers, id)
}
