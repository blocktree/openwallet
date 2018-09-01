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

// owtp全称OpenWallet Transfer Protocol，OpenWallet的一种点对点的分布式私有通信协议。
package owtp

import (
	"fmt"
	"github.com/blocktree/OpenWallet/log"
	"github.com/bwmarrin/snowflake"
	"github.com/mitchellh/mapstructure"
	"math/rand"
	"sync"
	"time"
	"strings"
	"errors"
)

type ConnectType int

const (

	//成功标识
	StatusSuccess uint64 = 200

	//客户端请求错误
	ErrBadRequest uint64 = 400
	//网络断开
	ErrNetworkDisconnected uint64 = 401
	//找不到方法
	ErrNotFoundMethod uint64 = 404
	//重放攻击
	ErrReplayAttack uint64 = 409
	//重放攻击
	ErrRequestTimeout uint64 = 408
	//服务器错误
	ErrInternalServerError uint64 = 500
	//请求与响应的方法不一致
	ErrResponseMethodDiffer uint64 = 501
	//协商失败
	ErrKeyAgreementFailed uint64 = 502

	//60X: 自定义错误
	ErrCustomError uint64 = 600

	Websocket string = "ws"

	MQ string = "mq"

	KeyAgreementMethod = "internal_keyAgreement"
)


//节点主配置 作为json解析工具
type MainConfig struct {
	Address     string
	ConnectType int
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
	listening bool
	//节点运行中
	running bool
	//读写锁
	mu sync.RWMutex
	//关闭连接时的回调
	disconnectHandler func(n *OWTPNode, peerInfo PeerInfo)
	//连接时的回调
	connectHandler func(n *OWTPNode, peerInfo PeerInfo)
	//节点存储器
	peerstore *owtpPeerstore
	//服务监听器
	listener Listener
	//授权证书
	cert Certificate
	//Broadcast   chan BroadcastMessage
	Join  chan Peer
	Leave chan Peer
	Stop  chan struct{}

	//通道的读写缓存大小
	ReadBufferSize, WriteBufferSize int
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
	node := NewOWTPNode(cert, 0, 0)
	return node
}

//NewOWTPNode 创建OWTP协议节点
func NewOWTPNode(cert Certificate, readBufferSize, writeBufferSize int) *OWTPNode {

	node := &OWTPNode{}
	node.nonceGen, _ = snowflake.NewNode(1)
	node.serveMux = NewServeMux(120)
	node.cert = cert
	node.peerstore = NewPeerstore()
	node.ReadBufferSize = readBufferSize
	node.WriteBufferSize = writeBufferSize
	node.Join = make(chan Peer)
	node.Leave = make(chan Peer)
	node.Stop = make(chan struct{})

	//内部配置一个协商密码处理过程
	node.HandleFunc(KeyAgreementMethod, node.keyAgreement)

	//马上执行
	go node.Run()

	return node
}

//NodeID 节点的ID
func (node *OWTPNode) NodeID() string {
	return node.cert.ID()
}

//Peerstore 节点存储器
func (node *OWTPNode) Peerstore() Peerstore {
	return node.peerstore
}

//Listen 监听TCP地址
func (node *OWTPNode) Listen(addr string) error {

	if node.listening {
		return fmt.Errorf("the node is listening, please close listener first")
	}

	l, err := ListenAddr(addr, node)
	if err != nil {
		return err
	}
	node.listener = l

	go func(listener Listener) {
		for {
			peer, err := listener.Accept()
			if err != nil {
				return
			}
			node.Join <- peer
		}
	}(l)

	node.listening = true

	return nil
}

//listening 是否监听中
func (node *OWTPNode) Listening() bool {
	return node.listening
}


//Connect 建立长连接
func (node *OWTPNode) Connect( pid string,config map[string]string) error {

	_, err := node.connect( pid,config)

	return err
}

