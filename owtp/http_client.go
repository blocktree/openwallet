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
	"github.com/imroc/req"
	"github.com/mr-tron/base58/base58"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
)

//HTTPClient 基于http的通信服务端
type HTTPClient struct {
	responseWriter  http.ResponseWriter
	request         *http.Request
	auth            Authorization
	handler         PeerHandler
	isHost          bool
	ReadBufferSize  int
	WriteBufferSize int
	pid             string
	isConnect       bool
	mu              sync.RWMutex //读写锁
	closeOnce       sync.Once
	done            func()
	config          map[string]string //节点配置
	httpClient      *req.Req
	baseURL         string
	authHeader      map[string]string
}

func HTTPDial(
	pid, url string,
	handler PeerHandler,
	header map[string]string) (*HTTPClient, error) {

	//var (
	//	httpHeader http.Header
	//)

	if handler == nil {
		return nil, errors.New("hander should not be nil! ")
	}

	//处理连接授权
	//authURL := url
	//if auth != nil && auth.EnableAuth() {
	//	authURL = auth.ConnectAuth(url)
	//}
	log.Info("Connecting URL:", url)

	//if header != nil {
	//	httpHeader = make(http.Header)
	//	for key, value := range header {
	//		httpHeader.Add(key, value)
	//	}
	//}

	client := &HTTPClient{
		pid:        pid,
		baseURL:    url,
		httpClient: req.New(),
		authHeader: header,
		handler:    handler,
	}

	client.isConnect = true
	client.isHost = true //我方主动连接
	client.handler.OnPeerOpen(client)

	return client, nil
}

func NewHTTPClientWithHeader(header http.Header, responseWriter http.ResponseWriter, request *http.Request, hander PeerHandler, enableSignature bool, done func()) (*HTTPClient, error) {

	/*
			| 参数名称 | 类型   | 是否可空   | 描述                                                                              |
		|----------|--------|------------|-----------------------------------------------------------------------------------|
		| a        | string | 是         | 节点公钥，base58                                                                 |
		| p        | string | 是         | 计算协商密码的临时公钥，base58                                                                 |
		| n        | uint32 | 123        | 请求序号。为了保证请求对应响应按序执行，并防御重放攻击，序号可以为随机数，但不可重复。 |
		| t        | uint32 | 1528520843 | 时间戳。限制请求在特定时间范围内有效，如10分钟。                                     |
		| c        | string | 是         | 协商密码，生成的密钥类别，aes-128-ctr，aes-128-cbc，aes-256-ecb等，为空不进行协商      |
		| s        | string | 是         | 组合[a+n+t]并sha256两次，使用钱包工具配置的本地私钥签名，最后base58编码             |
	*/

	var (
		//enableSig       bool
		//isConsult       bool
		tmpPublicKey    []byte
		remotePublicKey []byte
		err             error
		//nodeID          string
	)

	//log.Debug("header:", header)

	a := header.Get("a")

	//HTTP的节点ID都采用随机生成，因为是短连接
	_, tmpPublicKey = owcrypt.KeyAgreement_initiator_step1(owcrypt.ECC_CURVE_SM2_STANDARD)

	if len(a) == 0 {
		//开启签名授权一定要有授权公钥
		if enableSignature {
			return nil, fmt.Errorf("enableSignature is true, http header should have parameter: a")
		}

		//没有授权公钥，不授权的HTTP访问，不建立协商密码，不进行签名授权
		remotePublicKey = tmpPublicKey
	} else {
		//有授权公钥，必须授权的HTTP访问，不建立协商密码，进行签名授权
		remotePublicKey, err = base58.Decode(a)
		if err != nil {
			return nil, err
		}
	}

	//nodeID = base58.Encode(owcrypt.Hash(tmpPublicKey, 0, owcrypt.HASH_ALG_SHA256))

	auth := &OWTPAuth{
		remotePublicKey: remotePublicKey,
		enable:          enableSignature,
	}

	//err = auth.VerifyHeader()
	//if err != nil {
	//	return nil, err
	//}

	client, err := NewHTTPClient(auth.RemotePID(), responseWriter, request, hander, auth, done)

	return client, nil
}

func NewHTTPClient(pid string, responseWriter http.ResponseWriter, request *http.Request, hander PeerHandler, auth Authorization, done func()) (*HTTPClient, error) {

	if hander == nil {
		return nil, errors.New("hander should not be nil! ")
	}

	client := &HTTPClient{
		pid:            pid,
		auth:           auth,
		done:           done,
		responseWriter: responseWriter,
		request:        request,
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

	if c.isHost {
		addr := &MqAddr{
			NetWork: c.baseURL,
		}
		return addr
	} else {
		if c.request == nil {
			return nil
		}
		addr := &MqAddr{
			NetWork: c.request.RemoteAddr,
		}
		return addr
	}

}

//Send 发送消息
func (c *HTTPClient) Send(data DataPacket) error {

	if c.isHost {
		return c.sendHTTPRequest(data)
	} else {
		return c.writeResponse(data)
	}
}

//sendHTTPRequest 发送HTTP请求
func (c *HTTPClient) sendHTTPRequest(data DataPacket) error {

	if c.httpClient == nil {
		return errors.New("API url is not setup. ")
	}

	r, err := c.httpClient.Post(c.baseURL, req.BodyJSON(&data), req.Header(c.authHeader))

	log.Std.Info("%+v", r)

	if err != nil {
		return err
	}

	resp := gjson.ParseBytes(r.Bytes())

	packet := NewDataPacket(resp)

	//有可能存在数据已返回，上层才添加请求
	go c.handler.OnPeerNewDataPacketReceived(c, packet)

	return nil
}

//OpenPipe 打开通道
func (c *HTTPClient) OpenPipe() error {

	//对方节点时远程服务器，不需要打开读通道
	if c.isHost {
		return nil
	}

	if !c.IsConnected() {
		return fmt.Errorf("client is not connect")
	}

	//发送通道
	//go c.writePump()

	//监听消息
	go c.readPump()

	return nil
}

// writeResponse 输出数据
func (c *HTTPClient) writeResponse(data DataPacket) error {
	//log.Debug("writeResponse")
	respBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	defer c.Close()
	if c.responseWriter == nil {
		return fmt.Errorf("responseWriter is nil")
	}
	w := c.responseWriter
	w.Header().Set("Content-type","application/json")
	_,err = w.Write(respBytes)
	if err != nil {
		return fmt.Errorf("responseWriter is close")
	}
	w.(http.Flusher).Flush()
	return nil
}

// ReadPump 监听消息
func (c *HTTPClient) readPump() {

	defer func() {
		c.Close()
		//log.Debug("readPump end")
	}()

	s, _ := ioutil.ReadAll(c.request.Body)

	w := c.responseWriter

	result := string(s)

	if len(result) == 0 {
		w.Write([]byte("is not a Json"))
		w.(http.Flusher).Flush()
		log.Error("is not a Json ")
		return
	}

	//if Debug {

		//log.Debug("readPump: ", string(s))
	//}

	packet := NewDataPacket(gjson.ParseBytes(s))

	if len(packet.Method) == 0 {

		w.Write([]byte("method is null"))
		w.(http.Flusher).Flush()
		log.Error(" method is null ")
		return
	}

	c.handler.OnPeerNewDataPacketReceived(c, packet)

}
