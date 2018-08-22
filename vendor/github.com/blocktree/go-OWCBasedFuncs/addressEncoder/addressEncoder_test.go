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
	m_p2pkh_addr, err := AddressEncode(m_p2pkh_hash160, BTC_mainnetAddressP2PKH)

	if err != nil {
		t.Error("btc mainnet p2pkh address encode error!")
	} else {
		if m_p2pkh_addr != "19xD3nnvEiu7Uqd8irRvF3j5ExLb4ZtSju" {
			t.Error("btc mainnet p2pkh address encode wrong result!")
		} else {
			fmt.Println("encoded result:", m_p2pkh_addr)
		}
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
	m_p2sh_addr, err := AddressEncode(m_p2sh_hash160, BTC_mainnetAddressP2SH)

	if err != nil {
		t.Error("btc mainnet p2sh address encode error!")
	} else {
		if m_p2sh_addr != "3BYx8ciMdywxd2bbn5h9V7EAZtzLg2RhhX" {
			t.Error("btc mainnet p2sh address encode wrong result!")
		} else {
			fmt.Println("encoded result:", m_p2sh_addr)
		}

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
	m_pri_wif_comp, err := AddressEncode(m_privateKey, BTC_mainnetPrivateWIFCompressed)

	if err != nil {
		t.Error("btc mainnet prikey WIF-Compressed encode error!")
	} else {
		if m_pri_wif_comp != "L5P8PR3euZKUFsHJ3jRSzaLSXbBUXje8hR9fcsKzQSp9zoxZAqCS" {
			t.Error("btc mainnet prikey WIF-Compressed encode wrong result!")
		} else {
			fmt.Println("encoded result:", m_pri_wif_comp)
		}

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

func Test_bch_address(t *testing.T) {
	cashAddress := "bitcoincash:qpm2qsznhks23z7629mms6s4cwef74vcwvy22gdx6a"
	cashHash := []byte{0, 0x76, 0xa0, 0x40, 0x53, 0xbd, 0xa0, 0xa8, 0x8b, 0xda, 0x51, 0x77, 0xb8, 0x6a, 0x15, 0xc3, 0xb2, 0x9f, 0x55, 0x98, 0x73}

	fmt.Println("BCH CashAddress Encode test")

	cashAddressChk, err := AddressEncode(cashHash, BCH_mainnetAddressCash)
	if err != nil {
		t.Error("BCH cashaddress encode error")
	} else {
		if cashAddressChk != cashAddress {
			t.Error("BCH cashaddress encode result wrong")
		} else {
			fmt.Println("encode result:", cashAddressChk)
		}
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
