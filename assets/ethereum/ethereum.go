package ethereum

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"github.com/blocktree/OpenWallet/console"
	"github.com/bndr/gotabulate"
	"github.com/shopspring/decimal"
)

type WalletManager struct{}

//初始化配置流程
func (this *WalletManager) InitConfigFlow() error {
	file := filepath.Join(configFilePath, configFileName)
	fmt.Printf("You can run 'vim %s' to edit wallet's config.\n", file)
	return nil
}

//查看配置信息
func (this *WalletManager) ShowConfig() error {
	return printConfig()
}

//创建钱包流程
func (this *WalletManager) CreateWalletFlow() error {
	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	// 等待用户输入钱包名字
	name, err := console.InputText("Enter wallet's name: ", true)

	// 等待用户输入密码
	password, err := console.InputPassword(true, 8)

	_, keyFile, err := CreateNewWallet(name, password)
	if err != nil {
		return err
	}

	fmt.Printf("\n")
	fmt.Printf("Wallet create successfully, key path: %s\n", keyFile)

	return nil
}

//打印钱包列表
func printWalletList(list []*Wallet) {

	tableInfo := make([][]interface{}, 0)

	for i, w := range list {

		tableInfo = append(tableInfo, []interface{}{
			i, w.WalletID, w.Alias, w.Balance,
		})
	}

	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"No.", "ID", "Name", "Balance"})

	//打印信息
	fmt.Println(t.Render("simple"))

}

//创建地址流程
func (this *WalletManager) CreateAddressFlow() error {
	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	//查询所有钱包信息
	wallets, err := GetWalletList()
	if err != nil {
		fmt.Printf("The node did not create any wallet!\n")
		return err
	}

	//打印钱包
	printWalletList(wallets)

	fmt.Printf("[Please select a wallet account to create address] \n")

	//选择钱包
	num, err := console.InputNumber("Enter wallet No. : ", true)
	if err != nil {
		return err
	}

	if int(num) >= len(wallets) {
		return errors.New("Input number is out of index! ")
	}

	account := wallets[num]

	// 输入地址数量
	count, err := console.InputNumber("Enter the number of addresses you want: ", false)
	if err != nil {
		return err
	}

	if count > maxAddresNum {
		return errors.New(fmt.Sprintf("The number of addresses can not exceed %d\n", maxAddresNum))
	}

	//输入密码
	password, err := console.InputPassword(false, 8)

	log.Printf("Start batch creation\n")
	log.Printf("-------------------------------------------------\n")

	err = CreateBatchAddress(account.WalletID, password, count)
	if err != nil {
		return err
	}

	log.Printf("-------------------------------------------------\n")
	log.Printf("All addresses have created, file path:%s\n", EthereumKeyPath)

	return nil
}

//汇总钱包流程
func (this *WalletManager) SummaryFollow() error {
	return nil
}

//备份钱包流程
func (this *WalletManager) BackupWalletFlow() error {
	return nil
}

//查看钱包列表，显示信息
func (this *WalletManager) GetWalletList() error {
	return nil
}

//发送交易
func (this *WalletManager) TransferFlow() error {
	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	list, err := GetWalletList()
	if err != nil {
		return err
	}

	//打印钱包列表
	printWalletList(list)

	fmt.Printf("[Please select a wallet to send transaction] \n")

	//选择钱包
	num, err := console.InputNumber("Enter wallet No. : ", true)
	if err != nil {
		return err
	}

	if int(num) >= len(list) {
		return errors.New("Input number is out of index! ")
	}

	wallet := list[num]

	// 等待用户输入发送数量
	amount, err := console.InputRealNumber("Enter amount to send: ", true)
	if err != nil {
		return err
	}

	atculAmount, _ := decimal.NewFromString(amount)
	balance, _ := decimal.NewFromString(wallet.Balance)

	if atculAmount.GreaterThan(balance) {
		return errors.New("Input amount is greater than balance! ")
	}

	// 等待用户输入发送数量
	receiver, err := console.InputText("Enter receiver address: ", true)
	if err != nil {
		return err
	}

	fmt.Println("receiver: ", receiver)

	return nil
}

//恢复钱包
func (this *WalletManager) RestoreWalletFlow() error {
	return nil
}
