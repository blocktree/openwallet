# Bitcoin Transaction Driver
```
        transaction_test.go测试案例说明
        Test_case1    : from 公钥哈希地址 to  公钥哈希地址
        Test_case2    : from 脚本哈希地址 to  公钥哈希地址
        Test_case3    : from 混合地址    to  公钥哈希地址 
        Test_case4    : from bech32地址  to  公钥哈希地址 
        Test_case5    : from multisig地址  to  公钥哈希地址 
```
## 其他类比特币验证须知
```
        需要修改的数据在btc.go里面
        可针对不同平台进行调整
```
## 当前支持
```
        交易单构建
        交易单签名
        交易单验签

        支持时间锁定和追加交易费的交易单构建
        支持P2PKH类型目标地址
        支持P2SH类型目标地址
        支持P2PKH和P2SH混合类型目标地址
        支持Bech32新型地址
        默认启用隔离认证
        从multisig地址进行支付
```
## TODO
```
        stuff
```
## 用法：
### 创建空交易单 `CreateEmptyRawTransaction`
```
        前置条件:
                获取需要发送的utxo
                获取接收地址和找零地址
                确定手续费以及找零数额
        步骤:
                使用utxo的交易哈希(txid)和输出索引(vout)填充TxIn结构体的Prevout字段
                使用接收地址和找零地址以及对应数额填充TxOut结构体的Address字段和Amount字段
                确认交易单版本号
                确认交易单锁定时间
                确认交易是否可以追加手续费(replaceable)
                获取空交易单
        调用方式:
                CreateRawTransaction([]TxIn, []TxOut, transactionVersion, lockTime, replaceable)
        Tips:
                txid使用的是小端模式，即查询交易时的端序
                交易单版本号为目前的默认版本号02
```
### 创建用于签名的交易单哈希 `CreateRawTransactionHashForSign`
```
        前置条件:
                获取需要发送的utxo
                获取接收地址和找零地址
                确定手续费以及找零数额
        步骤:
                使用utxo的交易哈希(txid)和输出索引(vout)填充TxIn结构体的Prevout字段
                使用接收地址和找零地址以及对应数额填充TxOut结构体的Address字段和Amount字段
                确认交易单版本号
                确认交易单锁定时间
                确认交易是否可以追加手续费(replaceable)
                获取前置交易的锁定脚本
                确认签名类型
                获取交易单
        调用方式:
                CreateRawTransactionHashForSign([]TxIn []TxOut, transactionVersion, lockTime, replaceable, signType)
        Tips:
                txid使用的是小端模式，即查询交易时的端序
                交易单版本号为目前的默认版本号02
                签名类型一般为signAll
```
### 本地交易单签名 `SignEmptyRawTransaction`
```
        前置条件:
                获取空交易单emptyTrans
                获取utxo的锁定脚本以及脚本对应的私钥
        步骤:
                使用锁定脚本与私钥填充TxUnlock结构体
                确定签名类型
        调用方式:
                SignRawTransaction(emptyTrans, []TxUnlock, sigType)
        Tips:
                TxUnlock结构体数组的顺序应该与空交易单的utxo的txid顺序保持一致
                签名类型一般为signAll
```
### 客户端交易单哈希签名 `SignRawTransactionHash`
```
        前置条件:
                获得用于签名的交易单哈希
        步骤:
                获取前置交易的锁定脚本对应的私钥，填充TxUnlock结构体
                获取签名
        调用方式:
                SignRawTransactionHash(transForSig, []TxUnlock)
        Tips:
                TxUnlock结构体数组的顺序应该与空交易单的utxo的txid顺序保持一致
```
### 合并交易单 `InsertSignaturesToEmptyRawTransaction`
```
        前置条件:
                获得空交易单
                获得签名
        步骤:
                合并
        调用方式:
                InsertSignaturesToEmptyRawTransaction(emptyTrans, []SignaturePubkkey)
        Tips:
                签名数据结构体数组的顺序应该与utxo的txid顺序保持一致
```
### 交易单验签 `VerifyRawTransaction`
```
        前置条件:    
                获取签名后的交易单signedTrans
                获取utxo的锁定脚本
        步骤:
                使用锁定脚本填充TxUnlock结构体
        调用方式:
                VerifyRawTransaction(signedTrans, []TxUnlock)
        Tips:
                TxUnlock结构体数组的顺序应该与交易单的utxo的txid顺序保持一致
```
