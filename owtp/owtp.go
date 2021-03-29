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

// owtp全称OpenWallet Transfer Protocol，OpenWallet的一种点对点的分布式私有通信协议。
package owtp

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/bwmarrin/snowflake"
	"github.com/mr-tron/base58/base58"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type ConnectType int

const (

	//成功标识
	StatusSuccess uint64 = 200

	//客户端请求错误
	ErrBadRequest uint64 = 400
	//网络断开
	ErrUnauthorized uint64 = 401
	//通信密钥不正确
	ErrSecretKeyInvalid uint64 = 402
	//找不到方法
	ErrNotFoundMethod uint64 = 404
	//重放攻击
	ErrReplayAttack uint64 = 409
	//请求超时
	ErrRequestTimeout uint64 = 408
	//网络断开
	ErrNetworkDisconnected uint64 = 430
	//服务器错误
	ErrInternalServerError uint64 = 500
	//请求与响应的方法不一致
	ErrResponseMethodDiffer uint64 = 501
	//协商失败
	ErrKeyAgreementFailed uint64 = 502
	//拒绝服务
	ErrDenialOfService uint64 = 503

	//60X: 自定义错误
	ErrCustomError uint64 = 600
)

//连接方式
const (
	Websocket string = "ws"
	MQ        string = "mq"
	HTTP      string = "http"
)

//内置方法
const (

	//校验协商结果
	KeyAgreementMethod = "internal_keyAgreement"

	//准备前执行的方
	PrepareMethod = "internal_prepare"

	//结束时执行的方法
	FinishMethod = "internal_finish"

	//内置缓存主健关键字
	keyAgreementCipher = "internal_key_keyAgreementCipher"
)

var (
	Debug = false
)

//节点主配置 作为json解析工具
type ConnectConfig struct {
	Address            string `json:"address"`            //@required 连接IP地址
	ConnectType        string `json:"connectType"`        //@required 连接方式
	EnableSignature    bool   `json:"enableSignature"`    //是否开启owtp协议内签名，防重放
	Account            string `json:"account"`            //mq账户名
	Password           string `json:"password"`           //mq账户密码
	Exchange           string `json:"exchange"`           //mq需要字段
	WriteQueueName     string `json:"writeQueueName"`     //mq写入通道名
	ReadQueueName      string `json:"readQueueName"`      //mq读取通道名
	EnableSSL          bool   `json:"enableSSL"`          //是否开启链接SSL，https，wss
	ReadBufferSize     int    `json:"readBufferSize"`     //socket读取缓存
	WriteBufferSize    int    `json:"writeBufferSize"`    //socket写入缓存
	EnableKeyAgreement bool   `json:"enableKeyAgreement"` //是否开启协商密码
}

//节点主配置 作为json解析工具
type NodeConfig struct {
	TimeoutSEC int         `json:"timeoutSEC"` //超时时间
	Cert       Certificate `json:"cert"`       //证书
	Peerstore  Peerstore   //会话缓存
}

//OWTPNode 实现OWTP协议的节点
type OWTPNode struct {
	//nonce生成器
	nonceGen *snowflake.Node
	//缓存文件
	cacheFile string
	//默认路由
	serveMux *ServeMux
	//是否监听连接中
	//listening bool
	//节点运行中
	running bool
	//读写锁
	mu sync.RWMutex
	//关闭连接时的回调
	disconnectHandler func(n *OWTPNode, peerInfo PeerInfo)
	//连接时的回调
	connectHandler func(n *OWTPNode, peerInfo PeerInfo)
	//节点存储器
	peerstore Peerstore
	//在线节点
	onlinePeers map[string]Peer
	//服务监听器
	//listener Listener
	//服务监听器
	listeners map[string]Listener
	//授权证书
	cert Certificate
	//Broadcast   chan BroadcastMessage
	Join  chan Peer
	Leave chan Peer
	Stop  chan struct{}
	//请求超时（秒）
	timeoutSEC int
	//通道的读写缓存大小
	//ReadBufferSize, WriteBufferSize int
	//重新加载节点连接信息
	reloadPeerInfoHandler func(n *OWTPNode, peerID string) PeerInfo
}

//RandomOWTPNode 创建随机密钥节点
func RandomOWTPNode(consultType ...string) *OWTPNode {
	c := ""
	if len(consultType) > 0 {
		c = consultType[0]
	}
	cert, err := NewCertificate(RandomPrivateKey(), c)
	if err != nil {
		return nil
	}

	config := NodeConfig{
		Cert: cert,
	}

	node := NewNode(config)
	return node
}

