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

package cardano

import (
	"github.com/blocktree/openwallet/common"
	"testing"
	"log"
	"github.com/tidwall/gjson"
)

var wm *WalletManager

func init() {
	wm = NewWalletManager()
	wm.InitConfigFlow()
	wm.Config.ServerAPI = "http://47.106.34.230:10026/api/v1/"
	wm.WalletClient = NewClient(wm.Config.ServerAPI, true)
}

func TestGetWalletInfo(t *testing.T) {
	ret := wm.WalletClient.callGetWalletAPI("Ae2tdPwUPEZEL7DHjYaZBH4T4fXkrvpryAeGbjmqLy62pPXg6aLapS9xJom")
	t.Log(string(ret))

	ret = wm.WalletClient.callGetWalletAPI()
	t.Log(string(ret))

	list, err := wm.GetWalletInfo()
	if err != nil {
		t.Log(err)
	}


	//打印钱包列表
	wm.printWalletList(list)
}

func TestCreateNewWallet(t *testing.T) {
	//密钥32
	password := common.NewString("1234qwer").SHA256()
	words := genMnemonic()
	//annual only pact client fatigue choice arrive achieve country indoor engage coil spatial engine among paper dawn tackle bonus task lock pepper deny eye
	result := wm.WalletClient.callCreateWalletAPI("12 words wallet", words, password, true)
	t.Logf("%v\n", string(result))
}

func TestGetAccountInfo(t *testing.T) {

	tests := []struct {
		wid string
		aid string
	}{
		{
			wid: "Ae2tdPwUPEZEL7DHjYaZBH4T4fXkrvpryAeGbjmqLy62pPXg6aLapS9xJom",
			aid: "3669395481",
		},
		//{
		//	wid: "Ae2tdPwUPEZ2mvUueYDm8JCy7nSV6VybMkwf3umv3hFkB3y1Gvze9NUkDLT",
		//	aid: "123",
		//},
	}

	for i, test := range tests {

		accounts, err := wm.GetAccountInfo(test.wid, "")
		if err != nil {
			t.Errorf("GetAccountInfo[%d] failed unexpected error: %v", i, err)
		}
		for j, a := range accounts {
			t.Logf("GetAccountInfo[%d] acount[%d]  = %v", i, j, a)
		}

	}
}

func TestCreateAccount(t *testing.T) {

	tests := []struct {
		wid      string
		name     string
		password string
	}{
		//{
		//	wid:      "Ae2tdPwUPEZ7R5jNgL8SKRtG8QFTi2QuQrsScydp3GJAPStvivfjTXPTSKX",
		//	name:     "test1",
		//	password: common.NewString("1234qwer").SHA256(),
		//},
		{
			wid:      "Ae2tdPwUPEZEL7DHjYaZBH4T4fXkrvpryAeGbjmqLy62pPXg6aLapS9xJom",
			name:     "test2",
			password: common.NewString("1234qwer").SHA256(),
		},
	}

	for i, test := range tests {

		a, err := wm.CreateNewAccount(test.name, test.wid, test.password)
		if err != nil {
			t.Fatalf("CreateAccount[%d] failed unexpected error: %v", i, err)
		}
		t.Log(a)
		t.Logf("CreateAccount[%d] AcountID = %d", i, a.Index)
	}

}

func TestCreateBatchAddress(t *testing.T) {
	//	Ae2tdPwUPEZ7R5jNgL8SKRtG8QFTi2QuQrsScydp3GJAPStvivfjTXPTSKX

	var (
	//results = make(chan []*Address, 0)
	)

	password := common.NewString("1234qwer").SHA256()

	tests := []struct {
		wid string
		aid   int64
		count uint
	}{
		//{
		//	wid: "Ae2tdPwUPEZ7R5jNgL8SKRtG8QFTi2QuQrsScydp3GJAPStvivfjTXPTSKX",
		//	aid:   4257117526,
		//	count: 10,
		//},
		{
			wid: "Ae2tdPwUPEZEL7DHjYaZBH4T4fXkrvpryAeGbjmqLy62pPXg6aLapS9xJom",
			aid: 3669395481,
			count: 10,
		},
	}

	for i, test := range tests {

		addrs, _, err := wm.CreateBatchAddress(test.wid, test.aid, password, test.count)
		if err != nil {
			t.Fatalf("CreateBatchAddress[%d] failed unexpected error: %v", i, err)
		}

		for j, a := range addrs {
			t.Logf("CreateBatchAddress[%d] NewAddress[%d]  = %s", i, j, a.Address)
		}

	}
}


