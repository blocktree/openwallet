package btcLikeTxDriver

import (
	"errors"
	"strings"
)

var (
	ErrorInvalidAddress = errors.New("Invalid address!")

	BTCBech32Alphabet = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

	CHARSET_REV = []int8{
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		15, -1, 10, 17, 21, 20, 26, 30, 7, 5, -1, -1, -1, -1, -1, -1,
		-1, 29, -1, 24, 13, 25, 9, 8, 23, -1, 18, 22, 31, 27, 19, -1,
		1, 0, 3, 16, 11, 28, 12, 14, 6, 4, 2, -1, -1, -1, -1, -1,
		-1, 29, -1, 24, 13, 25, 9, 8, 23, -1, 18, 22, 31, 27, 19, -1,
		1, 0, 3, 16, 11, 28, 12, 14, 6, 4, 2, -1, -1, -1, -1, -1}
)

func catBytes(data1 []int8, data2 []int8) []int8 {
	return append(data1, data2...)
}

func expandPrefix(prefix string) []int8 {
	ret := make([]int8, len(prefix)*2+1)
	for i := 0; i < len(prefix); i++ {
		c := prefix[i]
		ret[i] = int8(c >> 5)
		ret[i+len(prefix)+1] = int8(c & 0x1f)
	}
	ret[len(prefix)] = 0
	return ret
}

func polyMod(v []int8) uint32 {
	c := uint32(1)
	for _, v_i := range v {
		c0 := uint8(c >> 25)

		c = ((c & 0x1ffffff) << 5) ^ uint32(v_i)

		if c0&1 != 0 {
			c ^= 0x3b6a57b2
		}
		if c0&2 != 0 {
			c ^= 0x26508e6d
		}
		if c0&4 != 0 {
			c ^= 0x1ea119fa
		}
		if c0&8 != 0 {
			c ^= 0x3d4233dd
		}
		if c0&16 != 0 {
			c ^= 0x2a1462b3
		}
	}
	return c ^ 1
}

func lowerCase(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return (c - 'A') + 'a'
	}
	return c
}

func verifyChecksum(prefix string, data []int8) bool {
	return polyMod(catBytes(expandPrefix(prefix), data)) == 0
}

func calcChecksum(prefix string, data []int8) []int8 {
	enc := catBytes(expandPrefix(prefix), data)
	ret := [6]int8{}
	tmp := make([]int8, len(enc)+6)

	copy(tmp, enc)

	mod := polyMod(tmp)

	for i := 0; i < 6; i++ {
		ret[i] = int8((mod >> (5 * (5 - uint(i)))) & 0x1f)
	}

	return ret[:]
}

func byteShl1(in *[]int8) {
	tmp := make([]int8, len(*in))
	copy(tmp, *in)
	for i := 0; i < len(tmp)-1; i++ {
		tmp1 := tmp[i] << 1
		tmp2 := tmp[i+1] >> 7
		tmp2 &= 1
		tmp1 |= tmp2
		tmp[i] = tmp1
	}
	tmp[len(tmp)-1] <<= 1
	copy(*in, tmp)
}
func byteShl5(in *[]int8) {
	for i := 0; i < 5; i++ {
		byteShl1(in)
	}
}

func extendPayload(payload []int8) []int8 {
	length := (len(payload)*8 + 4) / 5
	ret := make([]int8, length)
	i := 0
	j := 0
	for i = len(payload) * 8; i >= 5; i -= 5 {
		ret[j] = (payload[0] >> 3) & int8(0x1F)
		byteShl5(&payload)
		j++
	}
	if i > 0 {
		ret[j] = (payload[0] >> 3) & 0x1f
	}
	return ret
}

func unecxtendPayload(extendedPayload []int8) []int8 {
	length := len(extendedPayload) * 5 / 8

	ret := make([]int8, length)

	for i := 0; i < len(extendedPayload)-1; i++ {
		ret[length-1] |= extendedPayload[i]
		byteShl5(&ret)
	}
	ret[length-1] |= extendedPayload[len(extendedPayload)-1]

	return ret
}

func Bech32Encode(prefix, alphabet string, payload []byte) string {
	int8Payload := make([]int8, len(payload))
	for i := 0; i < len(payload); i++ {
		int8Payload[i] = int8(payload[i])
	}
	extendPayload := extendPayload(int8Payload)
	extendPayload = append([]int8{0}, extendPayload...)
	checksum := calcChecksum(prefix, extendPayload)
	combined := catBytes(extendPayload, checksum)

	ret := prefix + "1"
	for _, b := range combined {
		ret += alphabet[b : b+1]
	}
	return ret
}

func Bech32Decode(address string) ([]byte, error) {
	lower := false
	upper := false
	hasNumber := false
	prefixSize := 0
	for i := 0; i < len(address); i++ {
		c := address[i]
		if c >= 'a' && c <= 'z' {
			lower = true
			continue
		}
		if c >= 'A' && c <= 'Z' {
			upper = true
			continue
		}
		if c == '0' || (c >= '2' && c <= '9') {
			hasNumber = true
			continue
		}
		if c == '1' {
			if hasNumber || i == 0 || prefixSize != 0 {
				return nil, ErrorInvalidAddress
			}
			prefixSize = i
			continue
		}
		return nil, ErrorInvalidAddress
	}

	if upper && lower {
		return nil, ErrorInvalidAddress
	}

	prefixStr := strings.Split(address, "1")[0]
	prefixSize++
	valueSize := len(address) - prefixSize
	value := make([]int8, valueSize)
	for i := 0; i < valueSize; i++ {
		c := address[i+prefixSize]
		if c > 127 || CHARSET_REV[c] == -1 {
			return nil, ErrorInvalidAddress
		}
		value[i] = CHARSET_REV[c]
	}

	if !verifyChecksum(prefixStr, value) {
		return nil, ErrorInvalidAddress
	}

	tmp := make([]int8, len(value)-6)
	copy(tmp, value)

	ret := unecxtendPayload(tmp)
	bytePayload := make([]byte, len(ret))

	for i := 0; i < len(ret); i++ {
		bytePayload[i] = byte(ret[i])
	}
	return bytePayload, nil
}