//NewNode 创建OWTP协议节点
func NewNode(config NodeConfig) *OWTPNode {

	node := &OWTPNode{}

	if config.Cert.ID() == "" {
		cert, err := NewCertificate(RandomPrivateKey(), "")
		if err != nil {
			return nil
		}
		node.cert = cert
	} else {
		node.cert = config.Cert
	}

	if config.Peerstore == nil {
		node.peerstore = NewOWTPPeerstore()
	} else {
		node.peerstore = config.Peerstore
	}

	node.timeoutSEC = config.TimeoutSEC
	if node.timeoutSEC == 0 {
		node.serveMux = NewServeMux(DefaultTimoutSEC)
	} else {
		node.serveMux = NewServeMux(node.timeoutSEC)
	}

	node.nonceGen, _ = snowflake.NewNode(1)

	node.Join = make(chan Peer)
	node.Leave = make(chan Peer)
	node.Stop = make(chan struct{})
	node.onlinePeers = make(map[string]Peer)
	node.listeners = make(map[string]Listener)

	//内部配置一个协商密码处理过程
	node.serveMux.handleFuncInner(KeyAgreementMethod, node.keyAgreement)

	//马上执行
	go node.Run()

	return node
}

//NewOWTPNode 创建OWTP协议节点
func NewOWTPNode(cert Certificate, readBufferSize, writeBufferSize int) *OWTPNode {

	node := &OWTPNode{}
	node.nonceGen, _ = snowflake.NewNode(1)
	node.serveMux = NewServeMux(DefaultTimoutSEC)
	node.cert = cert
	node.peerstore = NewOWTPPeerstore()
	//node.ReadBufferSize = readBufferSize
	//node.WriteBufferSize = writeBufferSize
	node.Join = make(chan Peer)
	node.Leave = make(chan Peer)
	node.Stop = make(chan struct{})
	node.onlinePeers = make(map[string]Peer)
	node.listeners = make(map[string]Listener)

	//内部配置一个协商密码处理过程
	node.serveMux.handleFuncInner(KeyAgreementMethod, node.keyAgreement)

	//马上执行
	go node.Run()

	return node
}

//Certificate 节点证书
func (node *OWTPNode) Certificate() *Certificate {
	return &node.cert
}

//NodeID 节点的ID
func (node *OWTPNode) NodeID() string {
	return node.cert.ID()
}

//SetPeerstore 设置一个Peerstore指针
func (node *OWTPNode) SetPeerstore(store Peerstore) {
	node.peerstore = store
}

//Peerstore 节点存储器
func (node *OWTPNode) Peerstore() Peerstore {
	return node.peerstore
}

//Listen 监听TCP地址
func (node *OWTPNode) Listen(config ConnectConfig) error {

	addr := config.Address
	connectType := config.ConnectType
	enableSignature := config.EnableSignature
	//log.Debug("enableSignature:", enableSignature)
	if _, exist := node.listeners[connectType]; exist {
		return fmt.Errorf("the node [%s] is listening, please close listener first", connectType)
	}
	//if node.listening {
	//	return fmt.Errorf("the node is listening, please close listener first")
	//}

	if connectType == Websocket || connectType == MQ {
		l, err := WSListenAddr(addr, node.cert, enableSignature, node)
		if err != nil {
			return err
		}
		node.listeners[connectType] = l

		go func(listener Listener) {
			for {
				peer, err := listener.Accept()
				if err != nil {
					return
				}
				node.Join <- peer
			}
		}(l)

		//node.listening = true
	} else if connectType == HTTP {
		l, err := HttpListenAddr(addr, enableSignature, node)
		if err != nil {
			return err
		}
		node.listeners[connectType] = l
		//node.listening = true
	}

	return nil
}

//listening 是否监听中
func (node *OWTPNode) Listening(connectType string) bool {
	_, exist := node.listeners[connectType]
	return exist
}

//CloseListener 关闭监听
func (node *OWTPNode) CloseListener(connectType string) {
	if l, exist := node.listeners[connectType]; exist {
		l.Close()
		node.mu.Lock()
		delete(node.listeners, connectType)
		node.mu.Unlock()
	}
}

