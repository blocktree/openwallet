package addressEncoder

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func Test_btc_address(t *testing.T) {
	m_p2pkh_hash160 := []byte{0x62, 0x31, 0xf1, 0x00, 0x5e, 0x86, 0xc0, 0x3d, 0x5f, 0xbd, 0x41, 0x77, 0x69, 0x85, 0xd0, 0x94, 0xcc, 0xb6, 0x82, 0xd3}
	m_p2sh_hash160 := []byte{0x6c, 0x2a, 0xc3, 0xce, 0x63, 0x85, 0x1b, 0x90, 0x50, 0x4f, 0x75, 0xc5, 0xf3, 0x97, 0x87, 0x48, 0x22, 0xb5, 0x29, 0xc6}
	m_privateKey := []byte{0xf3, 0x8b, 0x35, 0x53, 0x51, 0x50, 0xef, 0x6e, 0x24, 0x46, 0xe4, 0xaa, 0x4d, 0x1e, 0x55, 0xaf, 0x31, 0xeb, 0xcc, 0x84, 0xef, 0x96, 0x04, 0xd4, 0x21, 0x12, 0xe5, 0xe8, 0x3e, 0xe9, 0xbc, 0x3a}

	fmt.Println("mainnet p2pkh address encode:")
	m_p2pkh_addr := AddressEncode(m_p2pkh_hash160, BTC_mainnetAddressP2PKH)

	if m_p2pkh_addr != "19xD3nnvEiu7Uqd8irRvF3j5ExLb4ZtSju" {
		t.Error("btc mainnet p2pkh address encode wrong result!")
	} else {
		fmt.Println("encoded result:", m_p2pkh_addr)
	}

	fmt.Println("mainnet p2pkh address decode")
	m_p2pkh_check, err := AddressDecode(m_p2pkh_addr, BTC_mainnetAddressP2PKH)
	if err != nil {
		t.Error("btc mainnet p2pkh address decode error!")
	} else {
		for i := 0; i < BTC_mainnetAddressP2PKH.hashLen; i++ {
			if m_p2pkh_check[i] != m_p2pkh_hash160[i] {
				t.Error("btc mainnet p2pkh address decode wrong result!")
				break
			}
		}
		fmt.Println("decode result:", hex.EncodeToString(m_p2pkh_check[:]))
	}

	fmt.Println("mainnet p2sh address encode:")
	m_p2sh_addr := AddressEncode(m_p2sh_hash160, BTC_mainnetAddressP2SH)

	if m_p2sh_addr != "3BYx8ciMdywxd2bbn5h9V7EAZtzLg2RhhX" {
		t.Error("btc mainnet p2sh address encode wrong result!")
	} else {
		fmt.Println("encoded result:", m_p2sh_addr)
	}

	fmt.Println("mainnet p2sh address decode")
	m_p2sh_check, err := AddressDecode(m_p2sh_addr, BTC_mainnetAddressP2SH)
	if err != nil {
		t.Error("btc mainnet p2sh address decode error!")
	} else {
		for i := 0; i < BTC_mainnetAddressP2SH.hashLen; i++ {
			if m_p2sh_check[i] != m_p2sh_hash160[i] {
				t.Error("btc mainnet p2sh address decode wrong result!")
				break
			}
		}
		fmt.Println("decode result:", hex.EncodeToString(m_p2sh_check[:]))
	}

	fmt.Println("mainnet prikey WIF-Compressed encode:")
	m_pri_wif_comp := AddressEncode(m_privateKey, BTC_mainnetPrivateWIFCompressed)

	if m_pri_wif_comp != "L5P8PR3euZKUFsHJ3jRSzaLSXbBUXje8hR9fcsKzQSp9zoxZAqCS" {
		t.Error("btc mainnet prikey WIF-Compressed encode wrong result!")
	} else {
		fmt.Println("encoded result:", m_pri_wif_comp)
	}

	fmt.Println("mainnet prikey WIF-Compressed decode:")
	m_pri_wif_comp_check, err := AddressDecode(m_pri_wif_comp, BTC_mainnetPrivateWIFCompressed)
	if err != nil {
		t.Error("btc mainnet prikey WIF-Compressed decode error!")
	} else {
		for i := 0; i < BTC_mainnetPrivateWIFCompressed.hashLen; i++ {
			if m_pri_wif_comp_check[i] != m_privateKey[i] {
				t.Error("btc mainnet prikey WIF-Compressed decode wrong result!")
				break
			}
		}
		fmt.Println("decode result:", hex.EncodeToString(m_pri_wif_comp_check[:]))
	}

}

