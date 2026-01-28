package openwallet

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"
)

// GetRandomSecure 使用加密安全的随机数生成器生成指定字节数组（推荐）
func GetRandomSecure(l int) ([]byte, error) {
	randomIV := make([]byte, l)
	if _, err := io.ReadFull(rand.Reader, randomIV); err != nil {
		return nil, err
	}
	return randomIV, nil
}

// HmacSHA256 获取字节数组签名值
func HmacSHA256(data, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum([]byte(nil))
}

// SendTransaction 广播交易单
func SendTransaction(decoder TransactionDecoder, wrapper WalletDAI, txData *TxData) (*Transaction, error) {
	if txData == nil {
		return nil, errors.New("txData is nil")
	}
	if txData.Data == "" {
		return nil, errors.New("txData.data is nil")
	}
	if txData.DataSign == "" {
		return nil, errors.New("txData.dataSign is nil")
	}
	if len(txData.SignerList) == 0 {
		return nil, errors.New("txData.SignerList is nil")
	}
	key, err := wrapper.GetTradeKey()
	if err != nil {
		return nil, err
	}
	txJSON := []byte(txData.Data)
	if hex.EncodeToString(HmacSHA256(txJSON, key)) != txData.DataSign {
		return nil, errors.New("txData.dataSign invalid")
	}
	rawTx := &RawTransaction{}
	if err := json.Unmarshal(txJSON, rawTx); err != nil {
		return nil, errors.New("rawTx json error: " + err.Error())
	}
	for accountID, keySignatures := range rawTx.Signatures {
		if keySignatures != nil {
			for k, keySignature := range keySignatures {
				keySignature.Signature = txData.SignerList[fmt.Sprintf("%s-%d", accountID, k)]
			}
		}
		rawTx.Signatures[accountID] = keySignatures
	}

	if err := decoder.VerifyRawTransaction(wrapper, rawTx); err != nil {
		return nil, err
	}

	return decoder.SubmitRawTransaction(wrapper, rawTx)
}

// BuildTransaction 构建普通交易单
func BuildTransaction(decoder TransactionDecoder, wrapper WalletDAI, rawTx *RawTransaction) (*TxData, error) {
	key, err := wrapper.GetTradeKey()
	if err != nil {
		return nil, err
	}
	if err := decoder.CreateRawTransaction(wrapper, rawTx); err != nil {
		return nil, err
	}
	nonce, err := GetRandomSecure(32)
	if err != nil {
		return nil, err
	}
	rawTx.CreateTime = time.Now().UnixMilli()
	rawTx.CreateNonce = hex.EncodeToString(nonce)
	txJSON, err := json.Marshal(rawTx)
	if err != nil {
		return nil, err
	}
	return &TxData{Data: string(txJSON), DataSign: hex.EncodeToString(HmacSHA256(txJSON, key))}, nil
}

// BuildSummaryTransaction 构建汇总交易单列表
func BuildSummaryTransaction(decoder TransactionDecoder, wrapper WalletDAI, sumRawTx *SummaryRawTransaction) ([]*TxData, error) {
	key, err := wrapper.GetTradeKey()
	if err != nil {
		return nil, err
	}
	rawTxArray, err := decoder.CreateSummaryRawTransactionWithError(wrapper, sumRawTx)
	if err != nil {
		return nil, fmt.Errorf("CreateSummaryRawTransactionJSON error: %s", err.Error())
	}
	if len(rawTxArray) == 0 {
		return nil, fmt.Errorf("CreateSummaryRawTransactionJSON create is nil")
	}
	now := time.Now().UnixMilli()
	txData := make([]*TxData, 0, len(rawTxArray))
	for k, v := range rawTxArray {
		nonce, err := GetRandomSecure(32)
		if err != nil {
			return nil, err
		}
		rawTx := v.RawTx
		rawTx.Sid = fmt.Sprintf("%s#%d", sumRawTx.Sid, k)
		rawTx.CreateTime = now
		rawTx.CreateNonce = hex.EncodeToString(nonce)
		txJSON, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		var code, message string
		if v.Error != nil {
			code = strconv.FormatUint(v.Error.code, 10)
			message = v.Error.Error()
		}
		txData = append(txData, &TxData{Data: string(txJSON), DataSign: hex.EncodeToString(HmacSHA256(txJSON, key)), Code: code, Message: message})
	}
	return txData, nil
}
