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
	"errors"
	"fmt"
	"github.com/blocktree/OpenWallet/log"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"net"
	"net/http"
	"sync"
	"time"
)

//局部常量
const (
	WriteWait      = 60 * time.Second //超时为6秒
	PongWait       = 30 * time.Second
	PingPeriod     = (PongWait * 9) / 10
	MaxMessageSize = 1 * 1024
)

//WebSocketClient 基于websocket的通信客户端
type WebSocketClient struct {
	auth            Authorization
	ws              *websocket.Conn
	handler         PeerHandler
	send            chan []byte
	isHost          bool
	ReadBufferSize  int
	WriteBufferSize int
	pid             string
	isConnect       bool
	mu              sync.RWMutex //读写锁
	closeOnce       sync.Once
	done            func()
	config          map[string]string //节点配置
}

// Dial connects a client to the given URL.
func Dial(
	pid, url string,
	handler PeerHandler,
	header map[string]string,
	ReadBufferSize, WriteBufferSize int) (*WebSocketClient, error) {

	var (
		httpHeader http.Header
	)

	if handler == nil {
		return nil, errors.New("hander should not be nil! ")
	}

	//处理连接授权
	//authURL := url
	//if auth != nil && auth.EnableAuth() {
	//	authURL = auth.ConnectAuth(url)
	//}
	log.Info("Connecting URL:", url)

	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 60 * time.Second,
		ReadBufferSize:   ReadBufferSize,
		WriteBufferSize:  WriteBufferSize,
	}

	if header != nil {
		httpHeader = make(http.Header)
		for key, value := range header {
			httpHeader.Add(key, value)
		}
	}

	ws, _, err := dialer.Dial(url, httpHeader)
	if err != nil {
		return nil, err
	}

	client, err := NewClient(pid, ws, handler, nil, nil)
	if err != nil {
		return nil, err
	}

	client.isConnect = true
	client.isHost = true //我方主动连接
	client.handler.OnPeerOpen(client)

	return client, nil
}

func NewClient(pid string, conn *websocket.Conn, hander PeerHandler, auth Authorization, done func()) (*WebSocketClient, error) {

	if hander == nil {
		return nil, errors.New("hander should not be nil! ")
	}

	client := &WebSocketClient{
		pid:  pid,
		ws:   conn,
		send: make(chan []byte, MaxMessageSize),
		auth: auth,
		done: done,
	}

	client.isConnect = true
	client.SetHandler(hander)

	return client, nil
}

func (c *WebSocketClient) PID() string {
	return c.pid
}

func (c *WebSocketClient) Auth() Authorization {

	return c.auth
}

func (c *WebSocketClient) SetHandler(handler PeerHandler) error {
	c.handler = handler
	return nil
}

func (c *WebSocketClient) IsHost() bool {
	return c.isHost
}

func (c *WebSocketClient) IsConnected() bool {
	return c.isConnect
}

func (c *WebSocketClient) GetConfig() map[string]string {
	return c.config
}

//Close 关闭连接
func (c *WebSocketClient) Close() error {
	var err error

	//保证节点只关闭一次
	c.closeOnce.Do(func() {

		if !c.isConnect {
			//log.Debug("end close")
			return
		}

		//调用关闭函数通知上级
		if c.done != nil {
			c.done()
			// Be nice to GC
			c.done = nil
		}

		err = c.ws.Close()
		c.isConnect = false
		c.handler.OnPeerClose(c, "client close")
	})
	return err
}

//LocalAddr 本地节点地址
func (c *WebSocketClient) LocalAddr() net.Addr {
	if c.ws == nil {
		return nil
	}
	return c.ws.LocalAddr()
}

//RemoteAddr 远程节点地址
func (c *WebSocketClient) RemoteAddr() net.Addr {
	if c.ws == nil {
		return nil
	}
	return c.ws.RemoteAddr()
}

//Send 发送消息
func (c *WebSocketClient) Send(data DataPacket) error {

	//添加授权
	if c.auth != nil && c.auth.EnableAuth() {
		if !c.auth.GenerateSignature(&data) {
			return errors.New("OWTP: authorization failed")
		}
	}
	//log.Emergency("Send DataPacket:", data)
	respBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	//if c.auth != nil && c.auth.EnableAuth() {
	//	respBytes, err = c.auth.EncryptData(respBytes)
	//	if err != nil {
	//		return errors.New("OWTP: EncryptData failed")
	//	}
	//}

	//log.Printf("Send: %s\n", string(respBytes))
	c.send <- respBytes
	return nil
}

//OpenPipe 打开通道
func (c *WebSocketClient) OpenPipe() error {

	if !c.IsConnected() {
		return fmt.Errorf("client is not connect")
	}

	//发送通道
	go c.writePump()

	//监听消息
	go c.readPump()

	return nil
}

// WritePump 发送消息通道
func (c *WebSocketClient) writePump() {

	ticker := time.NewTicker(PingPeriod) //发送心跳间隔事件要<等待时间
	defer func() {
		ticker.Stop()
		c.Close()
		//log.Debug("writePump end")
	}()
	for {
		select {
		case message, ok := <-c.send:
			//发送消息
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if Debug {
				log.Debug("Send: ", string(message))
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			//定时器的回调,发送心跳检查,
			err := c.write(websocket.PingMessage, []byte{})

			if err != nil {
				return //客户端不响应心跳就停止
			}

		}
	}
}

// write 输出数据
func (c *WebSocketClient) write(mt int, message []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(WriteWait)) //设置发送的超时时间点
	return c.ws.WriteMessage(mt, message)
}

// ReadPump 监听消息
func (c *WebSocketClient) readPump() {

	c.ws.SetReadDeadline(time.Now().Add(PongWait)) //设置客户端心跳响应的最后限期
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(PongWait)) //设置下一次心跳响应的最后限期
		return nil
	})
	defer func() {
		c.Close()
		//log.Debug("readPump end")
	}()

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			log.Error("peer:", c.PID(), "Read unexpected error: ", err)
			//close(c.send) //读取通道异常，关闭读通道
			return
		}

		//if c.auth != nil && c.auth.EnableAuth() {
		//	message, err = c.auth.DecryptData(message)
		//	if err != nil {
		//		log.Critical("OWTP: DecryptData failed")
		//		continue
		//	}
		//}

		if Debug {
			log.Debug("Read: ", string(message))
		}

		packet := NewDataPacket(gjson.ParseBytes(message))

		//开一个goroutine处理消息
		go c.handler.OnPeerNewDataPacketReceived(c, packet)

	}
}
