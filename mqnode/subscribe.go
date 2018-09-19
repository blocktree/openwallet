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
	"github.com/blocktree/OpenWallet/assets"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/OpenWallet/owtp"
	"github.com/blocktree/OpenWallet/timer"
	"time"
)

//GetChargeAddressVersion 获取要订阅的地址版本信息
func (m *BitBankNode) GetChargeAddressVersion() error {

	var (
		//err  error
		subs = make([]*Subscription, 0)
	)

	//检查是否连接
	if m.Node == nil {
		return ErrMerchantNodeDisconnected
	}

	//db, err := m.OpenDB()
	//if err != nil {
	//	return err
	//}

	m.mu.RLock()
	for _, s := range m.subscriptions {
		if s.Type == SubscribeTypeCharge {
			subs = append(subs, s)
		}
	}
	m.mu.RUnlock()

	//err = db.Find("type", SubscribeTypeCharge, &subs)
	//if err != nil {
	//	return err
	//}
	//
	//db.Close()

	for _, sub := range subs {

		//| 参数名称 | 类型   | 是否可空 | 描述     |
		//|----------|--------|----------|----------|
		//| coin     | string | 是       | 币名     |
		//| walletID | string | 是       | 钱包ID   |

		params := struct {
			Coin     string `json:"coin"`
			WalletID string `json:"walletID"`
		}{sub.Symbol, sub.WalletID}

		//获取订阅的地址版本
		GetChargeAddressVersion(m.Node, m.Config.MerchantNodeID, params,
			true,
			func(addressVer *AddressVersion, status uint64, msg string) {

				if addressVer != nil {
					innerdb, err := m.OpenDB()
					if err != nil {
						return
					}
					defer innerdb.Close()
					var oldVersion AddressVersion
					err = innerdb.One("Key", addressVer.Key, &oldVersion)

					if addressVer.Version > oldVersion.Version || err != nil {
						m.getAddressesCh <- *addressVer

						log.Info("get new address version:", addressVer.Version)
						log.Info("get new address total:", addressVer.Total)

						//更新记录
						innerdb.Save(addressVer)
					}

				}

			})
	}

	return nil
}

//GetChargeAddress 获取地址
func (m *BitBankNode) getChargeAddress() error {

	var (
		err   error
		limit = uint64(1000)
	)

	////检查是否连接
	if m.Node == nil {
		return ErrMerchantNodeDisconnected
	}

	//log.Printf("getChargeAddress running...\n")

	for {

		select {
		//接收通道发送的地址版本
		case v := <-m.getAddressesCh:

			getCount := uint64(0)

			//log.Printf("get address version: %v", v)
			for i := uint64(0); i < v.Total; i = i + limit {

				params := struct {
					Coin     string `json:"coin"`
					WalletID string `json:"walletID"`
					Offset   uint64 `json:"offset"`
					Limit    uint64 `json:"limit"`
				}{v.Coin, v.WalletID, i, limit}

				err = GetChargeAddress(m.Node, m.Config.MerchantNodeID, params,
					true,
					func(addrs []*openwallet.Address, status uint64, msg string) {

						if status == owtp.StatusSuccess {
							//log.Printf("GetMerchantWalletByID WalletID: %v\n", v.WalletID)
							wallet, blockErr := m.GetMerchantWalletByID(v.WalletID)
							if err != nil {
								log.Error("GetMerchantWalletByID unexpected error:", blockErr)
								return
							}

							//导入到每个币种的数据库
							mer := assets.GetMerchantAssets(v.Coin)
							//log.Printf("mer = %v", mer)
							if mer != nil {
								log.Debug("address count =", len(addrs))
								blockErr = mer.ImportMerchantAddress(wallet, wallet.SingleAssetsAccount(v.Coin), addrs)
								if blockErr != nil {
									log.Error("ImportMerchantAddress unexpected error:", blockErr)
								}

							}
							getCount = getCount + uint64(len(addrs))
						}
					})
				if err != nil {
					log.Error("GetChargeAddress unexpected error:", err)
					continue
				}

			}

			//检查地址条数是否完整
			if getCount != v.Total {
				// 扔到通道中，重新下载地址
				// 如果一直都获取不完整，或者对方统计的地址不对，就会使这个进程死循环
				//m.getAddressesCh <- v

				//删除这个不对的版本，重新在下一轮获取
				m.DeleteAddressVersion(&v)
			}

		}
	}

	return nil
}

//HandleTimerTask 设置定时任务
func (m *BitBankNode) HandleTimerTask(name string, handler func(), period time.Duration) {

	m.mu.Lock()
	defer m.mu.Unlock()

	if name == "" {
		return
	}
	if handler == nil {
		return
	}
	if _, exist := m.TaskTimers[name]; exist {
		return
	}

	if m.TaskTimers == nil {
		m.TaskTimers = make(map[string]*timer.TaskTimer)
	}

	//设置定时任务
	task := timer.NewTask(period, handler)
	m.TaskTimers[name] = task

}