func Test_ltc_address(t *testing.T) {
	m_p2pkh_hash160 := []byte{0x62, 0x31, 0xf1, 0x00, 0x5e, 0x86, 0xc0, 0x3d, 0x5f, 0xbd, 0x41, 0x77, 0x69, 0x85, 0xd0, 0x94, 0xcc, 0xb6, 0x82, 0xd3}
	address := AddressEncode(m_p2pkh_hash160, LTC_mainnetAddressP2PKH)
	fmt.Println(address)
	chk, err := AddressDecode(address, LTC_mainnetAddressP2PKH)
	if err != nil {
		t.Error("decode error")
	} else {
		fmt.Println(hex.EncodeToString(chk))
	}
}
func Test_btc_address_fromkey(t *testing.T) {
	pubkey := []byte{0x04, 0x7D, 0xB2, 0x27, 0xD7, 0x09, 0x4C, 0xE2, 0x15, 0xC3, 0xA0, 0xF5, 0x7E, 0x1B, 0xCC, 0x73, 0x25, 0x51, 0xFE, 0x35, 0x1F, 0x94, 0x24, 0x94, 0x71, 0x93, 0x45, 0x67, 0xE0, 0xF5, 0xDC, 0x1B, 0xF7, 0x95, 0x96, 0x2B, 0x8C, 0xCC, 0xB8, 0x7A, 0x2E, 0xB5, 0x6B, 0x29, 0xFB, 0xE3, 0x7D, 0x61, 0x4E, 0x2F, 0x4C, 0x3C, 0x45, 0xB7, 0x89, 0xAE, 0x4F, 0x1F, 0x51, 0xF4, 0xCB, 0x21, 0x97, 0x2F, 0xFD}
	hash := []byte{0x1e, 0x35, 0x3b, 0x2e, 0x11, 0x25, 0xa1, 0x4e, 0xdd, 0xc2, 0x2d, 0xce, 0x7d, 0x46, 0xd7, 0xc4, 0xb5, 0xea, 0xc9, 0x91}
	ret := AddressEncode(pubkey, BTC_mainnetAddressP2PKH)
	fmt.Println(ret)

	chk, err := AddressDecode(ret, BTC_mainnetAddressP2PKH)
	if err != nil {
		t.Error("decode error")
	} else {
		for i := 0; i < 20; i++ {
			if chk[i] != hash[i] {
				t.Error("decode wrong result")
			}
		}
		fmt.Println(hex.EncodeToString(chk))
	}
}
func Test_btc_bech32_address(t *testing.T) {

	address := "bc1qvgclzqz7smqr6haag9mknpwsjnxtdqkncr64kd"
	ret, err := AddressDecode(address, BTC_mainnetAddressBech32V0)
	if err != nil {
		t.Error("decode error")
	} else {
		fmt.Println(hex.EncodeToString(ret))
	}

	addresschk := AddressEncode(ret, BTC_mainnetAddressBech32V0)
	if addresschk != address {
		t.Error("encode error")
	} else {
		fmt.Println(addresschk)
	}
}
func Test_bch_address(t *testing.T) {
	cashAddress := "bitcoincash:qpm2qsznhks23z7629mms6s4cwef74vcwvy22gdx6a"
	cashHash := []byte{0, 0x76, 0xa0, 0x40, 0x53, 0xbd, 0xa0, 0xa8, 0x8b, 0xda, 0x51, 0x77, 0xb8, 0x6a, 0x15, 0xc3, 0xb2, 0x9f, 0x55, 0x98, 0x73}

	fmt.Println("BCH CashAddress Encode test")

	cashAddressChk := AddressEncode(cashHash, BCH_mainnetAddressCash)

	if cashAddressChk != cashAddress {
		t.Error("BCH cashaddress encode result wrong")
	} else {
		fmt.Println("encode result:", cashAddressChk)
	}

	fmt.Println("BCH CashAddress Decode test")
	cashHashChk, err := AddressDecode(cashAddress, BCH_mainnetAddressCash)
	if err != nil {
		t.Error("BCH cashaddress decode error")
	} else {
		for i := 0; i < len(cashHashChk); i++ {
			if cashHashChk[i] != cashHash[i] {
				t.Error("BCH cashaddress decode result wrong")
				break
			}
		}
		fmt.Println("decode result:", hex.EncodeToString(cashHashChk[:]))
	}

}

