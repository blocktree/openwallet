# BlockScanner

区块扫描器接口，由资产适配包实现区块扫描方法，通知已openwallet钱包体系设计的数据到应用层。

## 区块链交易单模型

几乎所有的区块链交易单结构，或采用utxo模型，或采用account模型。（DAG有向无环图未研究）

### UTXO模型

这类模型的区块链不能直接查询到余额的，每笔交易单都以多个输入（vins）和多个输出（vouts）组成。
输入合计数量必须大于或等于输出合计数量。矿工费 = 总输入 - 总输出。
每次构建新交易单，总是依赖地址未使用过的输出，我们称为unspent transaction output（utxo）。
所以要知道地址的余额，就是累计地址所有的utxo中的数量。

大部分区块链采用这种模型的：bitcoin系，bytom，qtum等等。


### account模型

采用这类模型的，可以直接在区块链中查到地址的余额。与utxo模型不同是，交易单只能有一对输入输出。

采用这种模型的：ethereum，nas等等。

## 区块扫描器使用流程

- 获取某个适配的区块扫描器接口。
- 通过AddAddress(address, sourceKey string)添加订阅地址，sourceKey就是账户的标识符，用于汇集提取结果。
- 通过AddObserver(obj BlockScanNotificationObject)添加监听者。
- BlockScanNotificationObject要实现接收数据的接口方法，并保存。
- 执行Run()启动扫描器运行。

## BlockScanNotificationObject接口说明

实现BlockScanNotificationObject接口。

- BlockScanNotify获取新区块。
- BlockExtractDataNotify交易单提取结果

```go

    //BlockScanNotify 新区块扫描完成通知
	BlockScanNotify(header *BlockHeader) error

	//BlockExtractDataNotify 区块提取结果通知
	BlockExtractDataNotify(sourceKey string, data *TxExtractData) error

```

### TxExtractData

blockscanner会从每笔交易单中，根据订阅地址提取出来，并以地址绑定的sourceKey汇集在一个提取结果中。
sourceKey其实就是账户的唯一标识符。

- []*TxInput，utxo模型的可以有多个输入，account模型只有1个。
- []*TxOutPut，utxo模型的可以有多个输出，account模型只有1个。
- *Transaction，openwallet的钱包模型是以账户为单位的，账户存在多个地址可能，所以账户在该交易单的转账数量是以[]*TxOutPut的总数量 - []*TxInput的总数量计算出来。









