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

package owtp

import (
	"encoding/json"
	"github.com/blocktree/openwallet/log"
	"sync"
)

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

	// GetString
	GetString(id string, key string) string

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
		peerInfos: make(map[string]PeerAttribute),
	}
	return &store
}

// SaveAddr 保存节点
func (store *owtpPeerstore) SavePeer(peer Peer) {

	config := peer.ConnectConfig()

	b, err := json.Marshal(config)
	if err == nil {
		store.Put(peer.PID(), peer.PID(), string(b))
	}
}

//PeerInfo 节点信息
func (store *owtpPeerstore) PeerInfo(id string) PeerInfo {

	var config ConnectConfig
	b, ok := store.Get(id, id).(string)
	if ok {
		err := json.Unmarshal([]byte(b), &config)
		if err != nil {
			log.Errorf("json.Unmarshal PeerInfo failed")
		}
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

// GetString
func (store *owtpPeerstore) GetString(id string, key string) string {
	val := store.Get(id, key)
	if val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}

	return ""
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