func Test_eth_address(t *testing.T) {

	keccak256_hash := []byte{0xdb, 0xf0, 0x3b, 0x40, 0x7c, 0x01, 0xe7, 0xcd, 0x3c, 0xbe, 0xa9, 0x95, 0x09, 0xd9, 0x3f, 0x8d, 0xdd, 0xc8, 0xc6, 0xfb}
	//decode_addr:=make([]byte,20)
	//var encode_addr string
	//str=eip55.Eip55_encode(addr[:])
	fmt.Println("ETH CashAddress Encode test")
	eth_encode_addr := AddressEncode(keccak256_hash, ETH_mainnetPublicAddress)

	if eth_encode_addr != "dbF03B407c01E7cD3CBea99509d93f8DDDC8C6FB" {
		t.Error("ETH cashaddress encode wrong result")
	} else {
		fmt.Println("ETH encode result:", string(eth_encode_addr))
	}

	fmt.Println("ETH Cashaddress Decode test")
	eth_decode_addr, err := AddressDecode(eth_encode_addr, ETH_mainnetPublicAddress)
	if err != nil {
		t.Error("ETH cashaddress decode error")
	} else {
		fmt.Printf("ETH decode result:")
		for _, t := range eth_decode_addr {
			fmt.Printf("0x%x ", t)
		}
		fmt.Printf("\n")
	}
}

