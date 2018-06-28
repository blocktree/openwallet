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
	"github.com/bwmarrin/snowflake"
	"math/rand"
	"sync"
	"time"
)

const (
	//找不到方法
	ErrNotFoundMethod uint64 = 100
	//请求与响应的方法不一致
	ErrResponseMethodDiffer uint64 = 101
	//重放攻击
	ErrReplayAttack uint64 = 102
	//成功标识
	StatusSuccess uint64 = 0
)

//OWTPNode 实现OWTP协议的节点
type OWTPNode struct {
	url string
	//客户端
	client *Client
	//nonce生成器
	nonceGen *snowflake.Node
	//默认路由
	serveMux *ServeMux
	//是否已经连接
	isConnected bool
	//读写锁
	mu sync.RWMutex
	//关闭连接时的回调
	disconnectHandler func(n *OWTPNode)
	//授权
	Auth Authorization
}

//NewOWTPNode 创建OWTP协议节点
func NewOWTPNode(nodeID int64, url string, auth Authorization) *OWTPNode {

	node := &OWTPNode{}
	node.url = url
	node.nonceGen, _ = snowflake.NewNode(nodeID)
	node.serveMux = &ServeMux{}
	node.Auth = auth
	return node
}

//Connect 建立长连接
func (node *OWTPNode) Connect() error {

	//已经连接过了
	if node.client != nil && node.isConnected {
		return nil
	}

	//建立链接，记录默认的客户端
	client, err := Dial(node.url, node.serveMux, node.Auth)
	if err != nil {
		return err
	}

	node.mu.Lock()
	//设置一个全局的webscoket
	node.client = client
	//node.client.SetCloseHandler(node.disconnect)
	node.isConnected = true
	node.serveMux.ResetQueue()
	node.mu.Unlock()

	go node.run()

	return nil
}

//SetCloseHandler 设置关闭连接时的回调
func (node *OWTPNode) SetCloseHandler(h func(n *OWTPNode)) {
	node.disconnectHandler = h
}

//run 运行监听连接关闭
func (node *OWTPNode) run() {
	for  {
		select {
		case <- node.client.close:
			node.isConnected = false
			node.serveMux.ResetQueue()
			node.disconnectHandler(node)
			return
		}
	}
}

//disconnect 断开连接实现回调
func (node *OWTPNode) disconnect(code int, text string) error {
	node.isConnected = false
	node.serveMux.ResetQueue()
	node.disconnectHandler(node)
	return nil
}

//Close 关闭节点
func (node *OWTPNode) Close() {
	//中断客户端连接
	node.client.Close()
	node.isConnected = false
	node.serveMux.ResetQueue()
}

//Call 向对方节点进行调用
func (node *OWTPNode) Call(
	method string,
	params interface{},
	reqFunc RequestFunc,
	sync bool) error {

	var (
		err      error
		respChan = make(chan Response, 0)
	)

	//检查是否已经连接服务
	if !node.isConnected {
		err = node.Connect() //重新连接
		if err != nil {
			return err
		}
	}

	//添加请求队列到Map，处理完成回调方法
	nonce := uint64(node.nonceGen.Generate().Int64())

	//封装数据包
	packet := DataPacket{
		Method:    method,
		Req:       WSRequest,
		Nonce:     nonce,
		Timestamp: time.Now().Unix(),
		Data:      params,
	}

	//发送请求
	err = node.client.Send(packet)
	if err != nil {
		return err
	}

	//添加请求到队列，异步或同步等待结果
	node.serveMux.AddRequest(nonce, method, reqFunc, respChan, sync)
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

//Disconnect 断开连接后回调
func (node *OWTPNode) Disconnect(handler func()) {

}

//GenerateRangeNum 生成范围内的随机整数
func GenerateRangeNum(min, max int) int {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	randNum := rand.Intn(max-min) + min
	return randNum
}
