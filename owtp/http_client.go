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
	"errors"
	"fmt"
	"github.com/blocktree/go-owcrypt"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/imroc/req"
	"github.com/mr-tron/base58/base58"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
)

//HTTPClient 基于http的通信服务端
type HTTPClient struct {
	responseWriter  http.ResponseWriter
	request         *http.Request
	_auth           Authorization
	handler         PeerHandler
	isHost          bool
	ReadBufferSize  int
	WriteBufferSize int
	pid             string
	isConnect       bool
	mu              sync.RWMutex //读写锁
	closeOnce       sync.Once
	config          ConnectConfig //节点配置
	httpClient      *req.Req
	baseURL         string
	authHeader      map[string]string
}

func HTTPDial(
	pid, url string,
	handler PeerHandler,
	header map[string]string,
	timeout time.Duration) (*HTTPClient, error) {

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
	log.Debug("Connecting URL:", url)

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

	client.httpClient.SetTimeout(timeout)

	client.isConnect = true
	client.isHost = true //我方主动连接
	client.handler.OnPeerOpen(client)

	return client, nil
}

func NewHTTPClientWithHeader(responseWriter http.ResponseWriter, request *http.Request, hander PeerHandler, enableSignature bool) (*HTTPClient, error) {

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
		tmpPublicKey    []byte
		remotePublicKey []byte
		err             error
	)

	//log.Debug("http header:", header)
	header := request.Header
	if header == nil {
		return nil, fmt.Errorf("HTTP Header is nil")
	}

	a := header.Get("a")

	if len(a) == 0 {
		//开启签名授权一定要有授权公钥
		if enableSignature {
			return nil, fmt.Errorf("enableSignature is true, http header should have parameter: a")
		}

		//HTTP的节点ID都采用随机生成，因为是短连接
		_, tmpPublicKey = owcrypt.KeyAgreement_initiator_step1(owcrypt.ECC_CURVE_SM2_STANDARD)

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

	client, err := NewHTTPClient(auth.RemotePID(), responseWriter, request, hander, auth)

	return client, nil
}

func NewHTTPClient(pid string, responseWriter http.ResponseWriter, request *http.Request, hander PeerHandler, auth Authorization) (*HTTPClient, error) {

	if hander == nil {
		return nil, errors.New("hander should not be nil! ")
	}

	client := &HTTPClient{
		pid:            pid,
		_auth:          auth,
		responseWriter: responseWriter,
		request:        request,
		config: ConnectConfig{
			ConnectType: HTTP,
			Address:     request.RemoteAddr,
		},
	}

	client.isConnect = true
	client.setHandler(hander)

	return client, nil
}

func (c *HTTPClient) PID() string {
	return c.pid
}

func (c *HTTPClient) EnableKeyAgreement() bool {
	return c._auth.EnableKeyAgreement()
}

func (c *HTTPClient) auth() Authorization {

	return c._auth
}

func (c *HTTPClient) setHandler(handler PeerHandler) error {
	c.handler = handler
	return nil
}

func (c *HTTPClient) IsHost() bool {
	return c.isHost
}

func (c *HTTPClient) IsConnected() bool {
	return c.isConnect
}

func (c *HTTPClient) ConnectConfig() ConnectConfig {
	return c.config
}

//Close 关闭连接
func (c *HTTPClient) close() error {
	var err error
	//保证节点只关闭一次
	c.closeOnce.Do(func() {

		if !c.isConnect {
			return
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
			//NetWork: c.request.RemoteAddr,
			NetWork: ClientIP(c.request),
		}
		return addr
	}

}

//Send 发送消息
func (c *HTTPClient) send(data DataPacket) error {

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

	if Debug {
		log.Std.Info("%+v", r)
	}

	if err != nil {
		return err
	}

	if r == nil || r.Response() == nil {
		return fmt.Errorf("response is empty ")
	}

	if r.Response().StatusCode != http.StatusOK {
		return fmt.Errorf("%s", r.Response().Status)
	}

	resp := gjson.ParseBytes(r.Bytes())

	packet := NewDataPacket(resp)

	//有可能存在数据已返回，上层才添加请求
	go c.handler.OnPeerNewDataPacketReceived(c, packet)

	return nil
}

//OpenPipe 打开通道
func (c *HTTPClient) openPipe() error {

	//HTTP不需要打开长连接通道

	return nil
}

// writeResponse 输出数据
func (c *HTTPClient) writeResponse(data DataPacket) error {
	respBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if Debug {
		log.Debug("Send: ", string(respBytes))
	}

	if c.responseWriter == nil {
		return fmt.Errorf("responseWriter is nil")
	}
	w := c.responseWriter
	w.Header().Set("Content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	_, err = w.Write(respBytes)
	if err != nil {
		return fmt.Errorf("responseWriter is close")
	}
	//w.(http.Flusher).Flush()
	return nil
}

// readRequest 读取请求
func (c *HTTPClient) HandleRequest() error {

	s, err := ioutil.ReadAll(c.request.Body)
	if err != nil {
		return err
	}

	if Debug {
		log.Debug("Read: ", string(s))
	}

	if len(s) == 0 {
		return fmt.Errorf("body is empty")
	}

	packet := NewDataPacket(gjson.ParseBytes(s))

	//转交给处理器处理数据包
	c.handler.OnPeerNewDataPacketReceived(c, packet)

	return nil

}

func ClientIP(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if ip := req.Header.Get(XRealIP); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get(XForwardedFor); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}
	return remoteAddr
}