func TestGetAddressInfo(t *testing.T) {

	tests := []struct {
		wid string
		aid string
	}{
		{
			wid: "Ae2tdPwUPEZEL7DHjYaZBH4T4fXkrvpryAeGbjmqLy62pPXg6aLapS9xJom",
			aid: "3669395481",
		},

	}

	for i, test := range tests {

		addrs, err := wm.GetAddressInfo(test.wid, test.aid)
		if err != nil {
			t.Errorf("GetAddressInfo[%d] failed unexpected error: %v", i, err)
		}
		for j, a := range addrs {
			t.Logf("GetAddressInfo[%d] address[%d]  = %v", i, j, a)
		}

	}
}

func TestSendTx(t *testing.T) {
	tests := []struct {
		wid    string
		aid    int64
		to     string
		amount uint64
	}{
		{
			wid:    "Ae2tdPwUPEZ7R5jNgL8SKRtG8QFTi2QuQrsScydp3GJAPStvivfjTXPTSKX",
			aid:    4257117526,
			to:     "DdzFFzCqrhsz6RtTEaL1WzrjAMzVdLwwQdQ5ivAA58hSWnG9oTk9LgczLAcvEEDxgKegfcLwWBdihBaHyBQyGwQg8ReRq6R41qXbmXkG",
			amount: 2500000,
		},
	}

	//密钥32
	password := common.NewString("1234qwer").SHA256()

	for i, test := range tests {

		tx, err := wm.SendTx(test.wid, test.aid, test.to, test.amount, password)
		if err != nil {
			t.Errorf("SendTx[%d] failed unexpected error: %v", i, err)
		} else {
			t.Logf("SendTx[%d] tx = %v", i, tx)
		}
	}

}

func TestSummaryFollow(t *testing.T) {
	//SummaryFollow()
}

func TestEstimateFees(t *testing.T) {

	tests := []struct {
		wid    string
		aid    int64
		to     string
		amount uint64
	}{
		{
			wid:    "Ae2tdPwUPEZ7R5jNgL8SKRtG8QFTi2QuQrsScydp3GJAPStvivfjTXPTSKX",
			aid:    4257117526,
			to:     "DdzFFzCqrhsz6RtTEaL1WzrjAMzVdLwwQdQ5ivAA58hSWnG9oTk9LgczLAcvEEDxgKegfcLwWBdihBaHyBQyGwQg8ReRq6R41qXbmXkG",
			amount: 2500000,
		},
	}

	//密钥32
	password := common.NewString("1234qwer").SHA256()
	for _, test := range tests {

		result, _ := wm.WalletClient.callEstimateFeesAPI(test.wid, test.aid, test.to, test.amount, password)
		err := isError(result)
		if err != nil {
			log.Printf("%v\n", err)
		}
		fees := gjson.GetBytes(result, "data.estimatedAmount")
		log.Printf("fees %s\n", fees.String())
	}


}


func TestDeleteWallet(t *testing.T) {
	wid := "Ae2tdPwUPEZMYbseGswFLCyN1vG6CjyGNcAvNujB3JXbuovD6GVc535GTtB"
	result := wm.WalletClient.callDeleteWallet(wid)
	content := gjson.ParseBytes(result)
	t.Logf("DeleteWalle %v\n", content)
}

func TestRestoreWallet(t *testing.T) {
	//密钥32
	password := common.NewString("12345678").SHA256()
	words := "traffic horror food daughter account satisfy alley stuff paddle urge hollow cook arena there include know catalog sting brother crack math buddy identify demise"
	//annual only pact client fatigue choice arrive achieve country indoor engage coil spatial engine among paper dawn tackle bonus task lock pepper deny eye
	result := wm.WalletClient.callCreateWalletAPI("test wallet", words, password, true)
	content := gjson.ParseBytes(result)
	t.Logf("RestoreWallet %v\n", content)
}

func TestBackupWalletkey(t *testing.T) {
	//BackupWalletkey()
}

func TestGetNodeInfo(t *testing.T) {
	ret := wm.WalletClient.callGetNodeInfo()
	t.Log(string(ret))
}