package btcLikeTxDriver

import (
	"encoding/hex"
	"github.com/blocktree/go-owcdrivers/addressEncoder"
	"strconv"
	"github.com/blocktree/OpenWallet/log"
)

type TxContract struct {
	vmVersion    []byte
	lenGasLimit  []byte
	gasLimit     []byte
	lenGasPrice  []byte
	gasPrice     []byte
	dataHex      []byte
	lenContract  []byte
	contractAddr []byte
	opCall       []byte
}

var (
	//小数位长度
	//coinDecimal decimal.Decimal = decimal.NewFromFloat(100000000)
)

func newTxContractForEmptyTrans(vcontract Vcontract, isTestNet bool) (*TxContract, error) {
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
	gasPriceInt, err := strconv.ParseInt(vcontract.GasPrice,10,64)
	if err != nil {
		return nil, err
	}
	gasPriceHex := strconv.FormatInt(gasPriceInt, 16)
	if len(gasPriceHex)%2 == 1 {
		gasPriceHex = "0" + gasPriceHex
	}
	gasPrice, err := reverseStringToBytes(gasPriceHex)
	if err != nil {
		return nil, err
	}

	//length of gasPrice
	lenGasPriceHex := strconv.FormatInt(int64(len(gasPrice)),16)
	if len(lenGasPriceHex)%2 == 1 {
		lenGasPriceHex = "0" + lenGasPriceHex
	}
	lenGasPrice, err := hex.DecodeString(lenGasPriceHex)
	if err != nil {
		return nil, err
	}

	//AmountTo32ByteArg
	amountDecimal := vcontract.SendAmount
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
	var addressToHash160 []byte
	if isTestNet {
		addressToHash160, _ = addressEncoder.AddressDecode(vcontract.To, addressEncoder.QTUM_testnetAddressP2PKH)
	}else {
		addressToHash160, _ = addressEncoder.AddressDecode(vcontract.To, addressEncoder.QTUM_mainnetAddressP2PKH)
	}

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

	if int64(len(vcontract.ContractAddr))%2 == 1 {
		log.Errorf("Contract address length error.")
	}
	lanAddressHex := strconv.FormatInt(int64(len(vcontract.ContractAddr))/2,16)
	lanAddress, err := hex.DecodeString(lanAddressHex)
	if err != nil {
		return nil, err
	}

	contractAddr, err := hex.DecodeString(vcontract.ContractAddr)
	if err != nil {
		return nil, err
	}

	opCall := []byte{0xC2}

	ret = TxContract{vmVersion,lenGasLimit,gasLimit,lenGasPrice,gasPrice,dataHex,lanAddress,contractAddr,opCall}
	return &ret, nil
}