//connect 建立长连接，内部调用
func (node *OWTPNode) connect( pid string,config map[string]string) (Peer, error) {


	if config == nil{
		return nil, fmt.Errorf("config  is nil")
	}

	addr := config["address"]

	if len(addr) == 0 {
		return nil, fmt.Errorf("address must contain by config")
	}

	auth, err := NewOWTPAuthWithCertificate(node.cert)

	//发起协商密钥
	err = auth.InitKeyAgreement()
	if err != nil {
		return nil, err
	}

	//链接类型
	connectType := config["connectType"]

	if len(connectType) == 0 {
		return nil, fmt.Errorf("connectType must contain by config")
	}

	//websocket类型
	if connectType == Websocket{

		url := "ws://" + strings.TrimSuffix(addr,"/") + "/" + pid

		//建立链接，记录默认的客户端
		client, err := Dial(pid, url, node, auth.AuthHeader(), node.ReadBufferSize, node.WriteBufferSize)
		if err != nil {
			return nil, err
		}
		//设置授权规则
		client.auth = auth
		//设置配置
		client.config = config
		return client, nil
	}

	//MQ类型
	if connectType == MQ{


		url := "amqp://admin:admin@" + strings.TrimSuffix(addr,"/") + "/"

		//建立链接，记录默认的客户端
		client, err := MQDial(pid, url,node)
		if err != nil {
			return nil, err
		}
		//设置授权规则
		client.auth = auth
		//设置配置
		client.config = config
		return client, nil
	}

	return nil,errors.New("connectType can't found! ")
}


//SetCloseHandler 设置关闭连接时的回调
func (node *OWTPNode) SetOpenHandler(h func(n *OWTPNode, peer PeerInfo)) {
	node.connectHandler = h
}

//SetCloseHandler 设置关闭连接时的回调
func (node *OWTPNode) SetCloseHandler(h func(n *OWTPNode, peer PeerInfo)) {
	node.disconnectHandler = h
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
			log.Info("Node Join:", peer.PID())
			log.Info("Node IP:", peer.RemoteAddr().String())
			node.peerstore.AddOnlinePeer(peer)
			node.peerstore.SavePeer(peer.PID(), peer)
			//加入后打开数据流通道
			if err := peer.OpenPipe(); err != nil {
				log.Error("peer:", peer.PID(), "open pipe failed")
				continue
			}

			//非我方主动连接的，并且开启授权，向其发起密钥协商
			if !peer.IsHost() && peer.Auth().EnableAuth() {
				//加入成功后，进行密码协商
				node.callKeyAgreement(peer)
			}

			if node.connectHandler != nil {
				go node.connectHandler(node, node.Peerstore().PeerInfo(peer.PID()))
			}

		case peer := <-node.Leave:
			//客户端离开
			log.Info("Node Leave:", peer.PID())
			node.serveMux.ResetRequestQueue(peer.PID())
			node.peerstore.RemoveOfflinePeer(peer.PID())

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
		log.Info("Total Nodes:", len(node.peerstore.onlinePeers))
	}

	return nil
}

//IsConnectPeer 是否连接某个节点
func (node *OWTPNode) IsConnectPeer(pid string) bool {
	peer := node.peerstore.GetOnlinePeer(pid)
	if peer == nil {
		return false
	}
	return peer.IsConnected()
}

//ClosePeer 断开连接节点
func (node *OWTPNode) ClosePeer(pid string) {

	//检查是否已经连接服务
	peer := node.peerstore.GetOnlinePeer(pid)
	if peer == nil {
		return
	}
	peer.Close()
}

