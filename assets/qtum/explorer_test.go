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

package qtum

import (
	"encoding/hex"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"net/url"
	"testing"
)

//http://192.168.32.107:20003/insight-api/

func TestGetBlockHeightByExplorer(t *testing.T) {
	height, err := tw.getBlockHeightByExplorer()
	if err != nil {
		t.Errorf("getBlockHeightByExplorer failed unexpected error: %v\n", err)
		return
	}
	t.Logf("getBlockHeightByExplorer height = %d \n", height)
}

func TestGetBlockHashByExplorer(t *testing.T) {
	hash, err := tw.getBlockHashByExplorer(249798)
	if err != nil {
		t.Errorf("getBlockHashByExplorer failed unexpected error: %v\n", err)
		return
	}
	t.Logf("getBlockHashByExplorer hash = %s \n", hash)
}

func TestGetBlockByExplorer(t *testing.T) {
	block, err := tw.getBlockByExplorer("42f358f456a7544bbc9d6cf694c4183e2049456da7f06ad9f475fe12757aefe9")
	if err != nil {
		t.Errorf("GetBlock failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetBlock = %v \n", block)
}

func TestListUnspentByExplorer(t *testing.T) {
	list, err := tw.listUnspentByExplorer("QUZMTeBQaChsqPbNsTH5ZxsqF3B4Hi3NCe")
	if err != nil {
		t.Errorf("listUnspentByExplorer failed unexpected error: %v\n", err)
		return
	}
	for i, unspent := range list {
		t.Logf("listUnspentByExplorer[%d] = %v \n", i, unspent)
	}

}

func TestGetTransactionByExplorer(t *testing.T) {
	//5aa478590ea82d3b6a308bdf5af0753caeab0aefefeb4f88a088c15fe305f59b
	//eb8e496f7dd23554d6d45de30beab384c8e0d023c9c7f1fbc15d90d10bb873f8
	raw, err := tw.getTransactionByExplorer("b3357dc8455ffe3fe2cc9a446169b67ea46a02a0344abbe11cdf1ae8d33a3cfd")
	if err != nil {
		t.Errorf("getTransactionByExplorer failed unexpected error: %v\n", err)
		return
	}
	t.Logf("getTransactionByExplorer = %v \n", raw)
}

func TestGetBalanceByExplorer(t *testing.T) {
	raw, err := tw.getBalanceByExplorer("QUZMTeBQaChsqPbNsTH5ZxsqF3B4Hi3NCe ")
	if err != nil {
		t.Errorf("getBalanceByExplorer failed unexpected error: %v\n", err)
		return
	}
	t.Logf("getBalanceByExplorer = %v \n", raw)
}

func TestGetMultiAddrTransactionsByExplorer(t *testing.T) {
	list, err := tw.getMultiAddrTransactionsByExplorer(0, 15, "QUZMTeBQaChsqPbNsTH5ZxsqF3B4Hi3NCe")
	if err != nil {
		t.Errorf("getMultiAddrTransactionsByExplorer failed unexpected error: %v\n", err)
		return
	}
	for i, tx := range list {
		t.Logf("getMultiAddrTransactionsByExplorer[%d] = %v \n", i, tx)
	}

}

func TestSocketIO(t *testing.T) {

	var (
		room       = "inv"
		endRunning = make(chan bool, 1)
	)

	c, err := gosocketio.Dial(
		gosocketio.GetUrl("192.168.32.107", 20003, false),
		transport.GetDefaultWebsocketTransport())
	if err != nil {
		return
	}

	err = c.On("tx", func(h *gosocketio.Channel, args interface{}) {
		log.Info("New transaction received: ", args)
	})
	if err != nil {
		return
	}

	err = c.On("block", func(h *gosocketio.Channel, args interface{}) {
		log.Info("New block received: ", args)
	})
	if err != nil {
		return
	}

	err = c.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {
		log.Info("Disconnected")
	})
	if err != nil {
		return
	}

	err = c.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		log.Info("Connected")
		h.Emit("subscribe", room)
	})
	if err != nil {
		return
	}

	<-endRunning
}

func TestEstimateFeeRateByExplorer(t *testing.T) {
	feeRate, _ := tw.estimateFeeRateByExplorer()
	t.Logf("EstimateFee feeRate = %s\n", feeRate.StringFixed(8))
	fees, _ := tw.EstimateFee(10, 2, feeRate)
	t.Logf("EstimateFee fees = %s\n", fees.StringFixed(8))
}

func TestURLParse(t *testing.T) {
	apiUrl, err := url.Parse("http://192.168.32.107:20003/insight-api/")
	if err != nil {
		t.Errorf("url.Parse failed unexpected error: %v\n", err)
		return
	}
	domain := apiUrl.Hostname()
	port := common.NewString(apiUrl.Port()).Int()
	t.Logf("%s : %d", domain, port)
}

func TestContractBaseAddress(t *testing.T) {
	addrToHash, err := addressEncoder.AddressDecode("QQCf96PCyonzmpDHWqafP86XmenwMPunk9", addressEncoder.QTUM_mainnetAddressP2PKH)
	if err != nil {
		t.Errorf("ContractBaseAddress failed unexpected error: %v\n", err)
		return
	}

	t.Logf("hash = %s", hex.EncodeToString(addrToHash))

	hash, _ := hex.DecodeString("f2033ede578e17fa6231047265010445bca8cf1c")
	hexToAddr := addressEncoder.AddressEncode(hash, addressEncoder.QTUM_mainnetAddressP2PKH)

	t.Logf("addr = %s", hexToAddr)
}

func TestGetAddressTokenBalanceByExplorer(t *testing.T) {
	token := openwallet.SmartContract{
		ContractID: "",
		Address: "f2033ede578e17fa6231047265010445bca8cf1c",
		Symbol: tw.Symbol(),
		Decimals: 8,
	}
	raw, err := tw.getAddressTokenBalanceByExplorer(token, "QbNmvcEPaBssXRmmiaQQQfFcE61po21yQK")
	if err != nil {
		t.Errorf("getAddressTokenBalanceByExplorer failed unexpected error: %v\n", err)
		return
	}
	t.Logf("getAddressTokenBalanceByExplorer = %v \n", raw)
}