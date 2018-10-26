package btcLikeTxDriver

import (
	"encoding/hex"
	"github.com/shopspring/decimal"
	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder"
	"strconv"
)

type TxContract struct {
	vmVersion    []byte
	lenGasLimit  []byte
	gasLimit     []byte
	lenGasPrice  []byte
	gasPrice     []byte
	dataHex      []byte
	address      []byte
	opCall       []byte
}

var (
	//小数位长度
	coinDecimal decimal.Decimal = decimal.NewFromFloat(100000000)
)

func newTxContractForEmptyTrans(vcontract Vcontract) (*TxContract, error) {
	var ret TxContract

	vmVersion, err := hex.DecodeString("0104")
	if err != nil {
		return nil, err
	}

	//十进制转十六进制
	//gasLimit
	gasLimitInt, err := strconv.ParseInt(vcontract.GasLimit,10,64)
	if err != nil {
		return nil, err
	}
	gasLimitHex := strconv.FormatInt(gasLimitInt, 16)
	if len(gasLimitHex)%2 == 1 {
		gasLimitHex = "0" + gasLimitHex
	}
	gasLimit, err := reverseStringToBytes(gasLimitHex)
	if err != nil {
		return nil, err
	}

	//Length of gasLimit
	lenGasLimitHex := strconv.FormatInt(int64(len(gasLimit)),16)
	if len(lenGasLimitHex)%2 == 1 {
		lenGasLimitHex = "0" + lenGasLimitHex
	}
	lenGasLimit, err := hex.DecodeString(lenGasLimitHex)
	if err != nil {
		return nil, err
	}

	//gasPrice
	gasPriceHex, err := strconv.ParseInt(vcontract.GasPrice.String(),16,64)
	if err != nil {
		return nil, err
	}
	gasPrice, err := reverseStringToBytes(string(gasPriceHex))
	if err != nil {
		return nil, err
	}

	//length of gasPrice
	lenGasPriceHex, err := strconv.ParseInt(string(len(gasPrice)),16,64)
	if err != nil {
		return nil, err
	}
	lenGasPrice, err := hex.DecodeString(string(lenGasPriceHex))
	if err != nil {
		return nil, err
	}

	//AmountTo32ByteArg
	amountDecimal := vcontract.SendAmount.Mul(coinDecimal)
	sotashiAmount := amountDecimal.IntPart()
	hexAmount := strconv.FormatInt(sotashiAmount, 16)
	defaultLen := 64
	addLen := defaultLen - len(hexAmount)
	var bytesArg string
	for i := 0; i<addLen; i++ {
		bytesArg = bytesArg + "0"
	}
	bytesArg = bytesArg + hexAmount

	//addrTo32bytesArg
	addressToHash160, _ := addressEncoder.AddressDecode(vcontract.To, addressEncoder.QTUM_testnetAddressP2PKH)
	//fmt.Printf("addressToHash160: %s\n",hex.EncodeToString(addressToHash160))
	addrTo32bytesArg := append([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, addressToHash160[:]...)
	//fmt.Printf("to32bytesArg: %s\n",hex.EncodeToString(to32bytesArg))

	//dataHex
	combineString := hex.EncodeToString(append([]byte{0xa9, 0x05, 0x9c, 0xbb}, addrTo32bytesArg[:]...))
	dataHexString := combineString + bytesArg
	dataHex, err := hex.DecodeString(dataHexString)
	if err != nil {
		return nil, err
	}

	address, err := hex.DecodeString(string(len(vcontract.ContractAddr)))
	if err != nil {
		return nil, err
	}

	opCall := []byte{0xC2}

	ret = TxContract{vmVersion,lenGasLimit,gasLimit,lenGasPrice,gasPrice,dataHex,address,opCall}
	return &ret, nil
}