package main

import (
	"fmt"
	//"gopkg.in/urfave/cli.v1"

	"os"
	"path/filepath"
	"strings"

	_ "github.com/blocktree/OpenWallet/testethereum/environment"
	"github.com/blocktree/OpenWallet/testethereum/tech"
)

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0])) //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if err != nil {
		//log.Fatal(err)
		fmt.Println(err)
	}
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
}

func main() {
	/*app := cli.NewApp()
	app.Name = "boom"
	app.Usage = "make an explosive entrance"
	app.Action = func(c *cli.Context) error {
	  fmt.Println("boom! I say!")
	  return nil
	}
	err := app.Run(os.Args)
	if err != nil {
	  fmt.Println(err)
	}*/

	//fmt.Println("change dir err:",err)
	//tech.TestNewWallet()
	//tech.TestBatchCreateAddr()p
	//tech.TestBitInt()
	//tech.TestTransferFlow()
	//tech.TestSummaryFlow()
	//tech.TestBackupWallet()
	//tech.TestRestoreWallet()
	//tech.TestConfigErcToken()
	//tech.TestERC20TokenTransfer()
	//tech.TestERC20TokenSummary()
	//tech.TestBigInt()
	//tech.TestDiffer()
	//tech.TestGetFuncAndFileName()
	//tech.PrepareTestForBlockScan()
	//tech.TestErc20GetEventLog()
	//tech.TestDbInf()
	//tech.TestBlockScan()
	//tech.TestBlockScanWhenFork()
	//tech.TestEIP155Signing()
	//tech.TestAddr()
	//tech.TestOWCrypt_sign()
	//tech.TestSlice()
	//tech.TestMap()
	//tech.TestEthereumSigningFunc()
	//tech.TestGetNonce()
	//tech.TestOfficialVerify()

	//tech.TestFrameWalletManager_CreateWallet()
	//tech.TestWalletManager_CreateAssetsAccount()
	//tech.TestFrameWalletManager_CreateAddress()
	//tech.ListInfoFromDb()
	//tech.TestGetPrivateKey()

	//tech.TestSendRawTransactionFromGethAccount()
	//tech.TestSendRawTransaction()
	//tech.TestFloat()

	//tech.TestWalletManager_CreateTransaction()
	//tech.TestWalletManager_SignTransaction()
	//tech.TestWalletManager_VerifyTransaction()
	//tech.TestWalletManager_SubmitTransaction()

	//tech.TestSlice2()
	//tech.TestStringAndSlice()

	//tech.TestCreateWallet2345()
	//tech.TestBatchCreateAddr2345()
	//tech.DumpEtc2345WalletDb()
	//tech.TestTransferFlow2345()
	//tech.TestInitCongfig2345()

	//tech.TestWalletLog()
	//tech.TestTokenDecode()
	//tech.TestTokenBalance()

	tech.TestWalletManager_SubmitTokenTransaction()
}
