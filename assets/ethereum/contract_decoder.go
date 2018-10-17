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

func (this *EthContractDecoder) GetTokenBalanceByAddress(contract openwallet.SmartContract, address ...string) ([]*openwallet.TokenBalance, error) {
	threadControl := make(chan int, 20)
	defer close(threadControl)
	resultChan := make(chan *openwallet.TokenBalance, 1024)
	defer close(resultChan)
	done := make(chan int, 1)
	var tokenBalanceList []*openwallet.TokenBalance
	count := len(address)

	go func() {
		log.Debugf("in save thread.")
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

		log.Debugf("in query thread.")
		balanceConfirmed, err := this.wm.WalletClient.ERC20GetAddressBalance2(address, contract.Address, "latest")
		if err != nil {
			log.Errorf("get address[%v] erc20 token balance failed, err=%v", address, err)
			return
		}

		log.Debugf("got balanceConfirmed of [%v] :%v", address, balanceConfirmed)

		balanceAll, err := this.wm.WalletClient.ERC20GetAddressBalance2(address, contract.Address, "pending")
		if err != nil {
			log.Errorf("get address[%v] erc20 token balance failed, err=%v", address, err)
			return
		}

		log.Debugf("got balanceAll of [%v] :%v", address, balanceAll)
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