func Test_DCRD_address(t *testing.T){
	//PubKeyHashAddrID test......
	mainnet_P2PKH_encodeAddress := "DsUZxxoHJSty8DCfwfartwTYbuhmVct7tJu"
	mainnet_P2PKH_decodeAddress :=[]byte{0x27, 0x89, 0xd5, 0x8c, 0xfa, 0x09, 0x57, 0xd2, 0x06, 0xf0,0x25, 0xc2, 0xaf, 0x05, 0x6f, 0xc8, 0xa7, 0x7c, 0xeb, 0xb0}
	fmt.Println("DCRD mainnet P2PKH  address decode test")
	mainnet_P2PKH_decodeAddressChk,err := AddressDecode(mainnet_P2PKH_encodeAddress, DCRD_mainnetAddressP2PKH)
	if err!=nil{
		t.Error("DCRD mainnet P2PKH address decode error")
	}else{
		fmt.Printf("DCRD   mainnet P2PKH address decode result:")
		for _, s := range mainnet_P2PKH_decodeAddressChk{
		 fmt.Printf("0x%x ", s)
		}
		fmt.Printf("\n")
	}

	fmt.Println("DCRD mainnet P2PKH  address encode test")
	mainnet_P2PKH_encodeAddressChk := AddressEncode(mainnet_P2PKH_decodeAddress,DCRD_mainnetAddressP2PKH)
	if mainnet_P2PKH_encodeAddressChk != mainnet_P2PKH_encodeAddress{
		t.Error("DCRD mainnet P2PKH address encode error")
	}else{
		fmt.Println("DCRD mainnet P2PKH address encode result:",string(mainnet_P2PKH_encodeAddressChk))
	}
	
	//PubKeyHashAddrID test......
	testnet_P2PKH_encodeAddress := "Tso2MVTUeVrjHTBFedFhiyM7yVTbieqp91h"
	testnet_P2PKH_decodeAddress :=[]byte{0xf1, 0x5d, 0xa1, 0xcb, 0x8d, 0x1b, 0xcb, 0x16, 0x2c, 0x6a,0xb4, 0x46, 0xc9, 0x57, 0x57, 0xa6, 0xe7, 0x91, 0xc9, 0x16}
	fmt.Println("DCRD testnet P2PKH  address decode test")
	testnet_P2PKH_decodeAddressChk,err := AddressDecode(testnet_P2PKH_encodeAddress, DCRD_testnetAddressP2PKH)
	if err !=nil{
		t.Error("DCRD testnet P2PKH address decode error")
	}else{
		fmt.Printf("DCRD   testnet P2PKH address decode result:")
		for _, s := range testnet_P2PKH_decodeAddressChk{
		 fmt.Printf("0x%x ", s)
		}
		fmt.Printf("\n")
	}
    fmt.Println("DCRD testnet P2PKH  address encode test")
	testnet_P2PKH_encodeAddressChk := AddressEncode(testnet_P2PKH_decodeAddress,DCRD_testnetAddressP2PKH)
	if testnet_P2PKH_encodeAddressChk != testnet_P2PKH_encodeAddress{
		t.Error("DCRD testnet P2PKH address encode error")
	}else{
		fmt.Println("DCRD mainnet P2PKH address encode result:",string(testnet_P2PKH_encodeAddressChk))
	}
}

func Test_DAS_address(t *testing.T){
	
	//NAS Account address test......
	Account_encodeAddress :="n1TV3sU6jyzR4rJ1D7jCAmtVGSntJagXZHC"
	fmt.Println("DAS account  address decode test")
	Account_decodeAddressChk,err :=AddressDecode(Account_encodeAddress,NAS_AccountAddress)
	if err!=nil{
		t.Error("NAS account address decode error")
	}else{
		fmt.Println("NAS account address decode result:")
		for _,s:=range Account_decodeAddressChk{
			fmt.Printf("0x%x ", s)
		}
		fmt.Printf("\n")
	}
	fmt.Println("DAS Account  address encode test")
	Account_encodeAddressChk := AddressEncode(Account_decodeAddressChk,NAS_AccountAddress)
	if Account_encodeAddressChk !="n1TV3sU6jyzR4rJ1D7jCAmtVGSntJagXZHC"{
		t.Error("NAS Account address encode error")
	}else{
		fmt.Println("NAS address encode result:",string(Account_encodeAddressChk))
	}

	//NAS smart contract address test......
	SmartContract_encodeAddress :="n1sLnoc7j57YfzAVP8tJ3yK5a2i56QrTDdK"
	fmt.Println("DAS smart contract  address decode test")
	SmartContract_decodeAddressChk,err :=AddressDecode(SmartContract_encodeAddress, NAS_SmartContractAddress)
	if err!=nil{
		t.Error("NAS smart contract address decode error")
	}else{
		fmt.Println("NAS smart contract address decode result:")
		for _,s:=range SmartContract_decodeAddressChk{
			fmt.Printf("0x%x ", s)
		}
		fmt.Printf("\n")
	}
	fmt.Println("DAS Account  address encode test")
	SmartContract_encodeAddressChk := AddressEncode(SmartContract_decodeAddressChk,NAS_SmartContractAddress)
	if SmartContract_encodeAddressChk !="n1sLnoc7j57YfzAVP8tJ3yK5a2i56QrTDdK"{
		t.Error("NAS smart contract encode error")
	}else{
		fmt.Println("NAS smart contract encode result:",string(SmartContract_encodeAddressChk))
	}


}