//Connect 建立长连接
func (node *OWTPNode) Connect(pid string, config ConnectConfig) (Peer, error) {

	if len(config.Address) == 0 {
		//连接信息已丢失，需要重新加载
		if node.reloadPeerInfoHandler != nil {
			peerInfo := node.reloadPeerInfoHandler(node, pid)
			config = peerInfo.Config
		}
	}

	peer, err := node.connect(pid, config)
	if err != nil {
		return nil, err
	}

	//连接配置EnableKeyAgreement = true，进行协商密码
	if config.EnableKeyAgreement {

		//该节点未开启，首先请求开启协商密码
		if peer != nil && peer.auth().EnableKeyAgreement() == false {
			log.Debugf("first connect to call KeyAgreement request")
			err = node.KeyAgreement(pid, "aes")
			if err != nil {
				return nil, err
			}
		}
	}

	return peer, nil
}

//connect 建立长连接，内部调用
func (node *OWTPNode) connect(pid string, config ConnectConfig) (Peer, error) {

	var (
		peer Peer
	)

	addr := config.Address
	connectType := config.ConnectType
	enableSignature := config.EnableSignature
	enableSSL := config.EnableSSL
	readBufferSize := config.ReadBufferSize
	writeBufferSize := config.WriteBufferSize
	timeout := time.Duration(node.timeoutSEC) * time.Second

	//检查是否已经连接服务
	peer = node.GetOnlinePeer(pid)
	if peer != nil && peer.IsConnected() && peer.ConnectConfig().Address != "" {
		//如果地址不一致，先关闭节点
		if addr != peer.ConnectConfig().Address || connectType != peer.ConnectConfig().ConnectType {
			//fmt.Printf("如果地址不一致，先关闭节点 ClosePeer, %s:%s \n",addr,peer.ConnectConfig().Address)
			node.ClosePeer(pid)
		} else {
			//log.Debugf("peer[%s] has connected", peer.PID())
			return peer, nil
		}
	}

	if len(addr) == 0 {
		return nil, fmt.Errorf("pid: %s, address must contain by config", pid)
	}

	auth, err := NewOWTPAuthWithCertificate(node.cert, enableSignature)

	//发起协商密钥
	//err = auth.InitKeyAgreement()
	if err != nil {
		return nil, err
	}

	if len(connectType) == 0 {
		return nil, fmt.Errorf("connectType must contain by config")
	}

	//websocket类型
	if connectType == Websocket {

		protocol := "ws://"

		if enableSSL {
			protocol = "wss://"
		}

		url := protocol + strings.TrimSuffix(addr, "/") + "/"

		//建立链接，记录默认的客户端
		client, err := Dial(pid, url, node, auth.HTTPAuthHeader(), readBufferSize, writeBufferSize)
		if err != nil {
			return nil, err
		}
		//设置授权规则
		client._auth = auth
		//设置配置
		client.config = config
		peer = client
	}

	//MQ类型
	if connectType == MQ {

		mqAccount := config.Account
		mqPassword := config.Password
		url := "amqp://" + mqAccount + ":" + mqPassword + "@" + strings.TrimSuffix(addr, "/") + "/"

		//建立链接，记录默认的客户端
		client, err := MQDial(pid, url, node)
		if err != nil {
			return nil, err
		}
		//设置授权规则
		client._auth = auth
		//设置配置
		client.config = config
		peer = client
	}

	//HTTP类型
	if connectType == HTTP {

		protocol := "http://"

		if enableSSL {
			protocol = "https://"
		}

		url := protocol + strings.TrimSuffix(addr, "/") + "/"

		//建立链接，记录默认的客户端
		client, err := HTTPDial(pid, url, node, auth.HTTPAuthHeader(), timeout)
		if err != nil {
			return nil, err
		}
		//设置授权规则
		client._auth = auth
		//设置配置
		client.config = config

		peer = client
	}

	if peer == nil {
		return nil, errors.New("connectType can't found! ")
	}

	node.AddOnlinePeer(peer)

	return peer, nil
}

//SetOpenHandler 设置开启连接时的回调
func (node *OWTPNode) SetOpenHandler(h func(n *OWTPNode, peer PeerInfo)) {
	node.connectHandler = h
}

//SetCloseHandler 设置关闭连接时的回调
func (node *OWTPNode) SetCloseHandler(h func(n *OWTPNode, peer PeerInfo)) {
	node.disconnectHandler = h
}

// SetReloadPeerInfoHandler 设置重新加载节点信息回调
func (node *OWTPNode) SetReloadPeerInfoHandler(h func(n *OWTPNode, peerID string) PeerInfo) {
	node.reloadPeerInfoHandler = h
}

