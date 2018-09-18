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
	"sync"
	"time"
	"net/http"
	"io/ioutil"
)

//HTTPClient 基于http的通信服务端
type HTTPClient struct {
	responseWriter  http.ResponseWriter
	request         *http.Request
	auth            Authorization
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

func HTTPDial(
	pid, url string,
	handler PeerHandler,
	header map[string]string,
	ReadBufferSize, WriteBufferSize int) ( error) {

	var (
		httpHeader http.Header
	)

	if handler == nil {
		return  errors.New("hander should not be nil! ")
	}

	//处理连接授权
	//authURL := url
	//if auth != nil && auth.EnableAuth() {
	//	authURL = auth.ConnectAuth(url)
	//}
	log.Debug("Connecting URL:", url)


	if header != nil {
		httpHeader = make(http.Header)
		for key, value := range header {
			httpHeader.Add(key, value)
		}
	}

	return  nil
}


func NewHTTPClient(pid string,responseWriter  http.ResponseWriter, request *http.Request , hander PeerHandler, auth Authorization, done func()) (*HTTPClient, error) {

	if hander == nil {
		return nil, errors.New("hander should not be nil! ")
	}

	client := &HTTPClient{
		pid:  pid,
		send: make(chan []byte, MaxMessageSize),
		auth: auth,
		done: done,
		responseWriter:responseWriter,
		request:request,
	}

	client.isConnect = true
	client.SetHandler(hander)

	return client, nil
}

func (c *HTTPClient) PID() string {
	return c.pid
}

func (c *HTTPClient) Auth() Authorization {

	return c.auth
}

func (c *HTTPClient) SetHandler(handler PeerHandler) error {
	c.handler = handler
	return nil
}

func (c *HTTPClient) IsHost() bool {
	return c.isHost
}

func (c *HTTPClient) IsConnected() bool {
	return c.isConnect
}

func (c *HTTPClient) GetConfig() map[string]string {
	return c.config
}

//Close 关闭连接
func (c *HTTPClient) Close() error {
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
		c.isConnect = false
		c.handler.OnPeerClose(c, "client close")
	})
	return err
}

//LocalAddr 本地节点地址
func (c *HTTPClient) LocalAddr() net.Addr {
	if c.request == nil {
		return nil
	}
	addr := &MqAddr{
		NetWork: c.request.Host,
	}
	return addr
}

//RemoteAddr 远程节点地址
func (c *HTTPClient) RemoteAddr() net.Addr {
	if c.request == nil {
		return nil
	}
	addr := &MqAddr{
		NetWork: c.request.RemoteAddr,
	}
	return addr
}

//Send 发送消息
func (c *HTTPClient) Send(data DataPacket) error {

	////添加授权
	//if c.auth != nil && c.auth.EnableAuth() {
	//	if !c.auth.GenerateSignature(&data) {
	//		return errors.New("OWTP: authorization failed")
	//	}
	//}
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
	if err := c.write(websocket.TextMessage, respBytes); err != nil {
		return nil
	}
	return nil
}

//OpenPipe 打开通道
func (c *HTTPClient) OpenPipe() error {

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
func (c *HTTPClient) writePump() {

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
func (c *HTTPClient) write(mt int, message []byte) error {
	defer c.Close()
	if c.responseWriter == nil {
		return fmt.Errorf("responseWriter is nil")
	}
	w := c.responseWriter
	//fmt.Fprint(c.responseWriter,[]byte(message))
	w.Write([]byte(message))
	w.(http.Flusher).Flush()
	return nil
}

// ReadPump 监听消息
func (c *HTTPClient) readPump() {
	s, _ := ioutil.ReadAll(c.request.Body)
	packet := NewDataPacket(gjson.ParseBytes(s))
	c.handler.OnPeerNewDataPacketReceived(c, packet)
}