//runSubscribeAddressTask 运行订阅地址任务
func (m *BitBankNode) runSubscribeAddressTask() {

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.subscribeAddressTask == nil {
		//启动钱包汇总程序
		task := timer.NewTask(PeriodOfTask, m.updateSubscribeAddress)
		m.subscribeAddressTask = task
		m.subscribeAddressTask.Start()

		//开启获取地址消费者
		go m.getChargeAddress()
	}
	log.Debug("Start Subscribe Address Task...")
	m.subscribeAddressTask.Restart()
}

//updateSubscribeAddress 更新地址
func (m *BitBankNode) updateSubscribeAddress() {

	var (
		err error
	)

	if len(m.subscriptions) == 0 {
		return
	}

	//log.Printf("Update Subscribe Address...\n")
	//获取订阅地址的最新版本
	err = m.GetChargeAddressVersion()
	if err != nil {
		log.Error("GetChargeAddressVersion unexpected error:", err)
	}
}

//SubmitNewRecharges 提交新的充值单
func (m *BitBankNode) SubmitNewRecharges(blockHeight uint64) error {

	var (
		//err      error
		pageSize = 10
	)

	//检查是否连接
	if m.Node == nil {
		return ErrMerchantNodeDisconnected
	}

	for _, s := range m.subscriptions {
		if s.Type == SubscribeTypeCharge {

			wallet, err := m.GetMerchantWalletByID(s.WalletID)
			if err != nil {
				log.Error("GetNewRecharges get wallet unexpected error:", err)
				continue
			}

			recharges, err := wallet.GetRecharges(false)
			if err != nil {
				//log.Error("GetNewRecharges get recharges unexpected error:", err)
				continue
			}

			//config, err := m.GetMerchantWalletConfig(s.Coin,s.WalletID)
			//if err != nil {
			//	log.Printf("GetNewRecharges get wallet config unexpected error: %v", err)
			//	continue
			//}

			if len(recharges) > 0 {

				//做成分页发送交易记录，避免一次卡死
			submitLoop:
				for {
					var subRecharges []*openwallet.Recharge

					if len(recharges) == 0 {
						break submitLoop
					}

					if len(recharges) <= pageSize {
						subRecharges = recharges
						recharges = recharges[:0]
					} else {
						subRecharges = recharges[:pageSize]
						recharges = recharges[pageSize:]
					}

					params := map[string]interface{}{
						"coin":      s.Symbol,
						"walletID":  s.WalletID,
						"recharges": subRecharges,
					}

					//db, inErr := wallet.OpenDB()
					//if inErr != nil {
					//	continue
					//}
					//tx, inErr := db.Begin(true)
					//if inErr != nil {
					//	db.Close()
					//	continue
					//}
					//
					//更新确认数
					for _, r := range subRecharges {
						//log.Printf("Submit Recharges: %v", *r)
						if r.BlockHeight > 0 {
							r.Confirm = int64(blockHeight - r.BlockHeight)
						} else {
							r.Confirm = 0
						}

						//确认数大于配置的确认数
						//if r.Confirm >= int64(config.Confirm) {
						//	//删除已超过确认数的充值记录
						//	tx.DeleteStruct(r)
						//
						//	log.Printf("delete recharge: %s \n ", r.Sid)
						//}
					}
					//tx.Commit()
					//db.Close()

					//提交充值记录
					SubmitRechargeTransaction(
						m.Node,
						m.Config.MerchantNodeID,
						params,
						true,
						func(confirms []uint64, status uint64, msg string) {
							//删除提交已确认的
							if status == owtp.StatusSuccess {

								for _, c := range confirms {

									if c < uint64(len(subRecharges)) {

										db, inErr := wallet.OpenDB()
										if inErr != nil {
											return
										}

										tx, inErr := db.Begin(true)
										if inErr != nil {
											db.Close()
											return
										}

										//标记已成功接收
										subRecharges[c].Received = true
										inErr = tx.Save(subRecharges[c])
										//inErr = tx.DeleteStruct(recharges[c])
										if inErr != nil {
											tx.Rollback()
											db.Close()
											return
										}

										tx.Commit()

										db.Close()
									}

								}

							}
						})
				}

			}
		}
	}

	return nil
}

//blockScanNotify 区块扫描结果通知
func (m *BitBankNode) BlockScanNotify(header *openwallet.BlockHeader) {
	//log.Printf("new block: %v", *header)
	//推送新的充值记录
	err := m.SubmitNewRecharges(header.Height)
	if err != nil {
		log.Error("SubmitNewRecharges unexpected error:", err)
	}
}


