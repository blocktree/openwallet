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
	"path/filepath"
	"github.com/blocktree/OpenWallet/owtp"
	"testing"
	"time"
)
var mqURL = "192.168.30.160:5672"
var nodeConfig NodeConfig
func init() {
	nodeConfig = NodeConfig{
		MerchantNodeURL: mqURL,
		CacheFile:       filepath.Join(merchantDir, cacheFile),
		ConnectType:owtp.MQ,
		Exchange:"DEFAULT_EXCHANGE",
		QueueName:"ALL_READ",
		ReceiveQueueName:"ALL_WRITE",
		Account:"admin",
		Password:"admin",
	}
}


func TestNewNode(t *testing.T) {

	nodeConfig := NodeConfig{
		MerchantNodeURL :mqURL,
		ConnectType     :"mq",
		Exchange     :  "DEFAULT_EXCHANGE",
		QueueName     :  "DEFAULT_QUEUE",
		ReceiveQueueName  :     "DEFAULT_QUEUE",
		Account    :   "admin",
		Password    :   "admin",
	}
	node,_ := NewBitNodeNode(nodeConfig)
	node.Run()
	time.Sleep(3 * time.Second)
	config := make(map[string]string)
	config["address"] = ":8675"
	config["connectType"] = "http"
	node.Node.Listen(config)

}