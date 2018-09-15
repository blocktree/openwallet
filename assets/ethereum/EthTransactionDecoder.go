package ethereum

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/blocktree/OpenWallet/logger"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-OWCrypt"
	"github.com/bytom/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type EthTxExtPara struct {
}

const (
	ADRESS_STATIS_OVERDATED_TIME = 30
)

type AddressTxStatistic struct {
	Address          string
	TransactionCount *uint64
	LastModifiedTime *time.Time
	Valid            *int //如果valid指针指向的整形为0, 说明该地址已经被清理线程清理
	AddressLocker    *sync.Mutex
	//1. 地址级别, 不可并发广播交易, 造成nonce混乱
	//2. 地址级别, 不可广播, 读取nonce同时进行, 会造成nonce混乱
}

func (this *AddressTxStatistic) UpdateTime() {
	now := time.Now()
	this.LastModifiedTime = &now
}

type EthTransactionDecoder struct {
	AddrTxStatisMap *sync.Map
	DecoderLocker   *sync.Mutex //保护一些全局不可并发的操作, 如对AddrTxStatisMap的初始化
}

func (this *EthTransactionDecoder) GetTransactionCount2(address string) (*AddressTxStatistic, uint64, error) {
	now := time.Now()
	valid := 1
	t := AddressTxStatistic{
		LastModifiedTime: &now,
		AddressLocker:    new(sync.Mutex),
		Valid:            &valid,
		Address:          address,
	}

	v, loaded := this.AddrTxStatisMap.LoadOrStore(address, t)
	//LoadOrStore返回后, AddressLocker加锁前, map中的nonce可能已经被清理了, 需要检查valid是否为1
	txStatis := v.(AddressTxStatistic)
	txStatis.AddressLocker.Lock()
	txStatis.AddressLocker.Unlock()
	if loaded {
		if *txStatis.Valid == 0 {
			return nil, 0, errors.New("the node is busy, try it again later. ")
		}
		txStatis.UpdateTime()
		return &txStatis, *txStatis.TransactionCount, nil
	}
	nonce, err := GetNonceForAddress(address)
	if err != nil {
		openwLogger.Log.Errorf("get nonce for address via rpc failed, err=%v", err)
		return nil, 0, err
	}
	*txStatis.TransactionCount = nonce
	return &txStatis, *txStatis.TransactionCount, nil
}

func (this *EthTransactionDecoder) GetTransactionCount(address string) (uint64, error) {
	if this.AddrTxStatisMap == nil {
		return 0, errors.New("map should be initialized before using.")
	}

	v, exist := this.AddrTxStatisMap.Load(address)
	if !exist {
		return 0, errors.New("no records found to the key passed through.")
	}

	txStatis := v.(AddressTxStatistic)
	return *txStatis.TransactionCount, nil
}

func (this *EthTransactionDecoder) SetTransactionCount(address string, transactionCount uint64) error {
	if this.AddrTxStatisMap == nil {
		return errors.New("map should be initialized before using.")
	}

	v, exist := this.AddrTxStatisMap.Load(address)
	if !exist {
		return errors.New("no records found to the key passed through.")
	}

	now := time.Now()
	valid := 1
	txStatis := AddressTxStatistic{
		TransactionCount: &transactionCount,
		LastModifiedTime: &now,
		AddressLocker:    new(sync.Mutex),
		Valid:            &valid,
		Address:          address,
	}

	if exist {
		txStatis.AddressLocker = v.(AddressTxStatistic).AddressLocker
	} else {
		txStatis.AddressLocker = &sync.Mutex{}
	}

	this.AddrTxStatisMap.Store(address, txStatis)
	return nil
}

func (this *EthTransactionDecoder) RemoveOutdatedAddrStatic() {
	addrStatisList := make([]AddressTxStatistic, 0)
	this.AddrTxStatisMap.Range(func(k, v interface{}) bool {
		addrStatis := v.(AddressTxStatistic)
		if addrStatis.LastModifiedTime.Before(time.Now().Add(-1 * (ADRESS_STATIS_OVERDATED_TIME * time.Minute))) {
			addrStatisList = append(addrStatisList, addrStatis)
		}
		return true
	})

	clear := func(statis *AddressTxStatistic) {
		statis.AddressLocker.Lock()
		defer statis.AddressLocker.Unlock()
		if statis.LastModifiedTime.Before(time.Now().Add(-1 * (ADRESS_STATIS_OVERDATED_TIME * time.Minute))) {
			*statis.Valid = 0
			this.AddrTxStatisMap.Delete(statis)
		}
	}

	for i, _ := range addrStatisList {
		clear(&addrStatisList[i])
	}
}

func (this *EthTransactionDecoder) RunClearAddrStatic() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			this.RemoveOutdatedAddrStatic()
		}
	}()
}

func VerifyRawTransaction(rawTx *openwallet.RawTransaction) error {
	if len(rawTx.To) != 1 {
		openwLogger.Log.Errorf("noly one to address can be set.")
		return errors.New("noly one to address can be set.")
	}

	return nil
}

