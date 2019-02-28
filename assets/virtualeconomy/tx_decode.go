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

package virtualeconomy

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-owcdrivers/virtualeconomyTransaction"
	owcrypt "github.com/blocktree/go-owcrypt"
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
	return decoder.CreateVSYSRawTransaction(wrapper, rawTx)
}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	return decoder.SignVSYSRawTransaction(wrapper, rawTx)
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	return decoder.VerifyVSYSRawTransaction(wrapper, rawTx)
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

	decimals := int32(8)

	tx := openwallet.Transaction{
		From:       rawTx.TxFrom,
		To:         rawTx.TxTo,
		Amount:     rawTx.TxAmount,
		Coin:       rawTx.Coin,
		TxID:       rawTx.TxID,
		Decimal:    decimals,
		AccountID:  rawTx.Account.AccountID,
		Fees:       rawTx.Fees,
		SubmitTime: time.Now().Unix(),
	}

	tx.WxID = openwallet.GenTransactionWxID(&tx)

	return &tx, nil
}

// func createEmptyTransaction(amount, fee, feeScale uint64, fromPub, to string) string {
// 	var (
// 		body = make(map[string]interface{}, 0)
// 	)
// 	body["timestamp"] = time.Now().UnixNano()
// 	body["amount"] = amount
// 	body["fee"] = fee
// 	body["feeScale"] = feeScale
// 	body["recipient"] = to
// 	body["senderPublicKey"] = fromPub
// 	body["attachment"] = ""
// 	body["signature"] = ""

// 	json, _ := json.Marshal(body)

// 	return string(json)
// }

func getTransactionHashForSig(timestamp, amount, fee, feeScale uint64, to string) string {
	var (
		typeID             = byte(0x02)
		timestampBytes     = make([]byte, 8)
		amountBytes        = make([]byte, 8)
		feeBytes           = make([]byte, 8)
		feeScaleBytes      = make([]byte, 2)
		attachmentLenBytes = []byte{0x00, 0x00}
	)

	binary.BigEndian.PutUint64(timestampBytes, timestamp)
	binary.BigEndian.PutUint64(amountBytes, amount)
	binary.BigEndian.PutUint64(feeBytes, fee)
	binary.BigEndian.PutUint16(feeScaleBytes, uint16(feeScale))

	txBytes := []byte{}
	txBytes = append(txBytes, typeID)
	txBytes = append(txBytes, timestampBytes...)
	txBytes = append(txBytes, amountBytes...)
	txBytes = append(txBytes, feeBytes...)
	txBytes = append(txBytes, feeScaleBytes...)

	toPubHash, _ := Decode(to, BitcoinAlphabet)

	txBytes = append(txBytes, toPubHash...)
	txBytes = append(txBytes, attachmentLenBytes...)

	hash := owcrypt.Hash(txBytes, 32, owcrypt.HASH_ALG_BLAKE2B)
	hash = owcrypt.Hash(hash, 32, owcrypt.HASH_ALG_KECCAK256)
	return hex.EncodeToString(hash)
}

