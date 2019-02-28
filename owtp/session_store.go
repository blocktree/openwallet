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
	"encoding/json"
	"fmt"
	"github.com/blocktree/OpenWallet/session"
	"time"
)

// SessionManager contains Provider and its configuration.
type SessionManager struct {
	provider session.Provider
	config   *session.ManagerConfig
}

// NewManager Create new Manager with provider name and json config string.
// provider name:
// 1. cookie
// 2. file
// 3. memory
// 4. redis
// 5. mysql
// json config:
// 1. is https  default false
// 2. hashfunc  default sha1
// 3. hashkey default beegosessionkey
// 4. maxage default is none
func NewSessionManager(provideName string, cf *session.ManagerConfig) (*SessionManager, error) {
	provider, error := session.GetProvider(provideName)
	if error != nil {
		return nil, error
	}

	if cf.Maxlifetime == 0 {
		cf.Maxlifetime = cf.Gclifetime
	}

	//if cf.EnableSidInHTTPHeader {
	//	if cf.SessionNameInHTTPHeader == "" {
	//		panic(errors.New("SessionNameInHTTPHeader is empty"))
	//	}
	//
	//	strMimeHeader := textproto.CanonicalMIMEHeaderKey(cf.SessionNameInHTTPHeader)
	//	if cf.SessionNameInHTTPHeader != strMimeHeader {
	//		strErrMsg := "SessionNameInHTTPHeader (" + cf.SessionNameInHTTPHeader + ") has the wrong format, it should be like this : " + strMimeHeader
	//		panic(errors.New(strErrMsg))
	//	}
	//}

	err := provider.SessionInit(cf.Maxlifetime, cf.ProviderConfig)
	if err != nil {
		return nil, err
	}

	//if cf.SessionIDLength == 0 {
	//	cf.SessionIDLength = 16
	//}

	return &SessionManager{
		provider,
		cf,
	}, nil
}

// GetProvider return current manager's provider
func (store *SessionManager) GetProvider() session.Provider {
	return store.provider
}

// SaveAddr 保存节点
func (store *SessionManager) SavePeer(peer Peer) {

	config := peer.ConnectConfig()

	b, err := json.Marshal(config)
	if err == nil {
		store.Put(peer.PID(), peer.PID(), string(b))
	}
}

//PeerInfo 节点信息
func (store *SessionManager) PeerInfo(id string) PeerInfo {

	var config ConnectConfig
	b, ok := store.Get(id, id).(string)
	if ok {
		json.Unmarshal([]byte(b), &config)
	}
	return PeerInfo{
		ID:     id,
		Config: config,
	}
}

// Get 获取节点属性
func (store *SessionManager) Get(id string, key string) interface{} {
	session, err := store.GetSessionStore(id)
	if err != nil {
		return nil
	}
	return session.Get(key)
}

// GetString
func (store *SessionManager) GetString(id string, key string) string {
	val := store.Get(id, key)
	if val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}

	return ""
}

// Put 设置节点属性
func (store *SessionManager) Put(id string, key string, val interface{}) error {
	session, err := store.GetSessionStore(id)
	if err != nil {
		return err
	}
	return session.Set(key, val)
}

//Delete
func (store *SessionManager) Delete(id string, key string) error {
	session, err := store.GetSessionStore(id)
	if err != nil {
		return err
	}
	return session.Delete(key)
}

//Destroy
func (store *SessionManager) Destroy(id string) error {
	session, err := store.GetSessionStore(id)
	if err != nil {
		return err
	}
	session.Flush()
	store.SessionDestroy(id)
	return nil
}

// SessionDestroy Destroy session by its id in http request cookie.
func (store *SessionManager) SessionDestroy(pid string) {
	sid, _ := store.sessionID(pid)
	store.provider.SessionDestroy(sid)
}

// GetSessionStore Get SessionStore by its id.
func (store *SessionManager) GetSessionStore(pid string) (sessions session.Store, err error) {

	if store == nil {
		return nil, fmt.Errorf("SessionManager is empty. ")
	}

	if pid == "" {
		return nil, fmt.Errorf("pid is empty. ")
	}

	// Generate a new session
	sid, errs := store.sessionID(pid)
	if errs != nil {
		return nil, errs
	}

	sessions, err = store.provider.SessionRead(sid)
	if err != nil {
		return nil, err
	}

	return sessions, err
}

// GC Start session gc process.
// it can do gc in times after gc lifetime.
func (store *SessionManager) GC() {
	store.provider.SessionGC()
	time.AfterFunc(time.Duration(store.config.Gclifetime)*time.Second, func() { store.GC() })
}

// GetActiveSession Get all active sessions count number.
func (store *SessionManager) GetActiveSession() int {
	return store.provider.SessionAll()
}

func (store *SessionManager) sessionID(pid string) (string, error) {
	return store.config.SessionIDPrefix + pid, nil
}
