package qtum

import (
	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder"
	"fmt"
	"encoding/hex"
)

func AddressToArg(address string) ([]byte, error) {

	addressToHash160, _ := addressEncoder.AddressDecode(address, addressEncoder.QTUM_testnetAddressP2PKH)
	fmt.Printf("addressToHash160: %s\n",hex.EncodeToString(addressToHash160))

	to32bytesArg := append([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, addressToHash160[:]...)
	fmt.Printf("to32bytesArg: %s\n",hex.EncodeToString(to32bytesArg))

	return to32bytesArg, nil
}

func (wm *WalletManager)GetUnspentByAddress(contractAddress, address string) (error) {

	//to32bytesArg, err := AddressToArg(address)
	//if err != nil {
	//	return err
	//}
	//
	//combineString := append([]byte{0x70, 0xa0, 0x82, 0x31}, to32bytesArg[:]...)
	//fmt.Printf("combineString: %s\n",hex.EncodeToString(combineString))
	//
	//request := []interface{}{
	//	contractAddress,
	//	combineString,
	//}
	//
	//result, err := wm.walletClient.Call("callcontract", request)
	//if err != nil {
	//	return  err
	//}
	//


	return nil
}