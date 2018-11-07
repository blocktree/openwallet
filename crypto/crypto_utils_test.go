package crypto

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/mr-tron/base58/base58"
	"testing"
)

func TestHmacSHA1(t *testing.T) {
	var (
		appkey     = "2d68067484a20f1a346b3cf28a898ed7f5736f5bacf0fe60449da95efdb97ad4"
		appsecret  = "0dd1e322907ad7f55deaa35fec2aac97cae7931454d734364bc63f3e9b9f993a"
		timestamp  = "1506565393"
		period     = "3600"
		ciphertext []byte
	)

	ciphertext = HmacSHA1(appsecret, []byte(appkey+timestamp+period))

	fmt.Println("ciphertext = ", string(ciphertext))
}

func TestAESDecrypt(t *testing.T) {
	entext := "HQTqsjyaQUVjCT2gUoUvxSNAwUXic3qdNUPG18gZHYJYehtAULGnCjY2eYyt"
	keytext := "CxyAogk73eLTtQb8hmXvCm7ehb1tyfhv3NC6iD2nV2B9"
	bit, _ := base58.Decode(entext)
	key, _ := base58.Decode(keytext)
	fmt.Printf("key hex(%d): %s\n", len(key), hex.EncodeToString(key))
	fmt.Printf("bit hex(%d): %s\n", len(bit), hex.EncodeToString(bit))
	plaintext, _ := AESDecrypt(bit, key)

	t.Logf("plaintext: %s", string(plaintext))
}

func TestEncryptJSON(t *testing.T) {

	keytext := "CxyAogk73eLTtQb8hmXvCm7ehb1tyfhv3NC6iD2nV2B9"
	key, _ := base58.Decode(keytext)

	params := map[string]interface{}{
		"name": "chance",
		"age": 18,
	}

	plainText, _ := json.Marshal(params)
	//7b22616765223a31382c226e616d65223a226368616e6365227d
	//7b22616765223a31382c226e616d65223a226368616e6365227d
	fmt.Printf("plainText hex(%d): %s\n", len(plainText), hex.EncodeToString(plainText))
	chipText, _ := AESEncrypt(plainText, key)
	fmt.Printf("chipText hex(%d): %s\n", len(chipText), hex.EncodeToString(chipText))
	base := base58.Encode(chipText)

	fmt.Printf("base58 = %s\n", base)
	//1bf9d9ff6bbaf559e6b74479b0e770b480de59c5aa9ff044478834a0b1ec6986
	enbase, _ := base58.Decode(base)
	fmt.Printf("enbase hex(%d): %s\n", len(enbase), hex.EncodeToString(enbase))
	raw, _ :=AESDecrypt(enbase, key)
	fmt.Printf("raw = %s\n", string(raw))
}