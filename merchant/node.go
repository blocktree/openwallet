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

package merchant

import (
	"errors"
	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/owtp"
	"github.com/blocktree/OpenWallet/timer"
	"github.com/blocktree/OpenWallet/log"
	"sync"
	"time"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/OpenWallet/assets"
)

var (
	PeriodOfTask = 5 * time.Second
	//通道的读写缓存大小
	ReadBufferSize = 1024 * 1024
	WriteBufferSize = 1024 * 1024
)

//商户节点
type MerchantNode struct {

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
}

func NewMerchantNode(config NodeConfig) (*MerchantNode, error) {

	m := MerchantNode{}
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

	cert, err := owtp.NewCertificate(config.PrivateKey, "")

	if err != nil {
		return nil, err
	}

	//创建节点，连接商户
	node := owtp.NewOWTPNode(cert, ReadBufferSize, WriteBufferSize)
	m.Node = node
	m.Config = config

	//断开连接后，重新连接
	m.Node.SetCloseHandler(func(n *owtp.OWTPNode, peer owtp.Peer) {
		log.Info("merchantNode disconnect.")
		m.disconnected <- struct{}{}
	})

	m.isReconnect = true
	m.ReconnectWait = 10
	m.getAddressesCh = make(chan AddressVersion, 5)

	//设置路由
	m.setupRouter()

	return &m, nil
}

//resetSubscriptions 重置订阅表
func (m *MerchantNode) resetSubscriptions(news []*Subscription) {

	//清除旧订阅
	for _, s := range m.subscriptions {

		am := assets.GetMerchantAssets(s.Coin)
		if am == nil {
			continue
		}

		//移除旧的订阅观察者
		am.RemoveMerchantObserverForBlockScan(m)

		//TODO: 重新订阅扫描地址
	}

	m.mu.Lock()
	m.subscriptions = nil
	m.subscriptions = news
	m.mu.Unlock()

	//加载新订阅
	for _, s := range m.subscriptions {

		am := assets.GetMerchantAssets(s.Coin)
		if am == nil {
			continue
		}

		wallet, err := m.GetMerchantWalletByID(s.WalletID)
		if err != nil {
			continue
		}

		//加入到区块链观测者表单
		am.AddMerchantObserverForBlockScan(m, wallet)
	}
}

//OpenDB 访问数据库
func (m *MerchantNode) OpenDB() (*storm.DB, error) {
	return storm.Open(m.Config.CacheFile)
}

//SaveToDB 保存到商户数据库
func (m *MerchantNode) SaveToDB(data interface{})  error {
	db, err := m.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Save(data)
}

//GetMerchantWalletByID 获取商户钱包
func (m *MerchantNode) GetMerchantWalletByID(walletID string) (*openwallet.Wallet, error) {

	db, err := m.OpenDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var wallet openwallet.Wallet
	err = db.One("WalletID", walletID, &wallet)
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}


//GetMerchantWalletList 获取商户钱包列表
func (m *MerchantNode) GetMerchantWalletList() ([]*openwallet.Wallet, error) {

	db, err := m.OpenDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var wallets []*openwallet.Wallet
	err = db.All(&wallets)
	if err != nil {
		return nil, err
	}

	return wallets, nil
}


//GetMerchantAccountByID 获取商户资产账户
func (m *MerchantNode) GetMerchantAccountByID(accountID string) (*openwallet.AssetsAccount, error) {

	db, err := m.OpenDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var account openwallet.AssetsAccount
	err = db.One("AccountID", accountID, &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}


//GetMerchantAccountList 获取商户资产账户列表
func (m *MerchantNode) GetMerchantAccountList(coin string) ([]*openwallet.AssetsAccount, error) {

	db, err := m.OpenDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var acouunts []*openwallet.AssetsAccount
	err = db.Find("Symbol", coin, &acouunts)
	if err != nil {
		return nil, err
	}

	return acouunts, nil
}

//GetMerchantWalletConfig 获取商户钱包配置信息
func (m *MerchantNode) GetMerchantWalletConfig(coin string, walletID string) (*openwallet.WalletConfig, error) {

	db, err := m.OpenDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var wallet openwallet.WalletConfig
	err = db.One("Key", coin + "_" + walletID, &wallet)
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

//Run 运行商户节点管理
func (m *MerchantNode) Run() error {

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

	//节点运行时
	for {
		select {
		case <-m.reconnect:
			//重新连接
			log.Info("Connecting to", m.Config.MerchantNodeURL)
			err = m.Node.Connect(m.Config.MerchantNodeURL, m.Config.NodeID)
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
func (m *MerchantNode) IsConnected() error {

	if m.Node == nil {
		return ErrMerchantNodeDisconnected
	}

	if !m.Node.IsConnectPeer(m.Config.NodeID) {
		return ErrMerchantNodeDisconnected
	}
	return nil
}

//Stop 停止运行
func (m *MerchantNode) Stop() {
	close(m.disconnected)
}


//StartTimerTask 启动定时任务
func (m *MerchantNode) StartTimerTask() {

	log.Info("Merchant timer task start...")

	//启动订阅地址任务
	m.runSubscribeAddressTask()
}

//StopTimerTask 停止定时任务
func (m *MerchantNode) StopTimerTask() {

	log.Info("Merchant timer task stop...")

	//停止地址订阅任务
	m.subscribeAddressTask.Pause()
	//停止交易记录订阅任务

}

//DeleteAddressVersion 删除地址版本
func (m *MerchantNode) DeleteAddressVersion(a *AddressVersion) error {

	db, err := m.OpenDB()
	if err != nil {
		return err
	}

	db.DeleteStruct(a)
	db.Close()

	return nil
}

/********** 商户服务相关方法【主动】 **********/
