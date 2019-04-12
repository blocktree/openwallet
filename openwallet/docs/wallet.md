# 钱包管理模型开发

钱包与资产账户是一对多的关系，资产账户与地址是一对多的关系。
openw包实现了一个单机版的钱包管理模型。其集成了AssetsAdapter接口，可用于测试整个区块链资产适配器的功能。
钱包管理系统开发者可参考openw包，基于openwallet框架实现自己的钱包管理系统。

## WalletDAI实现

WalletDAI是钱包数据访问接口。该接口由钱包管理系统实现，这样AssetsAdapter便可通过该接口查询钱包相关数据，主要完成交易单构建工作。

如下接口定义：

```go


//WalletDAI 钱包数据访问接口
type WalletDAI interface {
	//获取当前钱包
	GetWallet() *Wallet
	//根据walletID查询钱包
	GetWalletByID(walletID string) (*Wallet, error)

	//获取单个资产账户
	GetAssetsAccountInfo(accountID string) (*AssetsAccount, error)
	//查询资产账户列表
	GetAssetsAccountList(offset, limit int, cols ...interface{}) ([]*AssetsAccount, error)
	//根据地址查询资产账户
	GetAssetsAccountByAddress(address string) (*AssetsAccount, error)

	//获取单个地址
	GetAddress(address string) (*Address, error)
	//查询地址列表
	GetAddressList(offset, limit int, cols ...interface{}) ([]*Address, error)
	//设置地址的扩展字段
	SetAddressExtParam(address string, key string, val interface{}) error
	//获取地址的扩展字段
	GetAddressExtParam(address string, key string) (interface{}, error)

	//解锁钱包，指定时间内免密
	UnlockWallet(password string, time time.Duration) error
	//获取钱包HDKey
	HDKey(password ...string) (*hdkeystore.HDKey, error)
}

```