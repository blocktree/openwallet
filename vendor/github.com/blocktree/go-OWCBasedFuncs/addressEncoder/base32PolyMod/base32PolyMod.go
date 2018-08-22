package base32PolyMod

import (
	"errors"
	"strings"
)

var (
	ErrorInvalidAddress = errors.New("Invalid address!")

	charRev = []int8{
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		15, -1, 10, 17, 21, 20, 26, 30, 7, 5, -1, -1, -1, -1, -1, -1,
		-1, 29, -1, 24, 13, 25, 9, 8, 23, -1, 18, 22, 31, 27, 19, -1,
		1, 0, 3, 16, 11, 28, 12, 14, 6, 4, 2, -1, -1, -1, -1, -1,
		-1, 29, -1, 24, 13, 25, 9, 8, 23, -1, 18, 22, 31, 27, 19, -1,
		1, 0, 3, 16, 11, 28, 12, 14, 6, 4, 2, -1, -1, -1, -1, -1}
)
var (
	bitcoincashExpandPrefix = []int8{2, 9, 20, 3, 15, 9, 14, 3, 1, 19, 8, 0}
)

func catBytes(data1 []int8, data2 []int8) []int8 {
	return append(data1, data2...)
}

func expandPrefix(prefix string) []int8 {
	if prefix == "bitcoincash" {
		return bitcoincashExpandPrefix
	}
	return nil
}

func polyMod(V []int8) int64 {
	c := int64(1)
	for _, d := range V {
		c0 := int8(c >> 35)
		c = ((c & int64(0x07ffffffff)) << 5) ^ int64(d)
		if (c0 & 0x01) != 0 {
			c ^= int64(0x98f2bc8e61)
		}
		if (c0 & 0x02) != 0 {
			c ^= int64(0x79b76d99e2)
		}
		if (c0 & 0x04) != 0 {
			c ^= int64(0xf33e5fb3c4)
		}
		if (c0 & 0x08) != 0 {
			c ^= int64(0xae2eabe2a8)
		}
		if (c0 & 0x10) != 0 {
			c ^= int64(0x1e4f43e470)
		}
	}
	return c ^ 1
}

func calcChecksum(expandedPrefix, payload []int8) []int8 {
	ret := [8]int8{}
	tmp := make([]int8, len(expandedPrefix)+len(payload)+8)

	copy(tmp, expandedPrefix)
	copy(tmp[len(expandedPrefix):], payload)

	mod := polyMod(tmp)

	for i := 0; i < 8; i++ {
		ret[i] = int8((mod >> (5 * (7 - uint(i)))) & 0x1f)
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
	type1 := payload[0]
	versionByte := type1 << 3
	encodeedSize := int8(0)

	switch (len(payload) - 1) * 8 {
	case 160:
		encodeedSize = 0
		break
	default:
		break
	}

	versionByte |= encodeedSize
	payload[0] = versionByte

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

func verifyChecksum(prefix, payload []int8) bool {
	return polyMod(catBytes(prefix, payload)) == 0
}

func unecxtendPayload(extendedPayload []int8) []int8 {
	extraBits := 5 - len(extendedPayload)*5%8
	length := len(extendedPayload) * 5 / 8

	ret := make([]int8, length)

	for i := 0; i < len(extendedPayload)-2; i++ {
		ret[length-1] |= extendedPayload[i]
		byteShl5(&ret)
	}
	ret[length-1] |= extendedPayload[len(extendedPayload)-2]
	for i := 0; i < extraBits; i++ {
		byteShl1(&ret)
	}
	ret[length-1] |= extendedPayload[len(extendedPayload)-1] >> uint8(5-extraBits)

	return ret
}

func Encode(prefix, alphabet string, payload []byte) string {
	int8Payload := make([]int8, len(payload))
	for i := 0; i < len(payload); i++ {
		int8Payload[i] = int8(payload[i])
	}
	extendPayload := extendPayload(int8Payload)
	checksum := make([]int8, 8)
	if prefix == "bitcoincash" {
		checksum = calcChecksum(bitcoincashExpandPrefix, extendPayload)
	}
	combined := catBytes(extendPayload, checksum)
	ret := prefix
	ret += ":"
	for _, b := range combined {
		ret += alphabet[b : b+1]
	}
	return ret
}

func Decode(address, alphabet string) ([]byte, error) {
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
		if c >= '0' && c <= '9' {
			hasNumber = true
			continue
		}
		if c == ':' {
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

	prefixStr := strings.Split(address, ":")[0]
	prefixSize++
	valueSize := len(address) - prefixSize
	value := make([]int8, valueSize)
	for i := 0; i < valueSize; i++ {
		c := address[i+prefixSize]
		if c > 127 || charRev[c] == -1 {
			return nil, ErrorInvalidAddress
		}
		value[i] = charRev[c]
	}

	if prefixStr == "bitcoincash" {
		if !verifyChecksum(bitcoincashExpandPrefix, value) {
			return nil, ErrorInvalidAddress
		}
	}

	tmp := make([]int8, len(value)-8)
	copy(tmp, value)

	ret := unecxtendPayload(tmp)

	bytePayload := make([]byte, len(ret))

	for i := 0; i < len(ret); i++ {
		bytePayload[i] = byte(ret[i])
	}
	return bytePayload, nil
}
