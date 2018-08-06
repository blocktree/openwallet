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

package iota

import (
	"testing"
	"github.com/iotaledger/giota"
)

var(
	//testServer          ="https://nodes.devnet.iota.org:443"
	ourServer           = "http://47.52.16.168:14265"
	seed                = "BAXSMBAOWINJDKR9AWTBKJZDVURXRSNCCSRXBCRSSSIERR9SAJICUCPC9QAJEIKYBGJJIKEXWPWLHD9HV"
	//skipTransferTest    = false
)

func TestServer(t *testing.T){
	//api := giota.NewAPI(testServer, nil)
	api := giota.NewAPI(ourServer, nil)
	resp, err := api.GetNodeInfo()
	if err != nil {
		t.Errorf("TestServer() expected err to be nil but got %v", err)

	}else if resp.AppName == "" {
		t.Errorf("TestServer() returned invalid response: %#v", resp)

	}else {
		t.Logf("TestServer() success：%#v", resp)
	}
}

/*
func TestAPIGetNodeInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var err error
	var resp *giota.GetNodeInfoResponse

	for i := 0; i < 5; i++ {
		var server = giota.RandomNode()
		api := giota.NewAPI(server, nil)
		resp, err = api.GetNodeInfo()
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Errorf("GetNodeInfo() expected err to be nil but got %v", err)

	}else if resp.AppName == "" {
		t.Errorf("GetNodeInfo() returned invalid response: %#v", resp)

	}else {
		t.Logf("GetNodeInfo() success：%#v", resp)
	}
}
*/

func TestNewWallet(t *testing.T){
	seed := giota.NewSeed()
	t.Logf("New Seed: %v",seed)
	t.Logf("This seed is for your new wallet, please keep your seed in a safe place.")
}

func TestGetWalletInfo(t *testing.T){

	adr,adrs,totalBalance,err := GetWalletInfo(seed)
	if err != nil{
		t.Error(err)
	}
	t.Log(adr, adrs)
	t.Logf("Total Balance = %d\n",totalBalance)
}

func TestCreateAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	trytesFrom:="NLYZQODTURLQHCFXSMHTBBLTKFTXQTJTPBRB9MMVGAJBOWAKKHZYNOHPDJVALFS9EETEOJBWNDTGKCHXO"
	trytes,err:=giota.ToTrytes(trytesFrom)
	if err != nil{
		t.Error(err)
	}
	index:=10
	security:=2
	adr,err:=giota.NewAddress(trytes,index,security) //without checksum.
	if err != nil {
		t.Errorf("TestNewAddress([]) expected err to be nil but got %v", err)
	}else {
		t.Logf("TestNewAddress() = %#v", adr)
		adrWithChecksum := adr.WithChecksum() //adrWithChecksum is trytes type.
		t.Logf("TestNewAddress() = %#v", adrWithChecksum)
	}
}

func TestCreateAddresses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	trytesFrom:="ZVPBLFLIEIGYTQGOGOMBDPKEUKEVHCXHUBJHLRYWDXPYMDLZNLNPQHPY9GNCEXCZZBSRMMNXBSXLKQEGA"
	trytes,err:=giota.ToTrytes(trytesFrom)
	if err != nil{
		t.Error(err)
	}
	start:=0
	count:=100
	security:=2

	backupFile,err := CreateAddresses(trytes,start,count,security)
	if err != nil{
		t.Error(err)
	}else {
		t.Logf("CreateAddresses successfully, backup path:%s",backupFile)
	}
}

//func init() {
//	ts := os.Getenv("ZVPBLFLIEIGYTQGOGOMBDPKEUKEVHCXHUBJHLRYWDXPYMDLZNLNPQHPY9GNCEXCZZBSRMMNXBSXLKQEGA")
//	if ts == "" {
//		skipTransferTest = true
//		return
//	}
//
//	s, err := giota.ToTrytes(ts)
//	if err != nil {
//		skipTransferTest = true
//	} else {
//		seed = s
//	}
//}

func TestGetUsedAddressesAndTotalBalances(t *testing.T) {
	//if skipTransferTest {
	//	t.Skip("transfer test skipped because a valid $TRANSFER_TEST_SEED was not specified")
	//}


	var (
		err  error
		adr  giota.Address
		adrs []giota.Address
	)

	trytes,err:=giota.ToTrytes(seed)
	if err != nil{
		t.Error(err)
	}

	for i := 0; i < 5; i++ {
		api := giota.NewAPI(giota.RandomNode(), nil)
		adr, adrs, err = giota.GetUsedAddress(api, trytes, 2)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Error(err)
	}

	t.Log(adr, adrs)
	if len(adrs) < 1 {
		t.Error("GetUsedAddress is incorrect")
	}

	//add by chenzhiwen
	var totalBalance int64
	for i:=0;i< len(adrs);i++{
		api := giota.NewAPI(giota.RandomNode(), nil)
		resp, err := api.GetBalances([]giota.Address{adrs[i]}, 100)
		if err == nil {
			totalBalance += resp.Balances[0]
		}
	}
	t.Logf("Total Balance = %d\n",totalBalance)


	var bal giota.Balances
	for i := 0; i < 5; i++ {
		api := giota.NewAPI(giota.RandomNode(), nil)
		bal, err = giota.GetInputs(api, trytes, 0, 10, 1000, 2)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Error(err)
	}

	t.Log(bal)
	if len(bal) < 1 {
		t.Error("GetInputs is incorrect")
	}
}


// nolint: gocyclo
func TestSendTransaction(t *testing.T) {

	var(
		address giota.Address
		value int64
		tag giota.Trytes
	)

	address="WRD9LXQGEM9WOIWWIRGCFDLOPMWBHZ9YFCXVAGZJBBHL9GKCSYFRAUJNM9DWGDIANUQMJ9FIZWLYDKKHW"
	value=20
	tag="MOUDAMEPO"

	err := SendTransaction(seed, address, value, tag)
	if err != nil{
		t.Errorf("TestSendTransaction() expected err to be nil but got: %v",err)
	}
}


func TestAPIGetBalancesByAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var err error
	var resp *giota.GetBalancesResponse

	for i := 0; i < 5; i++ {
		var server = giota.RandomNode()
		api := giota.NewAPI(server, nil)

		resp, err = api.GetBalances([]giota.Address{"IFVY9TUWMKDWFTNXUBLHRUEJMEBCWCKFHZRWQRMKYE9WENWZTJWBKBVDXIWRQR9AY9HG99TSFULEQVRA9"}, 100)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Errorf("GetBalances([]) expected err to be nil but got %v", err)
	}else {
		t.Logf("GetBalances() = %#v, Balance = %d", resp, resp.Balances)
	}
}