//Close 关闭节点
func (node *OWTPNode) Close() {

	if node.listener != nil {
		node.listener.Close()
		node.mu.Lock()
		node.listening = false
		node.mu.Unlock()
	}

	//中断所有客户端连接
	for _, peer := range node.peerstore.OnlinePeers() {
		peer.Close()
		node.serveMux.ResetRequestQueue(peer.PID())
	}

	//通知停止运行
	node.Stop <- struct{}{}

	//node.client.Close()
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
	peer := node.peerstore.GetOnlinePeer(pid)
	if peer == nil {
		newPeer := node.peerstore.GetPeer(pid)
		if newPeer == nil{
			return fmt.Errorf("the peer: %s is not in peer book", pid)
		}
		peerAddr := newPeer.RemoteAddr().String()
		if peerAddr == "" {
			return fmt.Errorf("the peer: %s is not in address book", pid)
		}

		peerInfo := node.peerstore.PeerInfo(pid)

		peer, err = node.connect(pid,peerInfo.Config) //重新连接
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
	}

	//向节点发送请求
	err = peer.Send(packet)
	if err != nil {
		return err
	}

	//添加请求到队列，异步或同步等待结果
	node.serveMux.AddRequest(peer.PID(), nonce, time, method, reqFunc, respChan, sync)
	if sync {
		//等待返回
		result := <-respChan
		reqFunc(result)
	}

	return nil
}

//HandleFunc 绑定路由器方法
func (node *OWTPNode) HandleFunc(method string, handler HandlerFunc) {
	node.serveMux.HandleFunc(method, handler)
}

//callKeyAgreement 发起协商计算
func (node *OWTPNode) callKeyAgreement(peer Peer) error {

	inputs := map[string]interface{}{
		"localPrivateKey": node.cert.privateKeyBytes,
		"localPublicKey":  node.cert.publicKeyBytes,
	}

	//计算密钥，并请求协商
	params, err := peer.Auth().RequestKeyAgreement(inputs)
	if err != nil {
		return err
	}

	err = node.Call(
		peer.PID(),
		KeyAgreementMethod,
		params,
		true,
		func(resp Response) {
			if resp.Status == StatusSuccess {
				sa := resp.JsonData().Get("sa").String()
				finalErr := peer.Auth().VerifyKeyAgreement(sa)
				if finalErr != nil {
					//协商失败，断开连接
					peer.Close()
				}
			}
		})
	return err
}

//keyAgreement 协商密钥
func (node *OWTPNode) keyAgreement(ctx *Context) {

	//检查是否已经连接服务
	peer := node.peerstore.GetOnlinePeer(ctx.PID)
	if peer == nil {
		responseError(fmt.Sprintf("peer: %s is not connected.", ctx.PID), ErrKeyAgreementFailed)
		return
	}

	pubkeyResponder := ctx.Params().Get("pubkeyResponder").String()
	tmpPubkeyResponder := ctx.Params().Get("tmpPubkeyResponder").String()
	sb := ctx.Params().Get("sb").String()

	inputs := map[string]interface{}{
		"remotePublicKey":    pubkeyResponder,
		"remoteTmpPublicKey": tmpPubkeyResponder,
		"sb":                 sb,
	}

	result, err := peer.Auth().ResponseKeyAgreement(inputs)
	if err != nil {
		responseError(err.Error(), ErrKeyAgreementFailed)
		return
	}

	ctx.Response(result, StatusSuccess, "success")
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
	//	peer.Send(*packet)
	//	return
	//}

	//验证授权
	if peer.Auth() != nil && peer.Auth().EnableAuth() {

		//授权检查
		if !peer.Auth().VerifySignature(packet) {
			log.Error("auth failed: ", packet)
			peer.Send(*packet) //发送验证失败结果
			return
		}
	}

	if packet.Req == WSRequest {

		//创建上下面指针，处理请求参数
		ctx := Context{PID: peer.PID(), Req: packet.Req, nonce: packet.Nonce, inputs: packet.Data, Method: packet.Method}

		node.serveMux.ServeOWTP(peer.PID(), &ctx)

		//处理完请求，推送响应结果给服务端
		packet.Req = WSResponse
		packet.Data = ctx.Resp
		peer.Send(*packet)
	} else if packet.Req == WSResponse {

		//创建上下面指针，处理响应
		var resp Response
		runErr := mapstructure.Decode(packet.Data, &resp)
		if runErr != nil {
			log.Error("Response decode error: ", runErr)
			return
		}

		ctx := Context{Req: packet.Req, nonce: packet.Nonce, inputs: nil, Method: packet.Method, Resp: resp}

		node.serveMux.ServeOWTP(peer.PID(), &ctx)

	}

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
	defer db.Close()

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
