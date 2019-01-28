package ethereum

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tool "github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/keystore"
	"github.com/blocktree/OpenWallet/log"
	"github.com/btcsuite/btcutil/hdkeychain"
	ethKStore "github.com/ethereum/go-ethereum/accounts/keystore"
)

//CreateNewWallet 创建钱包
//过时的函数,不推荐使用
func (this *WalletManager) CreateNewWallet(name, password string) (*Wallet, string, error) {

	//检查钱包名是否存在
	wallets, err := GetWalletKeys(this.GetConfig().KeyDir)
	if err != nil {
		return nil, "", errors.New(fmt.Sprintf("get wallet keys failed, err = %v", err))
	}

	for _, w := range wallets {
		if w.Alias == name {
			return nil, "", errors.New("The wallet's alias is duplicated!")
		}
	}

	//fmt.Printf("Verify password in bitcoin-core wallet...\n")
	seed, err := hdkeychain.GenerateSeed(32)
	if err != nil {
		return nil, "", err
	}

	extSeed, err := keystore.GetExtendSeed(seed, MasterKey)
	if err != nil {
		return nil, "", err
	}

	key, keyFile, err := keystore.StoreHDKeyWithSeed(this.GetConfig().KeyDir, name, password, extSeed, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return nil, "", err
	}

	w := Wallet{WalletID: key.RootId, Alias: key.Alias}

	return &w, keyFile, nil
}

//HDKey 获取钱包密钥，需要密码
//已经过期, 不推荐使用
func (w *Wallet) HDKey(password string, s *keystore.HDKeystore) (*keystore.HDKey, error) {
	fmt.Println("w.KeyFile:", w.KeyFile)
	key, err := s.GetKey(w.WalletID, w.KeyFile, password)
	if err != nil {
		return nil, err
	}
	return key, err
}

func (this *WalletManager) CreateBatchAddress(name, password string, count uint64) error {
	//读取钱包
	w, err := this.GetWalletInfo(this.GetConfig().KeyDir, this.GetConfig().DbPath, name)
	if err != nil {
		this.Log.Errorf(fmt.Sprintf("get wallet info, err=%v\n", err))
		return err
	}

	//验证钱包
	keyroot, err := w.HDKey(password, this.StorageOld)
	if err != nil {
		this.Log.Errorf(fmt.Sprintf("get HDkey, err=%v\n", err))
		return err
	}

	timestamp := uint64(time.Now().Unix())

	db, err := w.OpenDB(this.GetConfig().DbPath)
	if err != nil {
		this.Log.Errorf(fmt.Sprintf("open db, err=%v\n", err))
		return err
	}
	defer db.Close()

	ethKeyStore := ethKStore.NewKeyStore(this.GetConfig().EthereumKeyPath, ethKStore.StandardScryptN, ethKStore.StandardScryptP)

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	errcount := uint64(0)
	for i := uint64(0); i < count && errcount < count; {
		// 生成地址
		keyCombo, address, err := CreateNewPrivateKey(keyroot, timestamp, i)
		if err != nil {
			log.Error("Create new privKey failed unexpected error: ", err)
			errcount++
			continue
		}
		_, err = ethKeyStore.NewAccountForWalletBT(keyCombo, password)
		if err != nil {
			this.Log.Errorf("NewAccountForWalletBT failed, err = %v", err)
			errcount++
			continue
		}
		//ethKeyStore.StoreNewKeyForWalletBT(ethKeyStore, keyCombo, DefaultPasswordForEthKey)

		err = tx.Save(address)
		if err != nil {
			log.Error("save address for wallet failed, err=", err)
			errcount++
			continue
		}
		i++
	}

	return tx.Commit()
}