// Run 运行,go Run运行一条线程
func (node *OWTPNode) Run() error {

	defer func() {
		node.running = false
	}()

	if node.running {
		return fmt.Errorf("node is running")
	}

	node.running = true

	//开启节点管理运行时
	for {
		select {
		case peer := <-node.Join:
			//客户端加入
			log.Debug("Node Join:", peer.PID())
			log.Debug("Node IP:", peer.RemoteAddr().String())
			node.AddOnlinePeer(peer)
			node.peerstore.SavePeer(peer) //HTTP可能会无限增加
			//加入后打开数据流通道
			if err := peer.openPipe(); err != nil {
				log.Error("peer:", peer.PID(), "open pipe failed")
				continue
			}

			if node.connectHandler != nil {
				go node.connectHandler(node, node.Peerstore().PeerInfo(peer.PID()))
			}

		case peer := <-node.Leave:
			//客户端离开
			log.Debug("Node Leave:", peer.PID())

			node.serveMux.ResetRequestQueue(peer.PID())
			node.RemoveOfflinePeer(peer.PID())

			if node.disconnectHandler != nil {
				go node.disconnectHandler(node, node.Peerstore().PeerInfo(peer.PID()))
			}

		case <-node.Stop:
			return nil
			//case m := <-p.Broadcast:
			//	//推送消息给客户端
			//	beego.Debug("推送消息给客户端:", m)
			//	p.broadcastMessage(m)
			//	break
		}
		log.Debug("Total Nodes:", len(node.onlinePeers))
	}

	return nil
}

//IsConnectPeer 是否连接某个节点
func (node *OWTPNode) IsConnectPeer(pid string) bool {
	peer := node.GetOnlinePeer(pid)
	if peer == nil {
		return false
	}
	return peer.IsConnected()
}

//ClosePeer 断开连接节点
func (node *OWTPNode) ClosePeer(pid string) {

	//检查是否已经连接服务
	peer := node.GetOnlinePeer(pid)
	if peer == nil {
		return
	}
	peer.close()
}

//Close 关闭节点
func (node *OWTPNode) Close() {

	for _, listener := range node.listeners {
		listener.Close()
	}
	//fmt.Printf("node close: %s\n", node.NodeID())
	//中断所有客户端连接
	for _, peer := range node.OnlinePeers() {
		peer.close()
		node.serveMux.ResetRequestQueue(peer.PID())
	}

	//通知停止运行
	node.Stop <- struct{}{}

	//node.client.close()
}

//ConnectAndCall 通过连接配置并直接请求，如果节点在线使用当前连接请求
func (node *OWTPNode) ConnectAndCall(
	pid string,
	config ConnectConfig,
	method string,
	params interface{},
	sync bool,
	reqFunc RequestFunc) error {

	if sync {
		peer, err := node.Connect(pid, config) //重新连接
		if err != nil {
			return err
		}

		return node.Call(peer.PID(), method, params, sync, reqFunc)
	} else {
		go func() {
			peer, err := node.Connect(pid, config) //重新连接
			if err != nil {
				return
			}

			node.Call(peer.PID(), method, params, sync, reqFunc)
		}()

		return nil
	}
}

//CallSync 同步请求
func (node *OWTPNode) CallSync(
	pid string,
	method string,
	params interface{},
) (*Response, error) {

	var (
		err      error
		respChan = make(chan Response, 1)
	)

	err = node.Call(pid, method, params, true, func(resp Response) {
		respChan <- resp
	})

	if err != nil {
		return nil, err
	}

	response := <-respChan
	return &response, nil
}

