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
	"sync"
)

type PeerInfo map[string]interface{}

// Peer 节点
type Peer interface {
	Auth() Authorization
	PID() string                          //节点ID
	OpenPipe() error                      //OpenPipe 打开通道
	Send(data DataPacket) error           //发送请求
	IsHost() bool						  //是否主机，我方主动连接的节点
	IsConnected() bool                    //是否已经连接
	SetHandler(handler PeerHandler) error //设置节点的服务者
	Close() error                         //关闭节点
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
	SaveAddr(id string, addr string)

	// GetAddr 获取节点地址
	GetAddr(id string) string

	// DeleteAddr 删除节点的地址
	DeleteAddr(id string)

	// Get 获取节点属性
	Get(id string, key string) (interface{}, error)

	// Put 设置节点属性
	Put(id string, key string, val interface{}) error

	// Peers 节点列表
	OnlinePeers() []Peer

	// GetOnlinePeer 获取当前在线的Peer
	GetOnlinePeer(id string) Peer

	// AddOnlinePeer 添加在线节点
	AddOnlinePeer(peer Peer)

	//RemoveOfflinePeer 移除不在线的节点
	RemoveOfflinePeer(id string)
}

type owtpPeerstore struct {
	onlinePeers map[string]Peer
	peerAddrs   map[string]string
	peerInfos   map[string]PeerInfo
	mu          sync.RWMutex
}

// NewPeerstore 创建支持OWTP协议的Peerstore
func NewPeerstore() *owtpPeerstore {
	store := owtpPeerstore{
		onlinePeers: make(map[string]Peer),
		peerAddrs:   make(map[string]string),
		peerInfos:   make(map[string]PeerInfo),
	}
	return &store
}

// SaveAddr 保存节点地址
func (store *owtpPeerstore) SaveAddr(id string, addr string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if store.peerAddrs == nil {
		store.peerAddrs = make(map[string]string)
	}
	store.peerAddrs[id] = addr
}

// GetAddr 获取节点地址
func (store *owtpPeerstore) GetAddr(id string) string {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.peerAddrs == nil {
		return ""
	}
	return store.peerAddrs[id]
}

// DeleteAddr 删除节点的地址
func (store *owtpPeerstore) DeleteAddr(id string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.peerAddrs == nil {
		return
	}
	delete(store.peerAddrs, id)
}

// Get 获取节点属性
func (store *owtpPeerstore) Get(id string, key string) (interface{}, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.peerInfos == nil {
		return nil, fmt.Errorf("no peer for this peer.id")
	}
	peerInfo := store.peerInfos[id]
	if peerInfo == nil {
		return nil, fmt.Errorf("no peer for this peer.id")
	}
	return peerInfo[key], nil
}

// Put 设置节点属性
func (store *owtpPeerstore) Put(id string, key string, val interface{}) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.peerInfos == nil {
		store.peerInfos = make(map[string]PeerInfo)
	}
	peerInfo := store.peerInfos[id]
	if peerInfo == nil {
		peerInfo = make(map[string]interface{})
		store.peerInfos[id] = peerInfo
	}

	peerInfo[key] = val
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

