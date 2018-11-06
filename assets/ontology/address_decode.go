package ontology

import (
	"github.com/blocktree/go-owcdrivers/addressEncoder"
	"fmt"
)

type addressDecoder struct {
	wm *WalletManager //钱包管理者
}

//NewAddressDecoder 地址解析器
func NewAddressDecoder(wm *WalletManager) *addressDecoder {
	decoder := addressDecoder{}
	decoder.wm = wm
	return &decoder
}

//ScriptPubKeyToBech32Address scriptPubKey转Bech32地址
func ScriptPubKeyToBech32Address(scriptPubKey []byte, isTestnet bool) (string, error) {
	var (
		hash []byte
	)

	cfg := addressEncoder.BTC_mainnetAddressBech32V0
	if isTestnet {
		cfg = addressEncoder.BTC_testnetAddressBech32V0
	}

	if len(scriptPubKey) == 22 || len(scriptPubKey) == 34 {

		hash = scriptPubKey[2:]

		address := addressEncoder.AddressEncode(hash, cfg)

		return address, nil

	} else {
		return "", fmt.Errorf("scriptPubKey length is invalid")
	}

}
