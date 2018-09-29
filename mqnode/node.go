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

package mqnode

import (
	"errors"
	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/owtp"
	"github.com/blocktree/OpenWallet/timer"
	"github.com/blocktree/OpenWallet/log"
	"sync"
	"time"
	"github.com/blocktree/OpenWallet/manager"
	"github.com/blocktree/OpenWallet/openwallet"
	"fmt"
	"encoding/json"
	"github.com/tidwall/gjson"
)

var (
	PeriodOfTask = 5 * time.Second
	//通道的读写缓存大小
	ReadBufferSize = 1024 * 1024
	WriteBufferSize = 1024 * 1024
)

//商户节点
type BitBankNode struct {

	//读写锁，用于处理订阅更新和根据订阅获取数据的同步控制
	mu sync.RWMutex
	//节点配置
	Config NodeConfig
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
	//获取地址通道
	getAddressesCh chan AddressVersion
	//订阅列表
	subscriptions []*Subscription
	//订阅地址任务
	subscribeAddressTask *timer.TaskTimer
	//定时器任务
	TaskTimers map[string]*timer.TaskTimer

	manager *manager.WalletManager
}

func NewBitNodeNode(config NodeConfig) (*BitBankNode, error) {

	m := BitBankNode{}
	if len(config.MerchantNodeURL) == 0 {
		return nil, errors.New("merchant node url is not configed! ")
	}

	//授权配置
	//auth, err := owtp.NewOWTPAuth(
	//	config.NodeKey,
	//	config.PublicKey,
	//	config.PrivateKey,
	//	true,
	//	config.CacheFile,
	//)

	cert, err := owtp.NewCertificate(config.LocalPrivateKey, "")

	if err != nil {
		return nil, err
	}

	//创建节点，连接商户
	node := owtp.NewOWTPNode(cert, ReadBufferSize, WriteBufferSize)
	m.Node = node
	m.Config = config

	//断开连接后，重新连接
	m.Node.SetCloseHandler(func(n *owtp.OWTPNode, peer owtp.PeerInfo) {
		log.Info("merchantNode disconnect.")
		m.disconnected <- struct{}{}
	})

	m.Node.SetOpenHandler(func(n *owtp.OWTPNode, peer owtp.PeerInfo) {
		log.Info("merchantNode connected.")
	})

	m.isReconnect = true
	m.ReconnectWait = 10
	m.getAddressesCh = make(chan AddressVersion, 5)

	//设置路由
	m.setupRouter()

	return &m, nil
}


//OpenDB 访问数据库
func (m *BitBankNode) OpenDB() (*storm.DB, error) {
	return storm.Open(m.Config.CacheFile)
}

//SaveToDB 保存到商户数据库
func (m *BitBankNode) SaveToDB(data interface{})  error {
	db, err := m.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Save(data)
}

//Run 运行商户节点管理
func (m *BitBankNode) Run() error {

	var (
		err error
	)

	defer func() {
		close(m.reconnect)
		close(m.disconnected)
	}()

	m.reconnect = make(chan bool, 1)
	m.disconnected = make(chan struct{}, 1)

	//启动连接
	m.reconnect <- true

	log.Info("Merchant node running now...")
	go func(){
		var (
			endRunning = make(chan bool, 1)
		)

		m.manager.AddObserver(m)

		<-endRunning
	}()

	//节点运行时
	for {
		select {
		case <-m.reconnect:
			//重新连接
			log.Info("Connecting to", m.Config.MerchantNodeURL)
			config := map[string]string{
				"address": m.Config.MerchantNodeURL,
				"connectType": m.Config.ConnectType,
				"exchange": m.Config.Exchange,
				"queueName":m.Config.QueueName,
				"receiveQueueName":m.Config.ReceiveQueueName,
				"account":m.Config.Account,
				"password":m.Config.Password,
			}
			err = m.Node.Connect(m.Config.MerchantNodeID, config)
			if err != nil {
				log.Error("Connect merchant node faild unexpected error:", err)
				m.disconnected <- struct{}{}
			} else {
				log.Info("Connect merchant node successfully. \n")
			}

			//启动定时任务
			m.StartTimerTask()

		case <-m.disconnected:

			//停止定时任务
			m.StopTimerTask()

			if m.isReconnect {
				//重新连接，前等待
				log.Info("Reconnect after", m.ReconnectWait, "seconds...")
				time.Sleep(m.ReconnectWait * time.Second)
				m.reconnect <- true
			} else {
				//退出
				break
			}
		}
	}

	return nil
}