func (this *WalletManager) SendTransaction(wallet *Wallet, to string, amount *big.Int, password string, feesInSender bool) ([]string, error) {
	var txIds []string

	err := this.UnlockWallet(wallet, password)
	if err != nil {
		this.Log.Errorf("unlock wallet [%v]. failed, err=%v", wallet.WalletID, err)
		return nil, err
	}

	addrs, err := this.GetAddressesByWallet(this.GetConfig().DbPath, wallet)
	if err != nil {
		this.Log.Errorf("failed to get addresses from db, err = %v", err)
		return nil, err
	}

	sort.Sort(&AddrVec{addrs: addrs})
	//检查下地址排序是否正确, 仅用于测试
	for _, theAddr := range addrs {
		fmt.Println("theAddr[", theAddr.Address, "]:", theAddr.balance)
	}
	//amountLeft := *amount
	for i := len(addrs) - 1; i >= 0 && amount.Cmp(big.NewInt(0)) > 0; i-- {
		var amountToSend big.Int
		var fee *txFeeInfo

		fmt.Println("amount remained:", amount.String())
		//空账户直接跳过
		//if addrs[i].balance.Cmp(big.NewInt(0)) == 0 {
		//	this.Log.Infof("skip the address[%v] with 0 balance. ", addrs[i].Address)
		//	continue
		//}

		//如果该地址的余额足够支付转账
		if addrs[i].balance.Cmp(amount) >= 0 {
			amountToSend = *amount
			fee, err = this.GetSimpleTransactionFeeEstimated(addrs[i].Address, to, &amountToSend)
			if err != nil {
				this.Log.Errorf("%v", err)
				continue
			}

			balanceLeft := *addrs[i].balance
			balanceLeft.Sub(&balanceLeft, fee.Fee)

			//灰尘账户, 余额不足以发起一次transaction
			//fmt.Println("amount to send ignore fee:", amountToSend.String())
			if balanceLeft.Cmp(big.NewInt(0)) < 0 {
				errinfo := fmt.Sprintf("[%v] is a dust address, will skip. ", addrs[i].Address)
				this.Log.Errorf(errinfo)
				continue
			}

			//如果改地址的余额除去手续费后, 不足以支付转账, set 转账金额 = 账户余额 - 手续费
			if balanceLeft.Cmp(&amountToSend) < 0 {
				amountToSend = balanceLeft
				//fmt.Println("amount to send plus fee:", amountToSend.String())
			}

		} else {
			amountToSend = *addrs[i].balance
			fee, err = this.GetSimpleTransactionFeeEstimated(addrs[i].Address, to, &amountToSend)
			if err != nil {
				this.Log.Errorf("%v", err)
				continue
			}

			//灰尘账户, 余额不足以发起一次transaction
			if amountToSend.Cmp(fee.Fee) <= 0 {
				errinfo := fmt.Sprintf("[%v] is a dust address, will skip. ", addrs[i].Address)
				this.Log.Errorf(errinfo)
				continue
			}

			//fmt.Println("amount to send without fee, ", amountToSend.String(), " , fee:", fee.Fee.String())
			amountToSend.Sub(&amountToSend, fee.Fee)
			//fmt.Println("amount to send applied fee, ", amountToSend.String())
		}

		txid, err := this.SendTransactionToAddr(makeSimpleTransactionPara(addrs[i], to, &amountToSend, password, fee))
		if err != nil {
			this.Log.Errorf("SendTransactionToAddr failed, err=%v", err)
			if txid == "" {
				continue //txIds = append(txIds, txid)
			}
		}

		txIds = append(txIds, txid)
		amount.Sub(amount, &amountToSend)
	}

	return txIds, nil
}

func (this *WalletManager) BackupWallet(newBackupDir string, wallet *Wallet, password string) (string, error) {
	/*w, err := GetWalletInfo(wallet.WalletID)
	if err != nil {
		return "", err
	}*/

	err := this.UnlockWallet(wallet, password)
	if err != nil {
		this.Log.Errorf("unlock wallet failed, err=%v", err)
		return "", err
	}

	addressMap := make(map[string]int)
	files := make([]string, 0)

	//创建备份文件夹
	//newBackupDir := filepath.Join(BackupDir, w.FileName()+"-"+common.TimeFormat("20060102150405"))
	file.MkdirAll(newBackupDir)

	addrs, err := this.GetAddressesByWallet(this.GetConfig().DbPath, wallet)
	if err != nil {
		this.Log.Errorf("get addresses by wallet failed, err = %v", err)
		return "", err
	}

	//搜索出绑定钱包的地址
	for _, addr := range addrs {
		address := addr.Address
		address = strings.Trim(address, " ")
		address = strings.ToLower(address)
		addressMap[address] = 1
	}

	/*for k, v := range addressMap {
		fmt.Println("address[", k, "], exist[", v, "]")
	}*/

	rd, err := ioutil.ReadDir(this.GetConfig().EthereumKeyPath)
	if err != nil {
		this.Log.Errorf("open ethereum key path [%v] failed, err=%v", this.GetConfig().EthereumKeyPath, err)
		return "", err
	}

	//fmt.Println("rd length:", len(rd))
	for _, fi := range rd {
		if skipKeyFile(fi) {
			continue
		}

		//fmt.Println("file name:", fi.Name())
		parts := strings.Split(fi.Name(), "--")
		l := len(parts)
		if l == 0 {
			continue
		}

		theAddr := "0x" + parts[l-1]
		//fmt.Println("loop addr:", theAddr)
		if _, exist := addressMap[theAddr]; exist {
			files = append(files, fi.Name())
		} /*else {
			fmt.Println("address[", theAddr, "], exist[", addressMap[theAddr], "]")
		}*/
	}

	/*for _, keyfile := range files {
		cmd := "cp " + EthereumKeyPath + "/" + keyfile + " " + newBackupDir
		_, err = exec_shell(cmd)
		if err != nil {
			this.Log.Errorf("backup key faile failed, err = ", err)
			return "", err
		}
	}*/

	//fmt.Println("file list length:", len(files))

	//备份该钱包下的所有地址
	for _, keyfile := range files {
		err := file.Copy(this.GetConfig().EthereumKeyPath+"/"+keyfile, newBackupDir+"/")
		if err != nil {
			this.Log.Errorf("backup key faile failed, err = %v", err)
			return "", err
		}
	}

	//备份钱包key文件
	file.Copy(filepath.Join(this.GetConfig().KeyDir, wallet.FileName()+".key"), newBackupDir)

	//备份地址数据库
	file.Copy(filepath.Join(this.GetConfig().DbPath, wallet.FileName()+".db"), newBackupDir)

	return newBackupDir, nil
}

