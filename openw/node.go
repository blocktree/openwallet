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

package openw

import (
	"time"
	"github.com/blocktree/OpenWallet/owtp"
	"github.com/blocktree/OpenWallet/timer"
	"sync"
	"github.com/blocktree/OpenWallet/log"
)

var (
	PeriodOfTask = 5 * time.Second
	//通道的读写缓存大小
	ReadBufferSize = 1024 * 1024
	WriteBufferSize = 1024 * 1024
)

//Node openw钱包服务节点
type Node struct {

	//读写锁，用于处理订阅更新和根据订阅获取数据的同步控制
	mu sync.RWMutex
	//节点配置
	//Config NodeConfig
	//商户节点
	Node *owtp.OWTPNode
	//连接状态通道
	reconnect chan bool
	//断开状态通道
	disconnected chan struct{}
	//是否重连
	isReconnect bool
	//重连时的等待时间
	ReconnectWait time.Duration
	//订阅地址任务
	subscribeAddressTask *timer.TaskTimer
	//定时器任务
	TaskTimers map[string]*timer.TaskTimer
}

func NewNode(config interface{}) (*Node, error) {

	m := Node{}

	return &m, nil
}

//Run 运行商户节点管理
func (m *Node) Run() error {

	//var (
	//	err error
	//)

	defer func() {
		close(m.reconnect)
		close(m.disconnected)
	}()

	m.reconnect = make(chan bool, 1)
	m.disconnected = make(chan struct{}, 1)

	//启动连接
	m.reconnect <- true

	log.Info("Merchant node running now...")

	//节点运行时
	for {
		select {
		case <-m.reconnect:

		case <-m.disconnected:

		}
	}

	return nil
}

//IsConnected 检查节点是否连接
func (m *Node) IsConnected() error {
	
	return nil
}

//Stop 停止运行
func (m *Node) Stop() {

}
