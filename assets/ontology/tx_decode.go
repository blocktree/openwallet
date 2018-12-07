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

package ontology

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-owcdrivers/btcTransaction"
	"github.com/blocktree/go-owcdrivers/ontologyTransaction"
)

type TransactionDecoder struct {
	openwallet.TransactionDecoderBase
	wm *WalletManager //钱包管理者
}

//NewTransactionDecoder 交易单解析器
func NewTransactionDecoder(wm *WalletManager) *TransactionDecoder {
	decoder := TransactionDecoder{}
	decoder.wm = wm
	return &decoder
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	return decoder.CreateONTRawTransaction(wrapper, rawTx)
}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	return decoder.SignONTRawTransaction(wrapper, rawTx)
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	return decoder.VerifyONTRawTransaction(wrapper, rawTx)
}

func (decoder *TransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {
	if len(rawTx.RawHex) == 0 {
		return nil, fmt.Errorf("transaction hex is empty")
	}

	if !rawTx.IsCompleted {
		return nil, fmt.Errorf("transaction is not completed validation")
	}

	txid, err := decoder.wm.SendRawTransaction(rawTx.RawHex)
	if err != nil {
		return nil, err
	}

	rawTx.TxID = txid
	rawTx.IsSubmit = true
	tx := openwallet.Transaction{
		WxID: txid,
		TxID: txid,
	}
	return &tx, nil
}

func (decoder *TransactionDecoder) CreateONTRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		txState  ontologyTransaction.TxState
		gasPrice = ontologyTransaction.DefaultGasPrice
		gasLimit = ontologyTransaction.DefaultGasLimit
	)

	addresses, err := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID)

	if err != nil {
		return err
	}

	if len(addresses) == 0 {
		return fmt.Errorf("No addresses found in wallet [%s]", rawTx.Account.AccountID)
	}

	addressesBalanceList := make([]AddrBalance, 0, len(addresses))

	for i, addr := range addresses {
		balance, err := decoder.wm.RPCClient.getBalance(addr.Address)

		if err != nil {
			return err
		}
		balance.index = i
		addressesBalanceList = append(addressesBalanceList, *balance)
	}

	sort.Slice(addressesBalanceList, func(i int, j int) bool {
		return addressesBalanceList[i].ONTBalance.Cmp(addressesBalanceList[j].ONTBalance) >= 0
	})

	fee := big.NewInt(int64(gasLimit * gasPrice))

	var amountStr, to string
	for k, v := range rawTx.To {
		to = k
		amountStr = v
		break
	}
	keySignList := make([]*openwallet.KeySignature, 1, 1)

	if rawTx.Coin.ContractID == ontologyTransaction.ONGContractAddress {
		amount, err := convertFlostStringToBigInt(amountStr)
		if err != nil {
			return err
		}

		if amount.Cmp(big.NewInt(0)) == 0 { // ONG unbound
			txState.AssetType = ontologyTransaction.AssetONGWithdraw
			txState.From = to
			txState.To = to

			for i, a := range addressesBalanceList {
				if a.Address != to {
					continue
				}
				if a.ONGUnbound.Cmp(big.NewInt(0)) == 0 {
					log.Error("No unbound ONG to withdraw in address : "+to, err)
					return err
				}

				if a.ONGUnbound.Cmp(fee) <= 0 {
					log.Error("Unbound ONG is not enough to withdraw in address : "+to, err)
					return err
				}

				txState.Amount = amount.Sub(a.ONGUnbound, fee).Uint64()
				keySignList = append(keySignList, &openwallet.KeySignature{
					Address: &openwallet.Address{
						AccountID:   addresses[addressesBalanceList[i].index].AccountID,
						Address:     addresses[addressesBalanceList[i].index].Address,
						PublicKey:   addresses[addressesBalanceList[i].index].PublicKey,
						Alias:       addresses[addressesBalanceList[i].index].Alias,
						Tag:         addresses[addressesBalanceList[i].index].Tag,
						Index:       addresses[addressesBalanceList[i].index].Index,
						HDPath:      addresses[addressesBalanceList[i].index].HDPath,
						WatchOnly:   addresses[addressesBalanceList[i].index].WatchOnly,
						Symbol:      addresses[addressesBalanceList[i].index].Symbol,
						Balance:     addresses[addressesBalanceList[i].index].Balance,
						IsMemo:      addresses[addressesBalanceList[i].index].IsMemo,
						Memo:        addresses[addressesBalanceList[i].index].Memo,
						CreatedTime: addresses[addressesBalanceList[i].index].CreatedTime,
					},
				})
				break
			}

			if amount.Cmp(big.NewInt(0)) == 0 {
				log.Error("Address : "+to+" not found!", err)
				return err
			}

		} else { // ONG transaction
			txState.AssetType = ontologyTransaction.AssetONG
			txState.Amount = amount.Uint64()
			txState.To = to
			count := big.NewInt(0)
			countList := []uint64{}
			for _, a := range addressesBalanceList {
				if a.ONGBalance.Cmp(amount) < 0 {
					count.Add(count, a.ONGBalance)
					if count.Cmp(amount) >= 0 {
						countList = append(countList, a.ONGBalance.Sub(a.ONGBalance, count.Sub(count, amount)).Uint64())
						log.Error("The ONG of the account is enough,"+
							" but cannot be sent in just one transaction!\n"+
							"the amount can be sent in "+string(len(countList))+
							"times with amounts :\n"+strings.Replace(strings.Trim(fmt.Sprint(countList), "[]"), " ", ",", -1), err)
						return err
					} else {
						countList = append(countList, a.ONGBalance.Uint64())
					}
					continue
				}
				txState.From = a.Address
				break
			}

			if txState.From == "" {
				log.Error("No enough ONT to send!", err)
				return err
			}
		}
	} else { // ONT transaction
		amount, err := convertIntStringToBigInt(amountStr)
		if err != nil {
			return err
		}
		txState.AssetType = ontologyTransaction.AssetONT
		txState.Amount = amount.Uint64()
		txState.To = to
		count := big.NewInt(0)
		countList := []uint64{}
		for _, a := range addressesBalanceList {
			if a.ONTBalance.Cmp(amount) < 0 {
				count.Add(count, a.ONTBalance)
				if count.Cmp(amount) >= 0 {
					countList = append(countList, a.ONTBalance.Sub(a.ONTBalance, count.Sub(count, amount)).Uint64())
					log.Error("The ONT of the account is enough,"+
						" but cannot be sent in just one transaction!\n"+
						"the amount can be sent in "+string(len(countList))+
						"times with amounts :\n"+strings.Replace(strings.Trim(fmt.Sprint(countList), "[]"), " ", ",", -1), err)
					return err
				} else {
					countList = append(countList, a.ONTBalance.Uint64())
				}
				continue
			}
			txState.From = a.Address
			break
		}

		if txState.From == "" {
			log.Error("No enough ONT to send!", err)
			return err
		}
	}

	feeInONG, _ := convertBigIntToFloatDecimal(fee.String())

	rawTx.Fees = feeInONG.String()

	emptyTrans, err := ontologyTransaction.CreateEmptyRawTransaction(gasPrice, gasLimit, txState)
	if err != nil {
		return err
	}

	transHash, err := ontologyTransaction.CreateRawTransactionHashForSig(emptyTrans)
	if err != nil {
		return err
	}
	rawTx.RawHex = emptyTrans

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	keySigs := make([]*openwallet.KeySignature, 0)

	addr, err := wrapper.GetAddress(transHash.GetNormalTxAddress())
	if err != nil {
		return err
	}
	signature := openwallet.KeySignature{
		EccType: decoder.wm.Config.CurveType,
		Nonce:   "",
		Address: addr,
		Message: transHash.GetTxHashHex(),
	}

	keySigs = append(keySigs, &signature)

	rawTx.Signatures[rawTx.Account.AccountID] = keySigs

	rawTx.FeeRate = big.NewInt(int64(gasPrice)).String()

	rawTx.IsBuilt = true

	return nil
}

