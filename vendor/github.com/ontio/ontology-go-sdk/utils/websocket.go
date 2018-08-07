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

package utils

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

//WebSocketOptions provide some options for web socket client
type WebSocketOptions struct {
	HeartbeatInterval time.Duration //The interval of sending heartbeat
	HeartbeatPkg      []byte        //HeartbeatPkg is data package of heartbeat
}

//WebSocketClient use for client to operation web socket
type WebSocketClient struct {
	addr              string
	opts              *WebSocketOptions
	conn              *websocket.Conn
	existCh           chan interface{}
	OnConnect         func()
	OnClose           func()
	OnError           func(error)
	OnMessage         func([]byte)
	lastHeartbeatTime time.Time
	lock              sync.RWMutex
	status            bool
}

//Create WebSocketClient instance
func NewWebSocketClient(addr string, opts ...*WebSocketOptions) *WebSocketClient {
	var options *WebSocketOptions
	if len(opts) == 0 {
		options = &WebSocketOptions{
			HeartbeatInterval: 30 * time.Second,
			HeartbeatPkg:      []byte(`{"Action":"heartbeat"}`),
		}
	} else {
		options = opts[0]
	}
	return &WebSocketClient{
		addr:              addr,
		opts:              options,
		OnConnect:         func() {},
		OnClose:           func() {},
		OnError:           func(error) {},
		OnMessage:         func([]byte) {},
		existCh:           make(chan interface{}, 0),
		lastHeartbeatTime: time.Now(),
	}
}

//Connect to server
func (this *WebSocketClient) Connect() (err error) {
	this.conn, _, err = websocket.DefaultDialer.Dial(this.addr, nil)
	if err != nil {
		return err
	}
	this.OnConnect()
	this.status = true
	go this.doRecv()
	go this.heartbeat()
	return nil
}

func (this *WebSocketClient) updateHeartbeatTime() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.lastHeartbeatTime = time.Now()
}

func (this *WebSocketClient) getHeartbeatTime() time.Time {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.lastHeartbeatTime
}

//Send data to server
func (this *WebSocketClient) Send(data []byte) error {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if !this.status {
		return fmt.Errorf("WebSocket connect has already closed.")
	}
	return this.conn.WriteMessage(websocket.TextMessage, data)
}

func (this *WebSocketClient) doRecv() {
	for {
		_, data, err := this.conn.ReadMessage()
		if err != nil {
			if this.Status() {
				this.OnError(fmt.Errorf("WebSocketClient host:%s ReadMessage error:%s", this.addr, err))
			}
			return
		}

		this.updateHeartbeatTime()
		this.OnMessage(data)
	}
}

func (this *WebSocketClient) heartbeat() {
	timer := time.NewTimer(this.opts.HeartbeatInterval)
	defer timer.Stop()
	for {
		select {
		case <-this.existCh:
			return
		case <-timer.C:
			err := this.Send(this.opts.HeartbeatPkg)
			if err != nil {
				this.OnError(fmt.Errorf("WebSocketClient send heartbeat error:%s", err))
			}
			timer.Reset(this.opts.HeartbeatInterval)
		}
	}
}

//Status return the status of connection of client and server
func (this *WebSocketClient) Status() bool {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.status
}

//Close the connection of server
func (this *WebSocketClient) Close() error {
	this.lock.Lock()
	defer this.lock.Unlock()

	if !this.status {
		return nil
	}
	this.status = false
	close(this.existCh)
	this.OnClose()
	return this.conn.Close()
}
