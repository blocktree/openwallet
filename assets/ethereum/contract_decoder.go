package ethereum

import (
	"errors"
	"math/big"

	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
)

type EthContractDecoder struct {
	wm *WalletManager
}

type AddrBalance struct {
	Address      string
	Balance      *big.Int
	TokenBalance *big.Int
	Index        int
}

func (this *AddrBalance) SetTokenBalance(b *big.Int) {
	this.TokenBalance = b
}

func (this *AddrBalance) GetAddress() string {
	return this.Address
}

func (this *AddrBalance) ValidTokenBalance() bool {
	if this.Balance == nil {
		return false
	}
	return true
}

type AddrBalanceInf interface {
	SetTokenBalance(b *big.Int)
	GetAddress() string
	ValidTokenBalance() bool
}

func (this *WalletManager) GetBalanceByAddress(address ...string) ([]*openwallet.Balance, error) {
	type addressBalance struct {
		Address string
		Index   uint64
		Balance *openwallet.Balance
	}

	threadControl := make(chan int, 20)
	defer close(threadControl)
	resultChan := make(chan *addressBalance, 1024)
	defer close(resultChan)
	done := make(chan int, 1)
	count := len(address)
	resultBalance := make([]*openwallet.Balance, count)
	resultSaveFailed := false
	//save result
	go func() {
		for i := 0; i < count; i++ {
			addr := <-resultChan
			if addr.Balance != nil {
				resultBalance[addr.Index] = addr.Balance
			} else {
				resultSaveFailed = true
			}
		}
		done <- 1
	}()

	query := func(addr *addressBalance) {
		threadControl <- 1
		defer func() {
			resultChan <- addr
			<-threadControl
		}()

		balanceConfirmed, err := this.WalletClient.GetAddrBalance2(appendOxToAddress(addr.Address), "latest")
		if err != nil {
			log.Error("get address[", addr.Address, "] balance failed, err=", err)
			return
		}

		balanceAll, err := this.WalletClient.GetAddrBalance2(appendOxToAddress(addr.Address), "pending")
		if err != nil {
			log.Errorf("get address[%v] erc20 token balance failed, err=%v", address, err)
			return
		}

		//		log.Debugf("got balanceAll of [%v] :%v", address, balanceAll)
		balanceUnconfirmed := big.NewInt(0)
		balanceUnconfirmed.Sub(balanceAll, balanceConfirmed)

		balance := &openwallet.Balance{
			Symbol:  this.Symbol(),
			Address: addr.Address,
		}
		confirmed, err := ConverWeiStringToEthDecimal(balanceConfirmed.String())
		if err != nil {
			log.Errorf("ConverWeiStringToEthDecimal confirmed balance failed, err=%v", err)
			return
		}
		all, err := ConverWeiStringToEthDecimal(balanceAll.String())
		if err != nil {
			log.Errorf("ConverWeiStringToEthDecimal all balance failed, err=%v", err)
			return
		}

		unconfirmed, err := ConverWeiStringToEthDecimal(balanceUnconfirmed.String())
		if err != nil {
			log.Errorf("ConverWeiStringToEthDecimal unconfirmed balance failed, err=%v", err)
			return
		}

		balance.Balance = all.String()
		balance.UnconfirmBalance = unconfirmed.String()
		balance.ConfirmBalance = confirmed.String()
		addr.Balance = balance
	}

	for i, _ := range address {
		addrbl := &addressBalance{
			Address: address[i],
			Index:   uint64(i),
		}
		go query(addrbl)
	}

	<-done
	if resultSaveFailed {
		return nil, errors.New("get balance of addresses failed.")
	}
	return resultBalance, nil
}

func (this *WalletManager) GetTokenBalanceByAddress(contractAddr string, addrs ...AddrBalanceInf) error {
	threadControl := make(chan int, 20)
	defer close(threadControl)
	resultChan := make(chan AddrBalanceInf, 1024)
	defer close(resultChan)
	done := make(chan int, 1)
	count := len(addrs)
	var err error

	go func() {
		log.Debugf("in save thread.")
		for i := 0; i < count; i++ {
			addr := <-resultChan
			if !addr.ValidTokenBalance() {
				err = errors.New("query token balance failed")
			}
		}
		done <- 1
	}()

	queryBalance := func(addr AddrBalanceInf) {
		threadControl <- 1
		defer func() {
			resultChan <- addr
			<-threadControl
		}()

		balance, err := this.WalletClient.ERC20GetAddressBalance(addr.GetAddress(), contractAddr)
		if err != nil {
			log.Errorf("get address[%v] erc20 token balance failed, err=%v", addr.GetAddress(), err)
			return
		}

		addr.SetTokenBalance(balance)
	}

	for i, _ := range addrs {
		go queryBalance(addrs[i])
	}

	<-done

	return err
}

func (this *EthContractDecoder) GetTokenBalanceByAddress(contract openwallet.SmartContract, address ...string) ([]*openwallet.TokenBalance, error) {
	threadControl := make(chan int, 20)
	defer close(threadControl)
	resultChan := make(chan *openwallet.TokenBalance, 1024)
	defer close(resultChan)
	done := make(chan int, 1)
	var tokenBalanceList []*openwallet.TokenBalance
	count := len(address)

	go func() {
		//		log.Debugf("in save thread.")
		for i := 0; i < count; i++ {
			balance := <-resultChan
			if balance != nil {
				tokenBalanceList = append(tokenBalanceList, balance)
			}
			log.Debugf("got one balance.")
		}
		done <- 1
	}()

	queryBalance := func(address string) {
		threadControl <- 1
		var balance *openwallet.TokenBalance
		defer func() {
			resultChan <- balance
			<-threadControl
		}()

		//		log.Debugf("in query thread.")
		balanceConfirmed, err := this.wm.WalletClient.ERC20GetAddressBalance2(address, contract.Address, "latest")
		if err != nil {
			log.Errorf("get address[%v] erc20 token balance failed, err=%v", address, err)
			return
		}

		//		log.Debugf("got balanceConfirmed of [%v] :%v", address, balanceConfirmed)

		balanceAll, err := this.wm.WalletClient.ERC20GetAddressBalance2(address, contract.Address, "pending")
		if err != nil {
			log.Errorf("get address[%v] erc20 token balance failed, err=%v", address, err)
			return
		}

		//		log.Debugf("got balanceAll of [%v] :%v", address, balanceAll)
		balanceUnconfirmed := big.NewInt(0)
		balanceUnconfirmed.Sub(balanceAll, balanceConfirmed)

		balance = &openwallet.TokenBalance{
			Contract: &contract,
			Balance: &openwallet.Balance{
				Address:          address,
				Symbol:           contract.Symbol,
				Balance:          balanceAll.String(),
				ConfirmBalance:   balanceConfirmed.String(),
				UnconfirmBalance: balanceUnconfirmed.String(),
			},
		}
	}

	for i, _ := range address {
		go queryBalance(address[i])
	}

	<-done

	if len(tokenBalanceList) != count {
		log.Error("unknown errors occurred .")
		return nil, errors.New("unknown errors occurred .")
	}
	return tokenBalanceList, nil
}
