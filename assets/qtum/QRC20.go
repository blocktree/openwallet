package qtum

import (
	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder"
	"fmt"
	"encoding/hex"
	"strconv"
)

func AddressTo32bytesArg(address string) ([]byte, error) {

	addressToHash160, _ := addressEncoder.AddressDecode(address, addressEncoder.QTUM_testnetAddressP2PKH)
	fmt.Printf("addressToHash160: %s\n",hex.EncodeToString(addressToHash160))

	to32bytesArg := append([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, addressToHash160[:]...)
	fmt.Printf("to32bytesArg: %s\n",hex.EncodeToString(to32bytesArg))

	return to32bytesArg, nil
}

func (wm *WalletManager)GetUnspentByAddress(contractAddress, address string) (*QRC20Unspent,error) {

	to32bytesArg, err := AddressTo32bytesArg(address)
	if err != nil {
		return nil, err
	}

	combineString := hex.EncodeToString(append([]byte{0x70, 0xa0, 0x82, 0x31}, to32bytesArg[:]...))
	fmt.Printf("combineString: %s\n",combineString)

	request := []interface{}{
		contractAddress,
		combineString,
	}

	result, err := wm.walletClient.Call("callcontract", request)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Callcontract result: %s", result.String())

	QRC20Utox := NewQRC20Unspent(result)

	return QRC20Utox, nil
}

func AmountTo32bytesArg(amount int64) (string, error) {

	hexAmount := strconv.FormatInt(amount, 16)

	defaultLen := 64
	addLen := defaultLen - len(hexAmount)
	var bytesArg string

	for i := 0; i<addLen; i++ {
		bytesArg = bytesArg + "0"
	}

	bytesArg = bytesArg + hexAmount

	return bytesArg, nil
}

func (wm *WalletManager)QRC20Transfer(contractAddress string, from string, to string, gasPrice string, amount int64, gasLimit int64) (string, error){
	amountToArg, err := AmountTo32bytesArg(amount)
	if err != nil {
		return "", err
	}

	addressToArg, err := AddressTo32bytesArg(to)
	if err != nil {
		return "", err
	}

	combineString := hex.EncodeToString(append([]byte{0xa9, 0x05, 0x9c, 0xbb}, addressToArg[:]...))

	finalString := combineString + amountToArg
	fmt.Printf("finalString: %s\n",finalString)

	request := []interface{}{
		contractAddress,
		finalString,
		0,
		gasLimit,
		gasPrice,
		from,
	}

	result, err := wm.walletClient.Call("sendtocontract", request)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}