//RestoreWallet 恢复钱包
func (this *WalletManager) RestoreWallet(keyFile string, password string) error {
	fmt.Printf("Validating key file... \n")

	finfo, err := os.Stat(keyFile)
	if err != nil || !finfo.IsDir() {
		errinfo := fmt.Sprintf("stat file[%v] failed, err = %v\n", keyFile, err)
		this.Log.Errorf(errinfo)
		return err
	}
	/*parts := strings.Split(keyFile, "\\") //filepath.SplitList(keyFile)
	l := len(parts)
	if l == 0 {
		errinfo := fmt.Sprintf("wrong keyFile[%v] passed through...", keyFile)
		this.Log.Errorf(errinfo)
		return errors.New(errinfo)
	}
	*/
	dirName := finfo.Name()

	fmt.Println("dirName:", dirName)
	parts := strings.Split(dirName, "-")
	if len(parts) != 3 {
		errinfo := fmt.Sprintf("invalid directory name[%v] ", dirName)
		this.Log.Errorf(errinfo)
		return errors.New(errinfo)
	}

	_, err = time.ParseInLocation(TIME_POSTFIX, parts[2], time.Local)
	if err != nil {
		errinfo := fmt.Sprintf("check directory name[%v] time format failed ", dirName)
		this.Log.Errorf(errinfo)
		return errors.New(errinfo)
	}

	walletId := parts[1]
	//检查备份路径下key文件的密码
	walletKeyBackupPath := keyFile + "/" + parts[0] + "-" + walletId
	walletBackup, err := GetWalletKey(walletKeyBackupPath)
	if err != nil {
		this.Log.Errorf("parse the key file [%v] failed, err= %v.", walletKeyBackupPath, err)
		return err
	}
	err = verifyBackupWallet(walletBackup, keyFile, password)
	if err != nil {
		this.Log.Errorf("verify the backup wallet [%v] password failed, err= %v.", walletKeyBackupPath, err)
		return err
	}

	walletexist, err := this.GetWalletInfo(this.GetConfig().KeyDir, this.GetConfig().DbPath, walletId)
	if err != nil && err.Error() != WALLET_NOT_EXIST_ERR {
		errinfo := fmt.Sprintf("get wallet [%v] info failed, err = %v ", walletId, err)
		this.Log.Errorf(errinfo)
		return errors.New(errinfo)
	} else if err == nil {
		err = this.UnlockWallet(walletexist, password)
		if err != nil {
			this.Log.Errorf("unlock the existing wallet [%v] password failed, err= %v.", walletKeyBackupPath, err)
			return err
		}

		newBackupDir := filepath.Join(this.GetConfig().BackupDir+"/restore", walletexist.FileName()+"-"+tool.TimeFormat(TIME_POSTFIX))
		_, err := this.BackupWallet(newBackupDir, walletexist, password)
		if err != nil {
			errinfo := fmt.Sprintf("backup wallet[%v] before restore failed,err = %v ", walletexist.WalletID, err)
			this.Log.Errorf(errinfo)
			return errors.New(errinfo)
		}
	} else {

	}

	files, err := ioutil.ReadDir(keyFile)
	if err != nil {
		errinfo := fmt.Sprintf("open directory [%v] failed, err = %v ", keyFile, err)
		this.Log.Errorf(errinfo)
		return errors.New(errinfo)
	}

	filesMap := make(map[string]int)
	for _, fi := range files {
		// Skip any non-key files from the folder
		if skipKeyFile(fi) {
			continue
		}

		//		fmt.Println("filename:", fi.Name())
		if strings.Index(fi.Name(), "--") != -1 && strings.Index(fi.Name(), "UTC") != -1 {
			parts = strings.Split(fi.Name(), "--")
			if len(parts) == 0 {
				//				fmt.Println("1. skipped filename:", fi.Name())
				continue
			}
			if len(parts[len(parts)-1]) != len("50068fd632c1a6e6c5bd407b4ccf8861a589e776") {
				//				fmt.Println("2. skipped filename:", fi.Name())
				continue
			}
			filesMap[fi.Name()] = BACKUP_FILE_TYPE_ADDRESS
		} else if strings.Index(fi.Name(), ".key") != -1 && strings.Index(fi.Name(), "-") != -1 {
			filesMap[fi.Name()] = BACKUP_FILE_TYPE_WALLET_KEY
			//			fmt.Println("key filename:", fi.Name())
		} else if strings.Index(fi.Name(), ".db") != -1 && strings.Index(fi.Name(), "-") != -1 {
			filesMap[fi.Name()] = BACKUP_FILE_TYPE_WALLET_DB
			//			fmt.Println("db filename:", fi.Name())
		} /*else {
			fmt.Println("skipped filename:", fi.Name())
			continue
		}*/
	}

	for filename, filetype := range filesMap {
		src := keyFile + "/" + filename
		var dst string
		//		fmt.Println("src:", src)
		if filetype == BACKUP_FILE_TYPE_ADDRESS {
			dst = this.GetConfig().EthereumKeyPath + "/"
		} else if filetype == BACKUP_FILE_TYPE_WALLET_DB {
			dst = this.GetConfig().DbPath + "/"
			//			fmt.Println("db file:", filename)
		} else if filetype == BACKUP_FILE_TYPE_WALLET_KEY {
			dst = this.GetConfig().KeyDir + "/"
			//			fmt.Println("key file:", filename)
		} else {
			continue
		}

		err = file.Copy(src, dst)
		if err != nil {
			errinfo := fmt.Sprintf("copy file from [%v] to [%v] failed, err = %v", src, dst, err)
			this.Log.Errorf(errinfo)
			return errors.New(errinfo)
		}
	}

	return nil
}