func (decoder *TransactionDecoder) CreateVSYSRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	addresses, err := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID)

	if err != nil {
		return err
	}

	if len(addresses) == 0 {
		return fmt.Errorf("No addresses found in wallet [%s]", rawTx.Account.AccountID)
	}

	addressesBalanceList := make([]AddrBalance, 0, len(addresses))

	for i, addr := range addresses {
		balance, err := decoder.wm.Client.getBalance(addr.Address)

		if err != nil {
			return err
		}

		balance.index = i
		addressesBalanceList = append(addressesBalanceList, *balance)
	}

	sort.Slice(addressesBalanceList, func(i int, j int) bool {
		return addressesBalanceList[i].Balance.Cmp(addressesBalanceList[j].Balance) >= 0
	})

	fee := decoder.wm.Config.FeeCharge
	feeScale := decoder.wm.Config.FeeScale
	// fee := big.NewInt(int64(decoder.wm.Config.FeeCharge))

	var amountStr, to string
	for k, v := range rawTx.To {
		to = k
		amountStr = v
		break
	}
	// keySignList := make([]*openwallet.KeySignature, 1, 1)

	amount := big.NewInt(int64(convertFromAmount(amountStr)))
	amount = amount.Add(amount, big.NewInt(int64(fee)))
	from := ""
	fromPubkey := ""
	count := big.NewInt(0)
	countList := []uint64{}
	for _, a := range addressesBalanceList {
		if a.Balance.Cmp(amount) < 0 {
			count.Add(count, a.Balance)
			if count.Cmp(amount) >= 0 {
				countList = append(countList, a.Balance.Sub(a.Balance, count.Sub(count, amount)).Uint64())
				log.Error("The VSYS of the account is enough,"+
					" but cannot be sent in just one transaction!\n"+
					"the amount can be sent in "+string(len(countList))+
					"times with amounts :\n"+strings.Replace(strings.Trim(fmt.Sprint(countList), "[]"), " ", ",", -1), err)
				return err
			} else {
				countList = append(countList, a.Balance.Uint64())
			}
			continue
		}
		from = a.Address
		fromPubkey = addresses[a.index].PublicKey
		break
	}

	if from == "" {
		log.Error("No enough VSYS to send!", err)
		return err
	}

	rawTx.Fees = convertToAmount(fee)
	rawTx.FeeRate = convertToAmount(feeScale)

	publicKey, _ := hex.DecodeString(fromPubkey)
	xpub, err := owcrypt.CURVE25519_convert_Ed_to_X(publicKey)
	if err != nil {
		return err
	}
	fromPubkey = Encode(xpub, BitcoinAlphabet)

	txStruct := virtualeconomyTransaction.TxStruct{
		TxType:     virtualeconomyTransaction.TxTypeTransfer,
		To:         to,
		Amount:     convertFromAmount(amountStr),
		Fee:        fee,
		FeeScale:   uint16(feeScale),
		Attachment: "",
	}
	emptyTrans, err := virtualeconomyTransaction.CreateEmptyTransaction(txStruct)
	if err != nil {
		return err
	}
	rawTx.RawHex = emptyTrans

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	keySigs := make([]*openwallet.KeySignature, 0)

	// address, err := decoder.wm.Decoder.PublicKeyToAddress(xpub, decoder.wm.Config.IsTestNet)
	// if err != nil {
	// 	return err
	// }

	addr, err := wrapper.GetAddress(from)
	if err != nil {
		return err
	}
	signature := openwallet.KeySignature{
		EccType: decoder.wm.Config.CurveType,
		Nonce:   "",
		Address: addr,
		Message: emptyTrans,
	}

	keySigs = append(keySigs, &signature)

	rawTx.Signatures[rawTx.Account.AccountID] = keySigs

	rawTx.FeeRate = big.NewInt(int64(feeScale)).String()

	rawTx.IsBuilt = true

	return nil
}

func (decoder *TransactionDecoder) SignVSYSRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
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

			//签名交易
			/////////交易单哈希签名
			sigPub, err := virtualeconomyTransaction.SignTransaction(keySignature.Message, keyBytes)
			if err != nil {
				return fmt.Errorf("transaction hash sign failed, unexpected error: %v", err)
			} else {

				//for i, s := range sigPub {
				//	log.Info("第", i+1, "个签名结果")
				//	log.Info()
				//	log.Info("对应的公钥为")
				//	log.Info(hex.EncodeToString(s.Pubkey))
				//}

				// txHash.Normal.SigPub = *sigPub
			}

			keySignature.Signature = hex.EncodeToString(sigPub.Signature)
		}
	}

	log.Info("transaction hash sign success")

	rawTx.Signatures[rawTx.Account.AccountID] = keySignatures

	return nil
}

func (decoder *TransactionDecoder) VerifyVSYSRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		emptyTrans      = rawTx.RawHex
		signaturePubkey = virtualeconomyTransaction.SignaturePubkey{}
	)

	for accountID, keySignatures := range rawTx.Signatures {
		log.Debug("accountID Signatures:", accountID)
		for _, keySignature := range keySignatures {

			signature, _ := hex.DecodeString(keySignature.Signature)
			pubkey, _ := hex.DecodeString(keySignature.Address.PublicKey)

			edpub, _ := owcrypt.CURVE25519_convert_Ed_to_X(pubkey)

			signaturePubkey = virtualeconomyTransaction.SignaturePubkey{
				Signature: signature,
				PublicKey: edpub,
			}

			log.Debug("Signature:", keySignature.Signature)
			log.Debug("PublicKey:", keySignature.Address.PublicKey)
		}
	}

	pass := virtualeconomyTransaction.VerifyTransaction(emptyTrans, &signaturePubkey)

	signedTrans, err := virtualeconomyTransaction.CreateJSONRawForSendTransaction(emptyTrans, &signaturePubkey)
	if err != nil {
		return fmt.Errorf("transaction compose signatures failed")
	}

	if pass {
		log.Debug("transaction verify passed")
		rawTx.IsCompleted = true
		rawTx.RawHex = signedTrans.Raw
	} else {
		log.Debug("transaction verify failed")
		rawTx.IsCompleted = false
	}

	return nil
}

func (decoder *TransactionDecoder) GetRawTransactionFeeRate() (feeRate string, unit string, err error) {
	rate := decoder.wm.Config.FeeCharge
	return convertToAmount(rate), "TX", nil
}
