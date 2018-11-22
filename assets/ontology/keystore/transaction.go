package keystore

import (
	"fmt"
	"strconv"
)

func transfer() {

}

func GetAccountMulti(wallet *ClientImpl, passwd []byte, accAddr string) (*Account, error) {
	//Address maybe address in base58, label or index
	if accAddr == "" {
		defAcc, err := wallet.GetDefaultAccount(passwd)
		if err != nil {
			return nil, err
		}
		return defAcc, nil
	}
	acc, err := wallet.GetAccountByAddress(accAddr, passwd)
	if err != nil {
		return nil, fmt.Errorf("getAccountByAddress:%s error:%s", accAddr, err)
	}
	if acc != nil {
		return acc, nil
	}
	acc, err = wallet.GetAccountByLabel(accAddr, passwd)
	if err != nil {
		return nil, fmt.Errorf("getAccountByLabel:%s error:%s", accAddr, err)
	}
	if acc != nil {
		return acc, nil
	}
	index, err := strconv.ParseInt(accAddr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("cannot get account by:%s", accAddr)
	}
	acc, err = wallet.GetAccountByIndex(int(index), passwd)
	if err != nil {
		return nil, fmt.Errorf("getAccountByIndex:%d error:%s", index, err)
	}
	if acc != nil {
		return acc, nil
	}
	return nil, fmt.Errorf("cannot get account by:%s", accAddr)
}