func (decoder *TransactionDecoder) SignONTRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	key, err := wrapper.HDKey()
	if err != nil {
		return nil
	}

	keySignatures := rawTx.Signatures[rawTx.Account.AccountID]

	if keySignatures != nil {
		for _, keySignature := range keySignatures {

			childKey, err := key.DerivedKeyWithPath(keySignature.Address.HDPath, keySignature.EccType)
			keyBytes, err := childKey.GetPrivateKeyBytes()
			if err != nil {
				return err
			}
			log.Debug("privateKey:", hex.EncodeToString(keyBytes))

			//privateKeys = append(privateKeys, keyBytes)
			txHash := ontologyTransaction.TxHash{
				Hash: keySignature.Message,
				Normal: &ontologyTransaction.NormalTx{
					Address: keySignature.Address.Address,
					SigType: btcTransaction.SigHashAll,
				},
			}

			log.Debug("hash:", txHash.GetTxHashHex())

			//签名交易
			/////////交易单哈希签名
			sigPub, err := ontologyTransaction.SignRawTransactionHash(txHash.GetTxHashHex(), keyBytes)
			if err != nil {
				return fmt.Errorf("transaction hash sign failed, unexpected error: %v", err)
			} else {

				//for i, s := range sigPub {
				//	log.Info("第", i+1, "个签名结果")
				//	log.Info()
				//	log.Info("对应的公钥为")
				//	log.Info(hex.EncodeToString(s.Pubkey))
				//}

				//txHash.Normal.SigPub = *sigPub
			}

			keySignature.Signature = hex.EncodeToString(sigPub.Signature)
		}
	}

	log.Info("transaction hash sign success")

	rawTx.Signatures[rawTx.Account.AccountID] = keySignatures

	return nil
}

func (decoder *TransactionDecoder) VerifyONTRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		emptyTrans = rawTx.RawHex
		transHash  = make([]ontologyTransaction.TxHash, 0)
	)

	for accountID, keySignatures := range rawTx.Signatures {
		log.Debug("accountID Signatures:", accountID)
		for _, keySignature := range keySignatures {

			signature, _ := hex.DecodeString(keySignature.Signature)
			pubkey, _ := hex.DecodeString(keySignature.Address.PublicKey)

			signaturePubkey := ontologyTransaction.SigPub{
				Signature: signature,
				PublicKey: pubkey,
			}

			//sigPub = append(sigPub, signaturePubkey)

			txHash := ontologyTransaction.TxHash{
				Hash: keySignature.Message,
				Normal: &ontologyTransaction.NormalTx{
					Address: keySignature.Address.Address,
					SigType: btcTransaction.SigHashAll,
					SigPub:  signaturePubkey,
				},
			}

			transHash = append(transHash, txHash)

			log.Debug("Signature:", keySignature.Signature)
			log.Debug("PublicKey:", keySignature.Address.PublicKey)
		}
	}

	signedTrans, err := ontologyTransaction.InsertSignatureIntoEmptyTransaction(emptyTrans, transHash[0].Normal.SigPub)
	if err != nil {
		return fmt.Errorf("transaction compose signatures failed")
	}

	pass := ontologyTransaction.VerifyRawTransaction(signedTrans)

	if pass {
		log.Debug("transaction verify passed")
		rawTx.IsCompleted = true
		rawTx.RawHex = signedTrans
	} else {
		log.Debug("transaction verify failed")
		rawTx.IsCompleted = false
	}

	return nil
}