func (this *WalletManager) ERC20SendTransaction(wallet *Wallet, to string, amount *big.Int, password string, feesInSender bool) ([]string, error) {
	var txIds []string

	err := this.UnlockWallet(wallet, password)
	if err != nil {
		this.Log.Errorf("unlock wallet [%v]. failed, err=%v", wallet.WalletID, err)
		return nil, err
	}

	addrs, err := this.ERC20GetAddressesByWallet(this.GetConfig().DbPath, wallet)
	if err != nil {
		this.Log.Errorf("failed to get addresses from db, err = %v", err)
		return nil, err
	}

	sort.Sort(&TokenAddrVec{addrs: addrs})
	//检查下地址排序是否正确, 仅用于测试
	/*for _, theAddr := range addrs {
		fmt.Println("theAddr[", theAddr.Address, "]:", theAddr.tokenBalance)
	}*/

	for i := len(addrs) - 1; i >= 0 && amount.Cmp(big.NewInt(0)) > 0; i-- {
		var fee *txFeeInfo
		var amountToSend big.Int
		fmt.Println("amount remained:", amount.String())
		//空的token账户直接跳过
		//if addrs[i].tokenBalance.Cmp(big.NewInt(0)) == 0 {
		//	this.Log.Infof("skip the address[%v] with 0 balance. ", addrs[i].Address)
		//	continue
		//}

		if addrs[i].tokenBalance.Cmp(amount) >= 0 {
			amountToSend = *amount

		} else {
			amountToSend = *addrs[i].tokenBalance
		}

		dataPara, err := makeERC20TokenTransData(wallet.erc20Token.Address, to, &amountToSend)
		if err != nil {
			this.Log.Errorf("make token transaction data failed, err=%v", err)
			return nil, err
		}
		fee, err = this.GetERC20TokenTransactionFeeEstimated(addrs[i].Address, wallet.erc20Token.Address, dataPara)
		if err != nil {
			this.Log.Errorf("get erc token transaction fee estimated failed, err = %v", err)
			continue
		}

		if addrs[i].balance.Cmp(fee.Fee) < 0 {
			this.Log.Errorf("address[%v] cannot afford a token transfer with a fee [%v]", addrs[i].Address, fee.Fee)
			continue
		}

		txid, err := this.SendTransactionToAddr(makeERC20TokenTransactionPara(addrs[i], wallet.erc20Token.Address, dataPara, password, fee))
		if err != nil {
			this.Log.Errorf("SendTransactionToAddr failed, err=%v", err)
			if txid == "" {
				continue //txIds = append(txIds, txid)
			}
		}

		txIds = append(txIds, txid)
		amount.Sub(amount, &amountToSend)
	}

	return txIds, nil
}
