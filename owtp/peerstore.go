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

import "sync"

// Peerstore 节点存储器
type Peerstore interface {
	// SaveAddr 保存节点地址
	SavePeer(peer Peer)
	//
	//// GetAddr 获取节点地址
	//GetPeer(id string) Peer
	//
	//// DeleteAddr 删除节点的地址
	//DeletePeer(id string)

	//PeerInfo 节点信息
	PeerInfo(id string) PeerInfo

	// Get 获取节点属性
	Get(id string, key string) interface{}

	// Put 设置节点属性
	Put(id string, key string, val interface{}) error

	// Delete 设置节点属性
	Delete(id string, key string) error

	//Destroy 清空store数据
	Destroy(id string) error
}

type owtpPeerstore struct {
	//onlinePeers map[string]Peer
	//peers       map[string]Peer
	peerInfos map[string]PeerAttribute
	mu        sync.RWMutex
}

// NewPeerstore 创建支持OWTP协议的Peerstore
func NewOWTPPeerstore() *owtpPeerstore {
	store := owtpPeerstore{
		//onlinePeers: make(map[string]Peer),
		//peers:       make(map[string]Peer),
		peerInfos: make(map[string]PeerAttribute),
	}
	return &store
}

// SaveAddr 保存节点
func (store *owtpPeerstore) SavePeer(peer Peer) {

	config := peer.GetConfig()
	////链接类型
	//connectType := config["connectType"]
	//
	////HTTP类型，不做保存节点，因为都是短连接
	//if connectType == HTTP {
	//	if !peer.Auth().EnableKeyAgreement() {
	//		return
	//	}
	//	if !peer.Auth().EnableAuth() {
	//		return
	//	}
	//}

	if config != nil && len(config) > 0 {
		store.Put(peer.PID(), peer.PID(), config)
	}

	//store.mu.Lock()
	//defer store.mu.Unlock()
	//
	//if store.peers == nil {
	//	store.peers = make(map[string]Peer)
	//}
	//store.peers[id] = peer
}

//GetAddr 获取节点
//func (store *owtpPeerstore) GetPeer(id string) Peer {
//	store.mu.RLock()
//	defer store.mu.RUnlock()
//	if store.peers == nil {
//		return nil
//	}
//	return store.peers[id]
//}

//DeletePeer 删除节点的地址
//func (store *owtpPeerstore) DeletePeer(id string) {
//	store.mu.Lock()
//	defer store.mu.Unlock()
//	if store.peers == nil {
//		return
//	}
//	delete(store.peers, id)
//}

//PeerInfo 节点信息
func (store *owtpPeerstore) PeerInfo(id string) PeerInfo {
	//if store.peers == nil {
	//	return PeerInfo{}
	//}
	//peer := store.peers[id]

	config, ok := store.Get(id, id).(map[string]string)
	if !ok {
		config = make(map[string]string)
	}
	return PeerInfo{
		ID:     id,
		Config: config,
	}
}

// Get 获取节点属性
func (store *owtpPeerstore) Get(id string, key string) interface{} {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.peerInfos == nil {
		return nil
	}
	peerAttribute := store.peerInfos[id]
	if peerAttribute == nil {
		return nil
	}
	return peerAttribute[key]
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
	store.peerInfos[id] = peerAttribute
	return nil
}

//Delete
func (store *owtpPeerstore) Delete(id string, key string) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.peerInfos == nil {
		return nil
	}
	peerAttribute := store.peerInfos[id]
	if peerAttribute == nil {
		return nil
	}
	delete(peerAttribute, key)
	store.peerInfos[id] = peerAttribute
	return nil
}

//Destroy
func (store *owtpPeerstore) Destroy(id string) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	peerAttribute := store.peerInfos[id]
	if peerAttribute == nil {
		return nil
	}
	peerAttribute = make(PeerAttribute)
	store.peerInfos[id] = peerAttribute
	return nil
}