//Call 向对方节点进行调用
func (node *OWTPNode) Call(
	pid string,
	method string,
	params interface{},
	sync bool,
	reqFunc RequestFunc) error {

	var (
		err      error
		respChan = make(chan Response)
	)

	//检查是否已经连接服务
	peer := node.GetOnlinePeer(pid)
	if peer == nil {

		peerInfo := node.peerstore.PeerInfo(pid)

		peer, err = node.Connect(pid, peerInfo.Config) //重新连接
		if err != nil {
			return err
		}
	}

	//添加请求队列到Map，处理完成回调方法
	nonce := uint64(node.nonceGen.Generate().Int64())
	time := time.Now().Unix()

	//封装数据包
	packet := DataPacket{
		Method:    method,
		Req:       WSRequest,
		Nonce:     nonce,
		Timestamp: time,
		Data:      params,
		Version:   CurrentDataPacketVersion,
	}

	//如果开启了协商密码，添加协商密码参数
	if peer.auth() != nil && peer.auth().EnableKeyAgreement() {
		value := node.Peerstore().Get(peer.PID(), keyAgreementCipher)
		if value == nil {
			return fmt.Errorf("keyAgreement is enabled, but cipher is empty")
		}

		ka, ok := value.(*KeyAgreement)
		if !ok {
			return fmt.Errorf("keyAgreement is enabled, type error")
		}

		//加密数据
		err = node.encryptPacket(peer, &packet, ka.Key)
		if err != nil {
			return err
		}

		packet.SecretData = SecretData{
			PublicKeyInitiator:    ka.PublicKeyInitiator,
			TmpPublicKeyInitiator: ka.TmpPublicKeyInitiator,
			EncryptType:           ka.EncryptType,
		}

		if peer.IsHost() {
			//对方是服务端
			packet.SecretData.SA = ka.SA
		} else {
			//对方是客户端
			packet.SecretData.S2 = ka.S2
		}
	}

	if peer.auth() != nil && !peer.auth().GenerateSignature(&packet) {
		return errors.New("OWTP: authorization failed")
	}

	//添加请求到队列，异步或同步等待结果，应该在发送前就添加请求，如果发送失败，删除请求
	err = node.serveMux.AddRequest(peer, nonce, time, method, reqFunc, respChan, sync)
	if err != nil {
		return err
	}

	//向节点发送请求
	err = peer.send(packet)
	if err != nil {
		//发送失败移除请求
		node.serveMux.RemoveRequest(peer.PID(), nonce)
		return err
	}

	if sync {
		//等待返回
		result := <-respChan
		reqFunc(result)
	}

	return nil
}

//encryptPacket 加密数据包
func (node *OWTPNode) encryptPacket(peer Peer, packet *DataPacket, key string) error {

	//没有密钥不加密
	if len(key) == 0 {
		return nil
	}

	//var jsonByte []byte
	//
	//if encStr, ok := packet.Data.(string); ok {
	//	//无需转换为json
	//	jsonByte = []byte(encStr)
	//} else {
	//	enc, encErr := json.Marshal(packet.Data)
	//	if encErr != nil {
	//		return fmt.Errorf("json.Marshal data failed")
	//	}
	//	jsonByte = enc
	//}

	secretKey, decErr := base58.Decode(key)
	if decErr != nil {
		return decErr
	}

	chipErr := peer.auth().EncryptDataPacket(packet, secretKey)
	if chipErr != nil {
		return fmt.Errorf("encrypt data failed")
	}

	//fmt.Printf("plainText hex(%d): %s\n", len(enc), hex.EncodeToString(enc))
	//加载到授权中

	//chipText, chipErr := peer.auth().EncryptData(jsonByte, secretKey)
	//if chipErr != nil {
	//	return fmt.Errorf("encrypt data failed")
	//}

	//fmt.Printf("chipText hex(%d): %s\n", len(chipText), hex.EncodeToString(chipText))
	//packet.Data = string(chipText)

	return nil
}

//HandleFunc 绑定路由器方法
func (node *OWTPNode) HandleFunc(method string, handler HandlerFunc) {
	node.serveMux.HandleFunc(method, handler)
}

//HandlePrepareFunc 绑定准备前的处理方法
func (node *OWTPNode) HandlePrepareFunc(handler HandlerFunc) {
	node.serveMux.HandleFunc(PrepareMethod, handler)
}

//HandleFinishFunc 绑定结束后的处理方法
func (node *OWTPNode) HandleFinishFunc(handler HandlerFunc) {
	node.serveMux.HandleFunc(FinishMethod, handler)
}

//KeyAgreement 发起协商请求
//这是一个同步请求
func (node *OWTPNode) KeyAgreement(pid string, consultType string) error {
	//检查是否已经连接服务
	peer := node.GetOnlinePeer(pid)
	if peer == nil {
		return fmt.Errorf("remote peer is not connected")
	}
	return node.callKeyAgreement(peer, consultType)
}

//callKeyAgreement 发起协商计算
func (node *OWTPNode) callKeyAgreement(peer Peer, consultType string) error {

	var (
		err error
	)

	ka := &KeyAgreement{
		EncryptType: consultType,
	}

	//初始协商参数
	err = peer.auth().InitKeyAgreement(ka)
	if err != nil {
		return err
	}

	//保存请求协商密码的参数
	node.peerstore.Put(peer.PID(), keyAgreementCipher, ka)

	callErr := node.Call(
		peer.PID(),
		KeyAgreementMethod,
		nil,
		true,
		func(resp Response) {
			if resp.Status == StatusSuccess {

				err = nil
			} else {
				err = fmt.Errorf("keyAgreement failed, unexpected error: %s", resp.Msg)
			}
		})
	if callErr != nil {
		return callErr
	}

	return err
}

