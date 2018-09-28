package btcLikeTxDriver

import (
	"errors"
	"strings"
)

type TxOut struct {
	amount     []byte
	lockScript []byte
}

func newTxOutForEmptyTrans(vout []Vout) ([]TxOut, error) {
	var ret []TxOut

	for _, v := range vout {
		amount := uint64ToLittleEndianBytes(v.Amount)

		if strings.Index(v.Address, Bech32Prefix) == 0 {
			redeem, err := Bech32Decode(v.Address)
			if err != nil {
				return nil, errors.New("Invalid bech32 type address!")
			}

			redeem = append([]byte{byte(len(redeem))}, redeem...)
			redeem = append([]byte{0x00}, redeem...)

			ret = append(ret, TxOut{amount, redeem})
		}

		prefix, hash, err := DecodeCheck(v.Address)
		if err != nil {
			return nil, errors.New("Invalid address to send!")
		}

		if len(hash) != 0x14 {
			return nil, errors.New("Invalid address to send!")
		}

		hash = append([]byte{byte(len(hash))}, hash...)
		hash = append([]byte{OpCodeHash160}, hash...)
		if prefix == P2PKHPrefix {
			hash = append(hash, OpCodeEqualVerify, OpCodeCheckSig)
			hash = append([]byte{OpCodeDup}, hash...)
		} else if prefix == P2SHPrefix {
			hash = append(hash, OpCodeEqual)
		} else {
			return nil, errors.New("Invalid address to send!")
		}

		ret = append(ret, TxOut{amount, hash})
	}
	return ret, nil
}