//CreateRawTransaction 创建交易单
func (this *EthTransactionDecoder) CreateRawTransaction(wrapper *openwallet.WalletWrapper, rawTx *openwallet.RawTransaction) error {
	//check交易交易单基本字段
	err := VerifyRawTransaction(rawTx)
	if err != nil {
		openwLogger.Log.Errorf("Verify raw tx failed, err=%v", err)
		return err
	}
	//获取wallet
	addresses := wrapper.GetWallet().GetAddressesByAccount(rawTx.Account.AccountID)
	if len(addresses) == 0 {
		openwLogger.Log.Errorf("no addresses found in wallet[%v]", rawTx.Account.AccountID)
		return errors.New("no addresses found in wallet.")
	}
	type addrBalance struct {
		address string
		balance *big.Int
		index   int
	}

	addrsBalanceList := make([]addrBalance, 0, len(addresses))
	for i, addr := range addresses {
		balance, err := ConvertToBigInt(addr.Balance, 16)
		if err != nil {
			openwLogger.Log.Errorf("convert address [%v] balance [%v] to big.int failed, err = %v ", addr.Address, addr.Balance, err)
			return err
		}
		addrsBalanceList = append(addrsBalanceList, addrBalance{
			address: addr.Address,
			balance: balance,
			index:   i,
		})
	}

	sort.Slice(addrsBalanceList, func(i int, j int) bool {
		if addrsBalanceList[i].balance.Cmp(addrsBalanceList[j].balance) < 0 {
			return true
		}
		return false
	})

	signatureMap := make(map[string][]openwallet.KeySignature)
	keySignList := make([]openwallet.KeySignature, 0, 1)
	var amountStr, to string
	for k, v := range rawTx.To {
		to = k
		amountStr = v
		break
	}

	amount, err := ConvertToBigInt(amountStr, 16)
	if err != nil {
		openwLogger.Log.Errorf("convert tx amount to big.int failed, err=%v", err)
		return err
	}

	var fee *txFeeInfo
	var data string
	for i, _ := range addrsBalanceList {
		totalAmount := new(big.Int)
		if addrsBalanceList[i].balance.Cmp(amount) > 0 {
			if rawTx.Coin.IsContract {
				data, err = makeERC20TokenTransData(addrsBalanceList[i].address, to, amount)
				if err != nil {
					openwLogger.Log.Errorf("make token transaction data failed, err=%v", err)
					return err
				}
			}

			if rawTx.Coin.IsContract {
				fee, err = GetTransactionFeeEstimated(addrsBalanceList[i].address, to, nil, data)
			} else {
				fee, err = GetTransactionFeeEstimated(addrsBalanceList[i].address, to, amount, data)
			}

			if err != nil {
				openwLogger.Log.Errorf("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrsBalanceList[i].address, to, err)
				return err
			}

			if rawTx.FeeRate != "" {
				fee.GasPrice, err = ConvertToBigInt(rawTx.FeeRate, 16)
				if err != nil {
					openwLogger.Log.Errorf("fee rate passed through error, err=%v", err)
					return err
				}
				fee.CalcFee()
			}

			totalAmount.Add(totalAmount, fee.Fee)

			if addrsBalanceList[i].balance.Cmp(totalAmount) > 0 {
				//fromAddr = addrsBalanceList[i].address
				fromAddr := &openwallet.Address{
					AccountID: addresses[addrsBalanceList[i].index].AccountID,
					Address:   addresses[addrsBalanceList[i].index].Address,
					Alias:     addresses[addrsBalanceList[i].index].Alias,
					Tag:       addresses[addrsBalanceList[i].index].Tag,
					Index:     addresses[addrsBalanceList[i].index].Index,
					HDPath:    addresses[addrsBalanceList[i].index].HDPath,
					WatchOnly: addresses[addrsBalanceList[i].index].WatchOnly,
					Symbol:    addresses[addrsBalanceList[i].index].Symbol,
					Balance:   addresses[addrsBalanceList[i].index].Balance,
					IsMemo:    addresses[addrsBalanceList[i].index].IsMemo,
					Memo:      addresses[addrsBalanceList[i].index].Memo,
					CreatedAt: addresses[addrsBalanceList[i].index].CreatedAt,
				}

				keySignList = append(keySignList, openwallet.KeySignature{
					Address: fromAddr,
				})
				break
			}
		}
	}

	if len(keySignList) != 1 {
		return errors.New("no enough balance address found in wallet. ")
	}

	initTxStaticMap := func() {
		this.DecoderLocker.Lock()
		defer this.DecoderLocker.Unlock()
		if this.AddrTxStatisMap == nil {
			this.AddrTxStatisMap = new(sync.Map)
			this.RunClearAddrStatic()
		}
	}

	initTxStaticMap()

	_, nonce, err := this.GetTransactionCount2(keySignList[0].Address.Address)
	if err != nil {
		openwLogger.Log.Errorf("GetTransactionCount2 failed, err=%v", err)
		return err
	}

	signer := types.NewEIP155Signer(big.NewInt(chainID))
	tx := types.NewTransaction(nonce, ethcommon.HexToAddress(to),
		amount, fee.GasLimit.Uint64(), fee.GasPrice, []byte(data))
	msg := signer.Hash(tx)

	keySignList[0].Nonce = "0x" + strconv.FormatUint(nonce, 16)
	keySignList[0].Message = common.ToHex(msg[:])
	signatureMap[rawTx.Account.AccountID] = keySignList

	rawTx.IsBuilt = true

	return nil
}

/*func (this *EthTransactionDecoder) CreateRawTransaction(wrapper *openwallet.WalletWrapper, rawTx *openwallet.RawTransaction) error {
	if !rawTx.Coin.IsContract {
		return this.CreateSimpleTransaction(wrapper, rawTx)
	}else{

	}

	return nil
}*/

//SignRawTransaction 签名交易单
func (this *EthTransactionDecoder) SignRawTransaction(wrapper *openwallet.WalletWrapper, rawTx *openwallet.RawTransaction) error {
	return nil
}

//SendRawTransaction 广播交易单
func (this *EthTransactionDecoder) SubmitRawTransaction(wrapper *openwallet.WalletWrapper, rawTx *openwallet.RawTransaction) error {

	//check交易交易单基本字段
	err := VerifyRawTransaction(rawTx)
	if err != nil {
		openwLogger.Log.Errorf("Verify raw tx failed, err=%v", err)
		return err
	}
	if len(rawTx.Signatures) != 1 {
		openwLogger.Log.Errorf("len of signatures error. ")
		return errors.New("len of signatures error. ")
	}

	if _, exist := rawTx.Signatures[rawTx.Account.AccountID]; !exist {
		openwLogger.Log.Errorf("wallet[%v] signature not found ", rawTx.Account.AccountID)
		return errors.New("wallet signature not found ")
	}

	from := rawTx.Signatures[rawTx.Account.AccountID][0].Address.Address
	sig := rawTx.Signatures[rawTx.Account.AccountID][0].Signature
	var from, amountStr string
	for k, v := range rawTx.To {
		from = k
		amountStr = v
	}
	amount, err := ConvertToBigInt(amountStr, 10)
	if err != nil {
		openwLogger.Log.Errorf("amount convert to big int failed, err=%v", err)
		return err
	}

	txStatis, _, err := this.GetTransactionCount2(from)
	if err != nil {
		openwLogger.Log.Errorf("get transaction count2 faile, err=%v", err)
		return errors.New("get transaction count2 faile")
	}

	err = func() error {
		txStatis.AddressLocker.Lock()
		defer txStatis.AddressLocker.Unlock()
		nonceSigned, err := strconv.ParseUint(removeOxFromHex(rawTx.Signatures[rawTx.Account.AccountID][0].Nonce),
			16, 64)
		if err != nil {
			openwLogger.Log.Errorf("parse nonce from rawTx failed, err=%v", err)
			return errors.New("parse nonce from rawTx failed. ")
		}
		if nonceSigned != *txStatis.TransactionCount {
			openwLogger.Log.Errorf("nonce out of dated, please try to start ur tx once again. ")
			return errors.New("nonce out of dated, please try to start ur tx once again. ")
		}

		tx := types.NewTransaction(nonceSigned, ethcommon.HexToAddress("0x2a63b2203955b84fefe52baca3881b3614991b34"),
			amount, 121000, gasPrice, nil)

		return nil
	}()

	if err != nil {
		openwLogger.Log.Errorf("send raw transaction failed, err= %v", err)
		return err
	}

	return nil
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (this *EthTransactionDecoder) VerifyRawTransaction(wrapper *openwallet.WalletWrapper, rawTx *openwallet.RawTransaction) error {
	//check交易交易单基本字段
	err := VerifyRawTransaction(rawTx)
	if err != nil {
		openwLogger.Log.Errorf("Verify raw tx failed, err=%v", err)
		return err
	}

	if len(rawTx.Signatures) != 1 {
		openwLogger.Log.Errorf("len of signatures error. ")
		return errors.New("len of signatures error. ")
	}

	if _, exist := rawTx.Signatures[rawTx.Account.AccountID]; !exist {
		openwLogger.Log.Errorf("wallet[%v] signature not found ", rawTx.Account.AccountID)
		return errors.New("wallet signature not found ")
	}
	sig := rawTx.Signatures[rawTx.Account.AccountID][0].Signature
	msg := rawTx.Signatures[rawTx.Account.AccountID][0].Message
	pubkey := rawTx.Signatures[rawTx.Account.AccountID][0].Address.PublicKey
	//curveType := rawTx.Signatures[rawTx.Account.AccountID][0].EccType

	ret := owcrypt.Verify(common.FromHex(pubkey), nil, 0, common.FromHex(msg), 32, common.FromHex(sig),
		owcrypt.ECC_CURVE_SECP256K1|owcrypt.HASH_OUTSIDE_FLAG)
	if ret != owcrypt.SUCCESS {
		errinfo := fmt.Sprintf("verify error, ret:%v\n", "0x"+strconv.FormatUint(uint64(ret), 16))
		fmt.Println(errinfo)
		return errors.New(errinfo)
	}

	return nil
}