//keyAgreement 协商密钥
func (node *OWTPNode) keyAgreement(ctx *Context) {

	ctx.Response(nil, StatusSuccess, "success")

}

//wrapDataPacketForResponse 封装响应的数据包
func (node *OWTPNode) wrapDataPacketForResponse(peer Peer, ctx *Context) *DataPacket {

	packet := &DataPacket{}
	packet.Version = ctx.Version
	packet.Method = ctx.Method
	packet.Nonce = ctx.nonce

	//处理完请求，推送响应结果给服务端
	packet.Req = WSResponse
	packet.Data = ctx.Resp
	packet.Timestamp = time.Now().Unix()

	//如果开启了协商密码，添加协商密码参数
	if peer.auth().EnableKeyAgreement() {

		value := node.Peerstore().Get(peer.PID(), keyAgreementCipher)
		if value == nil {
			errMsg := fmt.Sprintf("keyAgreement is enabled, but pid: %s cipher is empty", peer.PID())
			packet.Data = responseError(errMsg, ErrKeyAgreementFailed)
			return packet
		}

		ka, ok := value.(*KeyAgreement)
		if !ok {
			packet.Data = responseError("keyAgreement is enabled, type error", ErrKeyAgreementFailed)
			return packet
		}

		//加密数据
		cryptErr := node.encryptPacket(peer, packet, ka.Key)
		if cryptErr != nil {
			log.Critical("OWTP: encryptData failed, unexpected err:", cryptErr)
			packet.Data = responseError("server encryptData failed", ErrInternalServerError)
			return packet
		}

		packet.SecretData = SecretData{
			PublicKeyResponder:    ka.PublicKeyResponder,
			TmpPublicKeyResponder: ka.TmpPublicKeyResponder,
			SB:                    ka.SB,
		}
	}

	return packet
}

//handleKeyAgreementForRequest 处理请求的协商密码，获得密钥
func (node *OWTPNode) handleKeyAgreementForRequest(peer Peer, packet *DataPacket) ([]byte, error) {
	var (
		//协商密码
		secretKey []byte
		ka        *KeyAgreement
	)

	//检查是否已经连接服务
	if peer == nil {
		return nil, fmt.Errorf("peer: %s is not connected. ", peer.PID())
	}

	//验证授权
	if peer.auth() != nil {

		//获取缓存中的协商密码组
		value := node.Peerstore().Get(peer.PID(), keyAgreementCipher)
		if value == nil {
			//不存在，重新构建协商密码组
			ka = &KeyAgreement{}

		} else {
			//协商密码存在
			k, ok := value.(*KeyAgreement)
			if !ok {
				return nil, fmt.Errorf("keyAgreement chipher is not saved")
			}
			ka = k

		}

		//发起方的请求协商密码的参数

		if peer.IsHost() {
			//对方是服务端，请求带SecretData.S2
			ka.S2 = packet.SecretData.S2
		} else {
			//对方是客户端，请求带SecretData.SA
			ka.SA = packet.SecretData.SA
		}

		ka.PublicKeyInitiator = packet.SecretData.PublicKeyInitiator
		ka.TmpPublicKeyInitiator = packet.SecretData.TmpPublicKeyInitiator

		//验证协商密码
		if !peer.auth().VerifyKeyAgreement(ka) {

			//协商不通过，需要重新协商
			log.Warning("keyAgreement is regenerating")

			//传入响应公私钥
			ka.PublicKeyResponder = base58.Encode(node.cert.PublicKeyBytes())
			ka.PrivateKeyResponder = base58.Encode(node.cert.PrivateKeyBytes())

			//请求协商
			err := peer.auth().RequestKeyAgreement(ka)
			if err != nil {
				return nil, err
			}

			//TODO:应该用一个临时密钥进行加密保存到缓存中
			node.peerstore.Put(peer.PID(), keyAgreementCipher, ka)

			//数据包配置响应方的协商密码参数
			packet.SecretData.PublicKeyResponder = ka.PublicKeyResponder
			packet.SecretData.TmpPublicKeyResponder = ka.TmpPublicKeyResponder
			packet.SecretData.SB = ka.SB

		}

		//加载到授权中
		secretKey, _ = base58.Decode(ka.Key)
		ka = nil

	}

	return secretKey, nil
}

