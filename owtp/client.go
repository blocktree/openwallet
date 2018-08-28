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

var (
	Debug = false
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

//Client 基于websocket的通信客户端
type Client struct {
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
}

// Dial connects a client to the given URL.
func Dial(
	pid, url string,
	handler PeerHandler,
	header map[string]string,
	ReadBufferSize, WriteBufferSize int) (*Client, error) {

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
	log.Debug("Connecting URL:", url)

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

func NewClient(pid string, conn *websocket.Conn, hander PeerHandler, auth Authorization, done func()) (*Client, error) {

	if hander == nil {
		return nil, errors.New("hander should not be nil! ")
	}

	client := &Client{
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

func (c *Client) PID() string {
	return c.pid
}

func (c *Client) Auth() Authorization {

	return c.auth
}

func (c *Client) SetHandler(handler PeerHandler) error {
	c.handler = handler
	return nil
}

func (c *Client) IsHost() bool {
	return c.isHost
}

func (c *Client) IsConnected() bool {
	return c.isConnect
}

//Close 关闭连接
func (c *Client) Close() error {
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
func (c *Client) LocalAddr() net.Addr {
	if c.ws == nil {
		return nil
	}
	return c.ws.LocalAddr()
}

//RemoteAddr 远程节点地址
func (c *Client) RemoteAddr() net.Addr {
	if c.ws == nil {
		return nil
	}
	return c.ws.RemoteAddr()
}

//close 内部关闭连接
//func (c *Client) close(active bool) error {
//
//	c.mu.Lock()
//	defer c.mu.Unlock()
//
//	//log.Debug("start close")
//	//
//	log.Debug("peer:", c.PID(), "c.isConnect:", c.isConnect)
//
//	if !c.isConnect {
//		//log.Debug("end close")
//		return errors.New("client has been closed")
//	}
//
//	c.ws.Close()
//	c.isConnect = false
//
//	//主动执行，给通道通知
//	if active {
//		//log.Debug("send close chan:", c.PID())
//		c.closeChan <- struct{}{}
//
//	}
//
//	c.handler.OnPeerClose(c, "client close")
//	//log.Debug("end close")
//	return nil
//}

//Send 发送消息
func (c *Client) Send(data DataPacket) error {

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

	if c.auth != nil && c.auth.EnableAuth() {
		respBytes, err = c.auth.EncryptData(respBytes)
		if err != nil {
			return errors.New("OWTP: EncryptData failed")
		}
	}

	//log.Printf("Send: %s\n", string(respBytes))
	c.send <- respBytes
	return nil
}

//OpenPipe 打开通道
func (c *Client) OpenPipe() error {

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
func (c *Client) writePump() {

	ticker := time.NewTicker(PingPeriod) //发送心跳间隔事件要<等待时间
	defer func() {
		ticker.Stop()
		c.Close()
		log.Debug("writePump end")
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
func (c *Client) write(mt int, message []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(WriteWait)) //设置发送的超时时间点
	return c.ws.WriteMessage(mt, message)
}

// ReadPump 监听消息
func (c *Client) readPump() {

	c.ws.SetReadDeadline(time.Now().Add(PongWait)) //设置客户端心跳响应的最后限期
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(PongWait)) //设置下一次心跳响应的最后限期
		return nil
	})
	defer func() {
		c.Close()
		log.Debug("readPump end")
	}()

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			log.Error("peer:", c.PID(), "Read unexpected error: ", err)
			//close(c.send) //读取通道异常，关闭读通道
			return
		}

		if c.auth != nil && c.auth.EnableAuth() {
			message, err = c.auth.DecryptData(message)
			if err != nil {
				log.Critical("OWTP: DecryptData failed")
				continue
			}
		}

		if Debug {
			log.Debug("Read: ", string(message))
		}

		packet := NewDataPacket(gjson.ParseBytes(message))

		//开一个goroutine处理消息
		go c.handler.OnPeerNewDataPacketReceived(c, packet)

	}
}