//IsConnected 检查商户节点是否连接
func (m *BitBankNode) IsConnected() error {

	if m.Node == nil {
		return ErrMerchantNodeDisconnected
	}

	if !m.Node.IsConnectPeer(m.Config.MerchantNodeID) {
		return ErrMerchantNodeDisconnected
	}
	return nil
}

//Stop 停止运行
func (m *BitBankNode) Stop() {
	close(m.disconnected)
}


//StartTimerTask 启动定时任务
func (m *BitBankNode) StartTimerTask() {

	log.Info("Merchant timer task start...")

}

//StopTimerTask 停止定时任务
func (m *BitBankNode) StopTimerTask() {

	log.Info("Merchant timer task stop...")

	//停止地址订阅任务
	//m.subscribeAddressTask.Pause()
	//停止交易记录订阅任务

}

//DeleteAddressVersion 删除地址版本
func (m *BitBankNode) DeleteAddressVersion(a *AddressVersion) error {

	db, err := m.OpenDB()
	if err != nil {
		return err
	}

	db.DeleteStruct(a)
	db.Close()

	return nil
}

/********** 商户服务相关方法【主动】 **********/



//BlockScanNotify 新区块扫描完成通知
func (bitBankNode *BitBankNode) BlockScanNotify(header *openwallet.BlockHeader) error {
	log.Info("header:", header)
	bitBankNode.CallTarget("pushNewBlock",header)
	return nil
}

//BlockTxExtractDataNotify 区块提取结果通知
func (bitBankNode *BitBankNode) BlockTxExtractDataNotify(account *openwallet.AssetsAccount, data *openwallet.TxExtractData) error {
	log.Info("account:", account)
	log.Info("data:", data)
	if data.Transaction != nil{
		result := map[string]interface{}{
			"appID":"",
			"walletID":account.WalletID,
			"accountID":account.AccountID,
			"dataType":2,
			"content":data.Transaction,
		}
		inbs, err := json.Marshal(result)
		if err == nil {
			log.Error("result:",gjson.ParseBytes(inbs))
		}
		bitBankNode.CallTarget("pushNotifications",result)
	}
	if data.TxInputs != nil{
		result := map[string]interface{}{
			"appID":"",
			"walletID":account.WalletID,
			"accountID":account.AccountID,
			"dataType":4,
			"content":data.TxInputs,
		}
		inbs, err := json.Marshal(result)
		if err == nil {
			log.Error("result:",gjson.ParseBytes(inbs))
		}
		bitBankNode.CallTarget("pushNotifications",result)
	}

	if data.TxOutputs != nil{
		result := map[string]interface{}{
			"appID":"",
			"walletID":account.WalletID,
			"accountID":account.AccountID,
			"dataType":3,
			"content":data.TxInputs,
		}
		inbs, err := json.Marshal(result)
		if err == nil {
			log.Error("result:",gjson.ParseBytes(inbs))
		}
		bitBankNode.CallTarget("pushNotifications",result)
	}


	result := map[string]interface{}{
		"appID":"",
		"walletID":account.WalletID,
		"accountID":account.AccountID,
		"dataType":1,
		"content":account.Balance,
	}
	inbs, err := json.Marshal(result)
	if err == nil {
		log.Error("result:",gjson.ParseBytes(inbs))
	}
	bitBankNode.CallTarget("pushNotifications",result)
	return nil
}

func (bitBankNode *BitBankNode) CallTarget(method string,params interface{}){
	bitBankNode.Node.Call(bitBankNode.Config.MerchantNodeID, method, params, false, func(resp owtp.Response) {
		fmt.Printf("BitBankNode call pushNotifications, params: %s,result: %s\n", params,resp.JsonData())
	})
}