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
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/OpenWallet/owtp"
	"log"
)

//CallGetChargeAddressVersion 获取要订阅的地址版本信息
func GetChargeAddressVersion(
	node *owtp.OWTPNode,
	params interface{},
	sync bool,
	callback func(addressVer *AddressVersion, status uint64, msg string)) error {

	//| 参数名称 | 类型   | 是否可空 | 描述     |
	//|----------|--------|----------|----------|
	//| coin     | string | 是       | 币名     |
	//| walletID | string | 是       | 钱包ID   |

	//获取订阅的地址版本
	err := node.Call(
		"getChargeAddressVersion",
		params,
		sync,
		func(resp owtp.Response) {
			//log.Printf(" getChargeAddressVersion Response: %v", resp)
			result := resp.JsonData()
			if result.Exists() {
				addressVersion := NewAddressVersion(result)
				callback(addressVersion, resp.Status, resp.Msg)
			} else {
				callback(nil, resp.Status, resp.Msg)
			}

		})

	return err
}

//GetChargeAddress 获取订阅的地址表
func GetChargeAddress(
	node *owtp.OWTPNode,
	params interface{},
	sync bool,
	callback func(addrs []*openwallet.Address, status uint64, msg string)) error {

	//| 参数名称 | 类型   | 是否可空 | 描述     |
	//|----------|--------|----------|----------|
	//| coin     | string | 是       | 币名     |
	//| walletID | string | 是       | 钱包ID   |
	//| offset   | int    | 是       | 偏移量   |
	//| limit    | int    | 是       | 读取条数 |

	//获取订阅的地址版本
	err := node.Call(
		"getChargeAddress",
		params,
		sync,
		func(resp owtp.Response) {
			addrs := make([]*openwallet.Address, 0)
			result := resp.JsonData()
			if result.Exists() {
				arrs := result.Get("addresses").Array()
				for _, ad := range arrs {
					a := openwallet.NewAddress(ad)
					addrs = append(addrs, a)
				}
			}

			callback(addrs, resp.Status, resp.Msg)

		})

	return err
}

//SubmitRechargeTrasaction 提交充值到账记录
func SubmitRechargeTrasaction(
	node *owtp.OWTPNode,
	params interface{},
	sync bool,
	callback func(confirms []uint64, status uint64, msg string)) error {

	log.Printf("Call Remote: submitRechargeTrasaction\n")

	//获取订阅的地址版本
	err := node.Call(
		"submitRechargeTrasaction",
		params,
		sync,
		func(resp owtp.Response) {
			confirms := make([]uint64, 0)
			result := resp.JsonData()
			if result.Exists() {
				arrs := result.Get("confirms").Array()
				for _, c := range arrs {
					confirms = append(confirms, c.Uint())
				}
			}
			callback(confirms, resp.Status, resp.Msg)

		})

	return err
}
