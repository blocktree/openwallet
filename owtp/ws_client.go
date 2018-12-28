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
	"github.com/blocktree/go-owcrypt"
	"github.com/gorilla/websocket"
	"github.com/mr-tron/base58/base58"
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

//WSClient 基于websocket的通信客户端
type WSClient struct {
	_auth           Authorization
	ws              *websocket.Conn
	handler         PeerHandler
	_send           chan []byte
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
	ReadBufferSize, WriteBufferSize int) (*WSClient, error) {

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

	client, err := NewWSClient(pid, ws, handler, nil, nil)
	if err != nil {
		return nil, err
	}

	client.isConnect = true
	client.isHost = true //我方主动连接
	client.handler.OnPeerOpen(client)

	return client, nil
}

func NewWSClientWithHeader(header http.Header, cert Certificate, conn *websocket.Conn, handler PeerHandler, enableSignature bool, done func()) (*WSClient, error) {

	/*
		| 参数名称 | 类型   | 是否可空        | 描述                                                                              |
		|----------|--------|----------------|---------------------------------------------------------------------------------|
		| a        | string | 否              | 节点公钥，base58，http带入将开启签名                                                |
		| n        | uint32 | (websocket必填) | 请求序号。为了保证请求对应响应按序执行，并防御重放攻击，序号可以为随机数，但不可重复。 |
		| t        | uint32 | (websocket必填) | 时间戳。限制请求在特定时间范围内有效，如10分钟。                                     |
		| s        | string | (websocket必填) | 组合[a+n+t]并sha256两次，使用钱包工具配置的本地私钥签名，最后base58编码         |
	*/

	var (
		//enableSig       bool
		//isConsult       bool
		tmpPublicKey    []byte
		remotePublicKey []byte
		err             error
		//nodeID          string
	)

	//log.Debug("http header:", header)

	a := header.Get("a")

	//HTTP的节点ID都采用随机生成，因为是短连接
	_, tmpPublicKey = owcrypt.KeyAgreement_initiator_step1(owcrypt.ECC_CURVE_SM2_STANDARD)

	if len(a) == 0 {
		//没有授权公钥，不授权的HTTP访问，不建立协商密码，不进行签名授权
		remotePublicKey = tmpPublicKey
	} else {
		//有授权公钥，必须授权的HTTP访问，不建立协商密码，进行签名授权
		remotePublicKey, err = base58.Decode(a)
		if err != nil {
			return nil, err
		}
	}

	//开启签名授权，验证header的签名是否合法
	if enableSignature {
		//校验header的签名
		if !VerifyHeaderSignature(header, remotePublicKey) {
			return nil, fmt.Errorf("the signature in http header is not invalid")
		}
	}

	auth := &OWTPAuth{
		remotePublicKey: remotePublicKey,
		enable:          enableSignature,
		localPublicKey:  cert.PublicKeyBytes(),
		localPrivateKey: cert.PrivateKeyBytes(),
	}

	client, err := NewWSClient(auth.RemotePID(), conn, handler, auth, done)

	return client, nil
}

func NewWSClient(pid string, conn *websocket.Conn, handler PeerHandler, auth Authorization, done func()) (*WSClient, error) {

	if handler == nil {
		return nil, errors.New("handler should not be nil! ")
	}

	client := &WSClient{
		pid:   pid,
		ws:    conn,
		_send: make(chan []byte, MaxMessageSize),
		_auth: auth,
		done:  done,
	}

	client.isConnect = true
	client.setHandler(handler)

	return client, nil
}

//VerifyHeaderSignature 校验签名，若验证错误，可更新错误信息到DataPacket中
func VerifyHeaderSignature(header http.Header, publicKey []byte) bool {

	a := header.Get("a")
	n := header.Get("n")
	t := header.Get("t")
	s := header.Get("s")

	if len(a) == 0 || len(s) == 0 || len(t) == 0 || len(n) == 0 {
		return false
	}

	plainText := fmt.Sprintf("%s%s%s", a, n, t)
	//log.Debug("VerifySignature plainText: ", plainText)
	hash := owcrypt.Hash([]byte(plainText), 0, owcrypt.HASh_ALG_DOUBLE_SHA256)
	//log.Debug("VerifySignature hash: ", hex.EncodeToString(hash))
	nodeID := owcrypt.Hash(publicKey, 0, owcrypt.HASH_ALG_SHA256)
	//log.Debug("VerifySignature remotePublicKey: ", hex.EncodeToString(auth.remotePublicKey))
	//log.Debug("VerifySignature nodeID: ", hex.EncodeToString(nodeID))
	signature, err := base58.Decode(s)
	if err != nil {
		return false
	}
	ret := owcrypt.Verify(publicKey, nodeID, 32, hash, 32, signature, owcrypt.ECC_CURVE_SM2_STANDARD)
	if ret != owcrypt.SUCCESS {
		return false
	}

	return true
}

func (c *WSClient) PID() string {
	return c.pid
}

func (c *WSClient) auth() Authorization {

	return c._auth
}

func (c *WSClient) setHandler(handler PeerHandler) error {
	c.handler = handler
	return nil
}

func (c *WSClient) IsHost() bool {
	return c.isHost
}

func (c *WSClient) IsConnected() bool {
	return c.isConnect
}

func (c *WSClient) GetConfig() map[string]string {
	return c.config
}

//Close 关闭连接
func (c *WSClient) close() error {
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
func (c *WSClient) LocalAddr() net.Addr {
	if c.ws == nil {
		return nil
	}
	return c.ws.LocalAddr()
}

//RemoteAddr 远程节点地址
func (c *WSClient) RemoteAddr() net.Addr {
	if c.ws == nil {
		return nil
	}
	return c.ws.RemoteAddr()
}

//Send 发送消息
func (c *WSClient) send(data DataPacket) error {

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
	c._send <- respBytes
	return nil
}

//OpenPipe 打开通道
func (c *WSClient) openPipe() error {

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
func (c *WSClient) writePump() {

	ticker := time.NewTicker(PingPeriod) //发送心跳间隔事件要<等待时间
	defer func() {
		ticker.Stop()
		c.close()
		//log.Debug("writePump end")
	}()
	for {
		select {
		case message, ok := <-c._send:
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
func (c *WSClient) write(mt int, message []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(WriteWait)) //设置发送的超时时间点
	return c.ws.WriteMessage(mt, message)
}

// ReadPump 监听消息
func (c *WSClient) readPump() {

	c.ws.SetReadDeadline(time.Now().Add(PongWait)) //设置客户端心跳响应的最后限期
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(PongWait)) //设置下一次心跳响应的最后限期
		return nil
	})
	defer func() {
		c.close()
		//log.Debug("readPump end")
	}()

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			log.Error("peer:", c.PID(), "Read unexpected error: ", err)
			//close(c.send) //读取通道异常，关闭读通道
			return
		}

		if Debug {
			log.Debug("Read: ", string(message))
		}

		packet := NewDataPacket(gjson.ParseBytes(message))

		//开一个goroutine处理消息
		go c.handler.OnPeerNewDataPacketReceived(c, packet)

	}
}