//handleKeyAgreementForResponse 处理协商密码的响应方结果
func (node *OWTPNode) handleKeyAgreementForResponse(peer Peer, packet *DataPacket) ([]byte, error) {
	var (
		//协商密码
		secretKey []byte
	)

	//验证授权
	if peer.auth() != nil {

		//已开启协商密码
		if peer.auth().EnableKeyAgreement() {

			//获取协商密码组
			value := node.Peerstore().Get(peer.PID(), keyAgreementCipher)
			if value == nil {
				return nil, fmt.Errorf("keyAgreement is not inited")
			}

			ka, ok := value.(*KeyAgreement)
			if !ok {
				return nil, fmt.Errorf("keyAgreement chipher is not saved")
			}

			//本地还没完成生成协商密码
			//if len(ka.SA) == 0 || len(ka.Key) == 0  {
			if len(ka.SA) == 0 || len(ka.Key) == 0 || ka.SB != packet.SecretData.SB {

				//加载响应方的协商密码参数
				ka.PublicKeyResponder = packet.SecretData.PublicKeyResponder
				ka.TmpPublicKeyResponder = packet.SecretData.TmpPublicKeyResponder
				ka.SB = packet.SecretData.SB

				//计算协商密码
				err := peer.auth().ResponseKeyAgreement(ka)
				if err != nil {
					log.Errorf("ResponseKeyAgreement unexpected error: %v", err.Error())
					//协商失败，断开连接
					peer.close()
					return nil, err
				}

				//保存协商密码
				node.Peerstore().Put(peer.PID(), keyAgreementCipher, ka)

			}
			//协商密码
			secretKey, _ = base58.Decode(ka.Key)

			ka = nil

		}
	}

	return secretKey, nil
}

func (node *OWTPNode) GetValueForPeer(peer Peer, key string) interface{} {
	return node.Peerstore().Get(peer.PID(), key)
}

func (node *OWTPNode) PutValueForPeer(peer Peer, key string, val interface{}) error {
	return node.Peerstore().Put(peer.PID(), key, val)
}

//OnPeerOpen 节点连接成功
func (node *OWTPNode) OnPeerOpen(peer Peer) {
	node.Join <- peer
}

//OnPeerClose 节点关闭
func (node *OWTPNode) OnPeerClose(peer Peer, reason string) {
	node.Leave <- peer
}

//OnPeerNewDataPacketReceived 节点获取新数据包
func (node *OWTPNode) OnPeerNewDataPacketReceived(peer Peer, packet *DataPacket) {

	////重复攻击检查
	//if !node.checkNonceReplay(packet) {
	//	log.Error("nonce duplicate: ", packet)
	//	peer.send(*packet)
	//	return
	//}

	var (
		//协商密码
		secretKey []byte
		err       error
	)

	if packet.Req == WSRequest {

		//处理请求的协商密码，获得密钥
		secretKey, err = node.handleKeyAgreementForRequest(peer, packet)
		if err != nil {
			log.Critical("keyAgreement failed: ", packet)
			packet.Req = WSResponse
			packet.Data = responseError(err.Error(), ErrKeyAgreementFailed)
			packet.SecretData = SecretData{}
			peer.send(*packet) //发送验证失败结果
			return
		}

		//创建上下面指针，处理请求参数
		ctx := Context{
			Version:       packet.Version,
			PID:           peer.PID(),
			Req:           packet.Req,
			RemoteAddress: peer.RemoteAddr().String(),
			nonce:         packet.Nonce,
			Method:        packet.Method,
			peerstore:     node.Peerstore(),
			Peer:          peer,
		}

		//授权检查，只检查请求过来的签名
		verfifySigned := peer.auth().VerifySignature(packet)
		if !verfifySigned {
			//终止路由方法
			ctx.ResponseStopRun(nil, ErrUnauthorized, "verify signature failed, unauthorized")
		}

		cryptErr := peer.auth().DecryptDataPacket(packet, secretKey)
		if cryptErr != nil {
			log.Critical("OWTP: DecryptData failed, unexpected err:", cryptErr)
		}
		//解密后填充到输入参数
		ctx.inputs = packet.Data

		node.serveMux.ServeOWTP(peer.PID(), &ctx)

		retPacket := node.wrapDataPacketForResponse(peer, &ctx)

		peer.send(*retPacket)
	} else if packet.Req == WSResponse {

		//创建上下面指针，处理响应
		var resp Response

		ctx := Context{
			Version:       packet.Version,
			Req:           packet.Req,
			RemoteAddress: peer.RemoteAddr().String(),
			nonce:         packet.Nonce,
			inputs:        nil,
			Method:        packet.Method,
			//Resp:          resp,
			peerstore: node.Peerstore(),
			Peer:      peer,
		}

		//处理协商密码的响应方结果
		secretKey, err = node.handleKeyAgreementForResponse(peer, packet)
		if err != nil {
			log.Critical("keyAgreement failed: ", packet)
			resp = responseError(err.Error(), ErrKeyAgreementFailed)
			ctx.Resp = resp
			node.serveMux.ServeOWTP(peer.PID(), &ctx)
			return
		}

		//rawData, ok := packet.Data.(string)
		//if !ok {
		//	log.Critical("Data type error")
		//	resp = responseError("Data type error", ErrBadRequest)
		//	ctx.Resp = resp
		//	node.serveMux.ServeOWTP(peer.PID(), &ctx)
		//	return
		//}
		//log.Debug("rawData:", rawData)
		//decryptData, cryptErr := peer.auth().DecryptData([]byte(rawData), secretKey)
		//log.Debug("decryptData:", string(decryptData))

		cryptErr := peer.auth().DecryptDataPacket(packet, secretKey)
		if cryptErr != nil {
			log.Critical("OWTP: DecryptData failed")
			resp = responseError("Decrypt data error", ErrKeyAgreementFailed)
			ctx.Resp = resp
			node.serveMux.ServeOWTP(peer.PID(), &ctx)
			return
		}

		decryptData := packet.Data.([]byte)
		runErr := json.Unmarshal(decryptData, &resp)
		//runErr := mapstructure.Decode(decryptData, &resp)
		if runErr != nil {
			log.Error("Response decode error: ", runErr)
			resp = responseError("Response decode error", ErrBadRequest)
			ctx.Resp = resp
			node.serveMux.ServeOWTP(peer.PID(), &ctx)
			return
		}

		ctx.Resp = resp
		node.serveMux.ServeOWTP(peer.PID(), &ctx)

	}

}

