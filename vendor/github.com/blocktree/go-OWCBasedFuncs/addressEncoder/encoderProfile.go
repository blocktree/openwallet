package addressEncoder

var (
	btcAlphabet       = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	zecAlphabet       = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	btcBech32Alphabet = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	ltcAlphabet       = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	ltcBech32Alphabet = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	bchLegacyAlphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	bchCashAlphabet   = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	xtzAlphabet       = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	hcAlphabet        = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	qtumAlphabet      = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	dcrdAlphabet      = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
)

type AddressType struct {
	encodeType   string //编码类型
	alphabet     string //码表
	checksumType string //checksum类型(prefix string when encode type is base32PolyMod)
	hashType     string //地址hash类型，传入数据为公钥时起效
	hashLen      int    //编码前的数据长度
	prefix       []byte //数据前面的填充
	suffix       []byte //数据后面的填充
}

var (
	//BTC stuff
	BTC_mainnetAddressP2PKH         = AddressType{"base58", btcAlphabet, "doubleSHA256", "h160", 20, []byte{0x00}, nil}
	BTC_mainnetAddressP2SH          = AddressType{"base58", btcAlphabet, "doubleSHA256", "h160", 20, []byte{0x05}, nil}
	BTC_mainnetAddressBech32V0      = AddressType{"bech32", btcBech32Alphabet, "bc", "h160", 20, nil, nil}
	BTC_mainnetPrivateWIF           = AddressType{"base58", btcAlphabet, "doubleSHA256", "", 32, []byte{0x80}, nil}
	BTC_mainnetPrivateWIFCompressed = AddressType{"base58", btcAlphabet, "doubleSHA256", "", 32, []byte{0x80}, []byte{0x01}}
	BTC_mainnetPublicBIP32          = AddressType{"base58", btcAlphabet, "doubleSHA256", "", 74, []byte{0x04, 0x88, 0xB2, 0x1E}, nil}
	BTC_mainnetPrivateBIP32         = AddressType{"base58", btcAlphabet, "doubleSHA256", "", 74, []byte{0x04, 0x88, 0xAD, 0xE4}, nil}
	BTC_testnetAddressP2PKH         = AddressType{"base58", btcAlphabet, "doubleSHA256", "h160", 20, []byte{0x6F}, nil}
	BTC_testnetAddressP2SH          = AddressType{"base58", btcAlphabet, "doubleSHA256", "h160", 20, []byte{0xC4}, nil}
	BTC_testnetAddressBech32V0      = AddressType{"bech32", btcBech32Alphabet, "tb", "h160", 20, nil, nil}
	BTC_testnetPrivateWIF           = AddressType{"base58", btcAlphabet, "doubleSHA256", "", 32, []byte{0xEF}, nil}
	BTC_testnetPrivateWIFCompressed = AddressType{"base58", btcAlphabet, "doubleSHA256", "", 32, []byte{0xEF}, []byte{0x01}}
	BTC_testnetPublicBIP32          = AddressType{"base58", btcAlphabet, "doubleSHA256", "", 74, []byte{0x04, 0x35, 0x87, 0xCF}, nil}
	BTC_testnetPrivateBIP32         = AddressType{"base58", btcAlphabet, "doubleSHA256", "", 74, []byte{0x04, 0x35, 0x83, 0x94}, nil}
	
	//ZEC stuff
	ZEC_mainnet_t_AddressP2PKH         = AddressType{"base58", zecAlphabet, "doubleSHA256", "h160", 20, []byte{0x1C,0xB8}, nil}
	ZEC_mainnet_t_AddressP2SH          = AddressType{"base58", zecAlphabet, "doubleSHA256", "h160", 20, []byte{0x1C,0xBD}, nil}
	ZEC_testnet_t_AddressP2PKH         = AddressType{"base58", zecAlphabet, "doubleSHA256", "h160", 20, []byte{0x1D,0x25}, nil}
	ZEC_testnet_t_AddressP2SH          = AddressType{"base58", zecAlphabet, "doubleSHA256", "h160", 20, []byte{0x1C,0xBA}, nil}

	//LTC stuff
	LTC_mainnetAddressP2PKH         = AddressType{"base58", ltcAlphabet, "doubleSHA256", "h160", 20, []byte{0x30}, nil}
	LTC_mainnetAddressP2SH          = AddressType{"base58", ltcAlphabet, "doubleSHA256", "h160", 20, []byte{0x05}, nil}
	LTC_mainnetAddressP2SH2         = AddressType{"base58", ltcAlphabet, "doubleSHA256", "h160", 20, []byte{0x32}, nil}
	LTC_mainnetAddressBech32V0      = AddressType{"bech32", ltcBech32Alphabet, "ltc", "h160", 20, nil, nil}
	LTC_mainnetPrivateWIF           = AddressType{"base58", ltcAlphabet, "doubleSHA256", "", 32, []byte{0xB0}, nil}
	LTC_mainnetPrivateWIFCompressed = AddressType{"base58", ltcAlphabet, "doubleSHA256", "", 32, []byte{0xB0}, []byte{0x01}}
	LTC_mainnetPublicBIP32          = AddressType{"base58", btcAlphabet, "doubleSHA256", "", 74, []byte{0x04, 0x88, 0xB2, 0x1E}, nil}
	LTC_mainnetPrivateBIP32         = AddressType{"base58", btcAlphabet, "doubleSHA256", "", 74, []byte{0x04, 0x88, 0xAD, 0xE4}, nil}
	LTC_testnetAddressP2PKH         = AddressType{"base58", ltcAlphabet, "doubleSHA256", "h160", 20, []byte{0x6F}, nil}
	LTC_testnetAddressP2SH          = AddressType{"base58", ltcAlphabet, "doubleSHA256", "h160", 20, []byte{0xC4}, nil}
	LTC_testnetAddressBech32V0      = AddressType{"bech32", ltcBech32Alphabet, "tltc", "h160", 20, nil, nil}
	LTC_testnetPrivateWIF           = AddressType{"base58", ltcAlphabet, "doubleSHA256", "", 32, []byte{0xEF}, nil}
	LTC_testnetPrivateWIFCompressed = AddressType{"base58", ltcAlphabet, "doubleSHA256", "", 32, []byte{0xEF}, []byte{0x01}}
	LTC_testnetPublicBIP32          = AddressType{"base58", btcAlphabet, "doubleSHA256", "", 74, []byte{0x04, 0x35, 0x87, 0xCF}, nil}
	LTC_testnetPrivateBIP32         = AddressType{"base58", btcAlphabet, "doubleSHA256", "", 74, []byte{0x04, 0x35, 0x83, 0x94}, nil}

	//BCH stuff
	BCH_mainnetAddressLegacy = AddressType{"base58", bchLegacyAlphabet, "doubleSHA256", "h160", 20, []byte{0x00}, nil}
	BCH_mainnetAddressCash   = AddressType{"base32PolyMod", bchCashAlphabet, "bitcoincash", "h160", 21, nil, nil}

	//XTZ stuff
	XTZ_mainnetAddress_tz1   = AddressType{"base58", xtzAlphabet, "doubleSHA256", "blake2b160", 20, []byte{0x06, 0xA1, 0x9F}, nil}
	XTZ_mainnetAddress_tz2   = AddressType{"base58", xtzAlphabet, "doubleSHA256", "blake2b160", 20, []byte{0x06, 0xA1, 0xA1}, nil}
	XTZ_mainnetAddress_tz3   = AddressType{"base58", xtzAlphabet, "doubleSHA256", "blake2b160", 20, []byte{0x06, 0xA1, 0xA4}, nil}
	XTZ_mainnetPublic_edpk   = AddressType{"base58", xtzAlphabet, "doubleSHA256", "", 32, []byte{0x0D, 0x0F, 0x25, 0xD9}, nil}
	XTZ_mainnetPrivate_edsk  = AddressType{"base58", xtzAlphabet, "doubleSHA256", "", 64, []byte{0x0D, 0x0F, 0x3A, 0x07}, nil}
	XTZ_mainnetPrivate_edsk2 = AddressType{"base58", xtzAlphabet, "doubleSHA256", "", 32, []byte{0x2B, 0xF6, 0x4E, 0x07}, nil}
	XTZ_mainnetPrivate_spsk  = AddressType{"base58", xtzAlphabet, "doubleSHA256", "", 32, []byte{0x11, 0xA2, 0xE0, 0xC9}, nil}
	XTZ_mainnetPrivate_p2sk  = AddressType{"base58", xtzAlphabet, "doubleSHA256", "blake2b160", 32, []byte{0x10, 0x51, 0xEE, 0xBD}, nil}

	//HC stuff
	HC_mainnetPublicAddress = AddressType{"base58", hcAlphabet, "doubleBlake256", "h160", 20, []byte{0x09, 0x7F}, nil}
	//ETH stuff
	ETH_mainnetPublicAddress = AddressType{"eip55", "", "keccak256", "", 20, nil, nil}

	//QTUM stuff
	QTUM_mainnetAddressP2PKH         = AddressType{"base58", qtumAlphabet, "doubleSHA256", "h160", 20, []byte{0x3A}, nil}
	QTUM_mainnetAddressP2SH          = AddressType{"base58", qtumAlphabet, "doubleSHA256", "h160", 20, []byte{0x32}, nil}
	QTUM_mainnetPrivateWIF           = AddressType{"base58", qtumAlphabet, "doubleSHA256", "", 32, []byte{0x80}, nil}
	QTUM_mainnetPrivateWIFCompressed = AddressType{"base58", qtumAlphabet, "doubleSHA256", "", 32, []byte{0x80}, []byte{0x01}}
	QTUM_mainnetPublicBIP32          = AddressType{"base58", qtumAlphabet, "doubleSHA256", "", 74, []byte{0x04, 0x88, 0xB2, 0x1E}, nil}
	QTUM_mainnetPrivateBIP32         = AddressType{"base58", qtumAlphabet, "doubleSHA256", "", 74, []byte{0x04, 0x88, 0xAD, 0xE4}, nil}
	QTUM_testnetAddressP2PKH         = AddressType{"base58", qtumAlphabet, "doubleSHA256", "h160", 20, []byte{0x78}, nil}
	QTUM_testnetAddressP2SH          = AddressType{"base58", qtumAlphabet, "doubleSHA256", "h160", 20, []byte{0x6E}, nil}
	QTUM_testnetPrivateWIF           = AddressType{"base58", qtumAlphabet, "doubleSHA256", "", 32, []byte{0xEF}, nil}
	QTUM_testnetPrivateWIFCompressed = AddressType{"base58", qtumAlphabet, "doubleSHA256", "", 32, []byte{0xEF}, []byte{0x01}}
	QTUM_testnetPublicBIP32          = AddressType{"base58", qtumAlphabet, "doubleSHA256", "", 74, []byte{0x04, 0x35, 0x87, 0xCF}, nil}
	QTUM_testnetPrivateBIP32         = AddressType{"base58", qtumAlphabet, "doubleSHA256", "", 74, []byte{0x04, 0x35, 0x83, 0x94}, nil}

	//DCRD
	DCRD_mainnetAddressP2PKH         =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x07, 0x3f}, nil} //PubKeyHashAddrID, stars with Ds
    DCRD_mainnetAddressP2PK          =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x13, 0x86}, nil} //PubKeyAddrID,stars with Dk
    DCRD_mainnetAddressPKHEdwards    =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x07, 0x1f}, nil}//PKHEdwardsAddrID,starts with De
	DCRD_mainnetAddressPKHSchnorr    =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x07, 0x01}, nil}//PKHSchnorrAddrID,starts with DS
    DCRD_mainnetAddressP2SH          =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x07, 0x1a}, nil} //ScriptHashAddrID,starts with Dc
	DCRD_mainnetAddressPrivate       =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x22, 0xde}, nil} // PrivateKeyID, starts with Pm
	
	DCRD_testnetAddressP2PKH         =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x0f, 0x21}, nil} //PubKeyHashAddrID,starts with Ts
    DCRD_testnetAddressP2PK          =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x28, 0xf7}, nil} //PubKeyAddrID, starts with Tk
    DCRD_testnetAddressPKHEdwards    =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x0f, 0x01}, nil}//PKHEdwardsAddrID,starts with Te
	DCRD_testnetAddressP2PKHSchnorr  =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x0e, 0xe3}, nil}//PKHSchnorrAddrID,starts with TS
    DCRD_testnetAddressP2SH          =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x0e, 0xfc}, nil} //ScriptHashAddrID,starts with Tc
    DCRD_testnetAddressPrivate       =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x23, 0x0e}, nil} //PrivateKeyID,starts with Pt

	DCRD_simnetAddressP2PKH          =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x0e, 0x91}, nil} //PubKeyHashAddrID,starts with Ss
	DCRD_simnetAddressP2PK           =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x27, 0x6f}, nil} //PubKeyAddrID,starts with Sk
    DCRD_simnetAddressPKHEdwards     =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x0e, 0x71}, nil}//PKHEdwardsAddrID,starts with Se
	DCRD_simnetAddressPKHSchnorr     =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x0e, 0x53}, nil}//PKHSchnorrAddrID,starts with SS
	DCRD_simnetAddressP2SH           =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x0e, 0x6c}, nil}//ScriptHashAddrID,starts with Sc
	DCRD_simnetAddressPrivate        =  AddressType{"base58", dcrdAlphabet, "doubleBlake256", "ripemd160", 20, []byte{0x23, 0x07}, nil}//PrivateKeyID, starts with Ps
)
