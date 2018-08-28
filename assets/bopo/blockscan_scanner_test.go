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

package bopo

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
)

func TestScanBlock(t *testing.T) {
	bs := NewFabricBlockScanner(tw)

	bs.scanBlock()
}

func TestAssignedScanBlock(t *testing.T) {

	bs := NewFabricBlockScanner(tw)

	accountID := "simonluo"
	// wallet, err := tw.GetWalletInfo(accountID)
	// if err != nil {
	// 	t.Errorf("BTCBlockScanner_Run failed unexpected error: %v\n", err)
	// 	return
	// }
	wallet := &openwallet.Wallet{}

	bs.AddWallet(accountID, wallet)

	bs.ScanBlock(uint64(231234))

}

func TestBlockScannerData(t *testing.T) {
	bs := NewFabricBlockScanner(tw)

	var currentHeight uint64
	currentHeight = 375350
	currentHeight = 377125
	currentHeight = 378820
	currentHeight = 330000
	for height := currentHeight; height <= currentHeight+1500; height++ { //Foreach Blocks
		// Load Block Info
		block, err := bs.wm.GetBlockContent(height)
		if err != nil {
			log.Std.Info("Get block [%d] faild: %v\n", height, err)
		}

		fmt.Printf("Height=[%d/%d]Len(TXs)=[%d]\tPreHash[%s]\n", height, currentHeight, len(block.Transactions), block.Previousblockhash)

		for i, v := range block.Transactions { // Foreach all transactions

			fmt.Printf("\tNo.[%2d]\tType=[%s]\tChaincodeID[%s]", i, v.Type, v.ChaincodeID)

			if payloadSpec, err := bs.wm.GetBlockPayload(base64.StdEncoding.EncodeToString(v.Payload)); err != nil {
				log.Std.Info("Decode TX [%d] Payload faild: %v\n", height, err)
			} else {
				//fmt.Println(payloadSpec)
				fmt.Printf("\tFrom[%s]to[%s]with[%d Pai]", payloadSpec.From, payloadSpec.To, payloadSpec.Amount)
				if payloadSpec.From == "5ZaPXfJaLNrGnXuyXunFE4xKxakEzgTIZQ" {
					fmt.Println("simonluo2")
					if payloadSpec.To == "5ZFVVP47Rf5j-k7LoiRcNozlc8dynbPYng" {
						fmt.Println("xcluo2")
						return
					}
				}
			}
			fmt.Printf("\n")
		}
		// time.Sleep(time.Second * 1 / 100)
	}
}