// Peers 节点列表
func (node *OWTPNode) OnlinePeers() []Peer {
	node.mu.RLock()
	defer node.mu.RUnlock()
	peers := make([]Peer, 0)
	for _, peer := range node.onlinePeers {
		peers = append(peers, peer)
	}
	return peers
}

// GetOnlinePeer 获取当前在线的Peer
func (node *OWTPNode) GetOnlinePeer(id string) Peer {
	node.mu.RLock()
	defer node.mu.RUnlock()
	if node.onlinePeers == nil {
		return nil
	}
	return node.onlinePeers[id]
}

// AddOnlinePeer 添加在线节点
func (node *OWTPNode) AddOnlinePeer(peer Peer) {
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.onlinePeers == nil {
		node.onlinePeers = make(map[string]Peer)
	}
	node.onlinePeers[peer.PID()] = peer
}

//RemoveOfflinePeer 移除不在线的节点
func (node *OWTPNode) RemoveOfflinePeer(id string) {
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.onlinePeers == nil {
		return
	}
	delete(node.onlinePeers, id)
}

//GenerateRangeNum 生成范围内的随机整数
func GenerateRangeNum(min, max int) int {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	randNum := rand.Intn(max-min) + min
	return randNum
}

/*

//checkNonceReplay 检查nonce是否重放
func (node *OWTPNode) checkNonceReplay(data *DataPacket) bool {

	//检查
	status, errMsg := node.checkNonceReplayReason(data)

	if status != StatusSuccess {
		resp := Response{
			Status: status,
			Msg:    errMsg,
			Result: nil,
		}
		data.Req = WSResponse
		data.Data = resp
		data.Timestamp = time.Now().Unix()
		return false
	}

	return true

}

//checkNonceReplayReason 检查是否重放攻击
func (node *OWTPNode) checkNonceReplayReason(data *DataPacket) (uint64, string) {

	if data.Nonce == 0 || data.Timestamp == 0 {
		//没有nonce直接跳过
		return ErrReplayAttack, "no nonce"
	}

	//检查是否重放
	db, err := storm.Open(node.cacheFile)
	if err != nil {
		return ErrReplayAttack, "client system cache error"
	}
	defer db.close()

	var existPacket DataPacket
	db.One("Nonce", data.Nonce, &existPacket)
	if &existPacket != nil {
		return ErrReplayAttack, "this is a replay attack"
	}

	return StatusSuccess, ""
}
*/

//responseError 返回一个错误数据包
func responseError(err string, status uint64) Response {

	resp := Response{
		Status: status,
		Msg:    err,
		Result: nil,
	}

	return resp
}
